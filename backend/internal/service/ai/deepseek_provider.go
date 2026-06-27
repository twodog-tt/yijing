package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/config"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/jsonutil"
)

type DeepSeekProvider struct {
	cfg    *config.Config
	mock   Provider
	client *http.Client
}

func NewDeepSeekProvider(cfg *config.Config, mock Provider) *DeepSeekProvider {
	timeout := time.Duration(cfg.DeepSeekTimeoutSeconds) * time.Second
	return &DeepSeekProvider{
		cfg:  cfg,
		mock: mock,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *DeepSeekProvider) Name() string {
	return model.AIProviderDeepSeek
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Stream         bool            `json:"stream"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Usage *chatUsage `json:"usage,omitempty"`
}

type chatUsage struct {
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens,omitempty"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens,omitempty"`
}

type deepSeekResult struct {
	Content               string
	PromptCacheHitTokens  int
	PromptCacheMissTokens int
}

func (p *DeepSeekProvider) GenerateFullInterpretation(ctx context.Context, input GenerateInput) (*GenerateOutput, error) {
	if err := validateInput(input); err != nil {
		return p.fallback(ctx, input, "invalid input: "+err.Error())
	}
	if strings.TrimSpace(p.cfg.DeepSeekAPIKey) == "" {
		return p.fallback(ctx, input, "api key missing")
	}

	start := time.Now()
	result, err := p.callAPI(ctx, input)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		log.Printf("[ai] divination_id=%d provider=deepseek model=%s duration_ms=%d success=false error=%q",
			input.DivinationID, p.cfg.DeepSeekModel, duration, truncateErr(err.Error()))
		return p.fallback(ctx, input, err.Error())
	}

	content, err := jsonutil.ExtractJSONObjectFromText(result.Content)
	if err != nil {
		log.Printf("[ai] divination_id=%d provider=deepseek model=%s duration_ms=%d success=false error=%q",
			input.DivinationID, p.cfg.DeepSeekModel, duration, "invalid json from model")
		return p.fallback(ctx, input, "invalid json from model")
	}

	content, err = jsonutil.EnsureRequiredFields(content)
	if err != nil {
		log.Printf("[ai] divination_id=%d provider=deepseek model=%s duration_ms=%d success=false error=%q",
			input.DivinationID, p.cfg.DeepSeekModel, duration, "incomplete json fields")
		return p.fallback(ctx, input, "incomplete json fields")
	}

	log.Printf("[ai] divination_id=%d provider=deepseek model=%s duration_ms=%d cache_hit_tokens=%d cache_miss_tokens=%d success=true",
		input.DivinationID, p.cfg.DeepSeekModel, duration, result.PromptCacheHitTokens, result.PromptCacheMissTokens)

	return &GenerateOutput{
		Provider:              model.AIProviderDeepSeek,
		Content:               content,
		RawResponse:           result.Content,
		ModelName:             p.cfg.DeepSeekModel,
		DurationMs:            duration,
		PromptCacheHitTokens:  result.PromptCacheHitTokens,
		PromptCacheMissTokens: result.PromptCacheMissTokens,
	}, nil
}

func (p *DeepSeekProvider) callAPI(ctx context.Context, input GenerateInput) (*deepSeekResult, error) {
	reqBody := chatRequest{
		Model: p.cfg.DeepSeekModel,
		Messages: []chatMessage{
			{Role: "system", Content: SystemPrompt()},
			{Role: "user", Content: BuildUserPrompt(input)},
		},
		Stream:         false,
		MaxTokens:      p.cfg.DeepSeekMaxOutputTokens,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(p.cfg.DeepSeekBaseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.DeepSeekAPIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("api error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("empty choices")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("empty content")
	}
	result := &deepSeekResult{Content: content}
	if parsed.Usage != nil {
		result.PromptCacheHitTokens = parsed.Usage.PromptCacheHitTokens
		result.PromptCacheMissTokens = parsed.Usage.PromptCacheMissTokens
	}
	return result, nil
}

func (p *DeepSeekProvider) fallback(ctx context.Context, input GenerateInput, reason string) (*GenerateOutput, error) {
	log.Printf("[ai] divination_id=%d provider=deepseek fallback=mock_fallback reason=%q question=%q",
		input.DivinationID, truncateErr(reason), truncateQuestion(input.Question, 50))

	out, err := p.mock.GenerateFullInterpretation(ctx, input)
	if err != nil {
		return nil, err
	}
	out.Provider = model.AIProviderMockFallback
	out.FallbackUsed = 1
	out.ErrorMessage = truncateLogMsg(reason)
	return out, nil
}

func truncateLogMsg(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	runes := []rune(s)
	if len(runes) > 500 {
		return string(runes[:500])
	}
	return s
}

func truncateErr(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}
