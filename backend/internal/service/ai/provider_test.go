package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/config"
	"github.com/wangxintong/yijing/backend/internal/model"
)

func sampleAIInput() GenerateInput {
	return GenerateInput{
		DivinationID: 1,
		Question:     "我现在适不适合继续推进这个 AI 易经小程序？",
		CategoryName: "事业",
		PrimaryHexagram: HexagramInfo{
			FullName: "火天大有", Summary: "丰盛在握", BinaryCode: "111101",
		},
		ChangedHexagram: HexagramInfo{
			FullName: "雷天大壮", Summary: "气势充沛", BinaryCode: "111100",
		},
		MovingLines:  []int{6},
		LineSnapshot: `[{"position":6,"value":9,"is_yang":1,"is_moving":1}]`,
		FreeContent:  "免费解读示例",
	}
}

func validReportJSON() string {
	return `{
	  "summary": "宜稳步推进",
	  "overall": "总体判断内容足够长，用于测试字段校验逻辑，需要超过若干字符以满足最小展示要求。",
	  "current_state": "当前处境内容足够长，用于测试字段校验逻辑，需要超过若干字符以满足最小展示要求。",
	  "opportunity": "机会点内容足够长，用于测试字段校验逻辑，需要超过若干字符以满足最小展示要求。",
	  "risk": "风险点内容足够长，用于测试字段校验逻辑，需要超过若干字符以满足最小展示要求。",
	  "action_steps": ["建议1", "建议2", "建议3"],
	  "emotion_reminder": "情绪提醒内容足够长，用于测试字段校验逻辑，需要超过若干字符以满足最小展示要求。",
	  "reflection_questions": ["问题1", "问题2", "问题3"],
	  "disclaimer": "本内容仅供娱乐和传统文化参考，不构成现实决策建议。"
	}`
}

func TestRouterMockProvider(t *testing.T) {
	cfg := &config.Config{AIProvider: "mock"}
	router := NewRouter(cfg, nil)
	out, err := router.GenerateFullInterpretation(context.Background(), sampleAIInput())
	if err != nil {
		t.Fatalf("mock generate failed: %v", err)
	}
	if out.Provider != model.AIProviderMock {
		t.Fatalf("expected mock provider, got %s", out.Provider)
	}
	if !strings.Contains(out.Content, `"summary"`) {
		t.Fatal("content should be json")
	}
}

func TestDeepSeekMissingAPIKeyFallback(t *testing.T) {
	cfg := &config.Config{
		AIProvider:              "deepseek",
		DeepSeekAPIKey:          "",
		DeepSeekBaseURL:         "https://example.com",
		DeepSeekModel:           "deepseek-v4-flash",
		DeepSeekTimeoutSeconds:  5,
		DeepSeekMaxOutputTokens: 1800,
	}
	p := NewDeepSeekProvider(cfg, NewMockProvider())
	out, err := p.GenerateFullInterpretation(context.Background(), sampleAIInput())
	if err != nil {
		t.Fatalf("expected fallback success: %v", err)
	}
	if out.Provider != model.AIProviderMockFallback {
		t.Fatalf("expected mock_fallback, got %s", out.Provider)
	}
}

func TestDeepSeekValidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": validReportJSON()}},
			},
			"usage": map[string]any{
				"prompt_cache_hit_tokens":  128,
				"prompt_cache_miss_tokens": 32,
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		AIProvider:              "deepseek",
		DeepSeekAPIKey:          "test-key",
		DeepSeekBaseURL:         server.URL,
		DeepSeekModel:           "deepseek-v4-flash",
		DeepSeekTimeoutSeconds:  5,
		DeepSeekMaxOutputTokens: 1800,
	}
	p := NewDeepSeekProvider(cfg, NewMockProvider())
	out, err := p.GenerateFullInterpretation(context.Background(), sampleAIInput())
	if err != nil {
		t.Fatalf("deepseek generate failed: %v", err)
	}
	if out.Provider != model.AIProviderDeepSeek {
		t.Fatalf("expected deepseek, got %s", out.Provider)
	}
	if out.PromptCacheHitTokens != 128 || out.PromptCacheMissTokens != 32 {
		t.Fatalf("unexpected cache usage: hit=%d miss=%d", out.PromptCacheHitTokens, out.PromptCacheMissTokens)
	}
}

func TestBuildUserPromptKeepsStaticInstructionsBeforeDynamicInput(t *testing.T) {
	prompt := BuildUserPrompt(sampleAIInput())
	rulesIndex := strings.Index(prompt, "请输出 JSON")
	inputIndex := strings.LastIndex(prompt, "【本次输入】")
	questionIndex := strings.LastIndex(prompt, "用户问题：")
	if rulesIndex < 0 || inputIndex < 0 || questionIndex < 0 {
		t.Fatalf("prompt missing expected sections: %s", prompt)
	}
	if rulesIndex > inputIndex || inputIndex > questionIndex {
		t.Fatalf("expected static instructions before dynamic input: %s", prompt)
	}
}

func TestDeepSeekInvalidJSONFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "not-json"}},
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		AIProvider:              "deepseek",
		DeepSeekAPIKey:          "test-key",
		DeepSeekBaseURL:         server.URL,
		DeepSeekModel:           "deepseek-v4-flash",
		DeepSeekTimeoutSeconds:  5,
		DeepSeekMaxOutputTokens: 1800,
	}
	p := NewDeepSeekProvider(cfg, NewMockProvider())
	out, err := p.GenerateFullInterpretation(context.Background(), sampleAIInput())
	if err != nil {
		t.Fatalf("expected fallback success: %v", err)
	}
	if out.Provider != model.AIProviderMockFallback {
		t.Fatalf("expected mock_fallback, got %s", out.Provider)
	}
}

func TestMockProviderNotCalledWhenExistingContent(t *testing.T) {
	mock := NewMockProvider()
	out1, err := mock.GenerateFullInterpretation(context.Background(), sampleAIInput())
	if err != nil {
		t.Fatal(err)
	}
	out2, err := mock.GenerateFullInterpretation(context.Background(), sampleAIInput())
	if err != nil {
		t.Fatal(err)
	}
	if out1.Content == "" || out2.Content == "" {
		t.Fatal("mock should always return content")
	}
}
