package qimen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/config"
)

const (
	qimenSystemPrompt = `你是“传统文化奇门问事助手”。

你的任务是根据 qimen-simple-v1 简化奇门文化规则下的结构化信息，生成一份温和、理性、适合普通用户阅读的完整报告。

重要边界：
1. 这是基于 qimen-simple-v1 简化学习版，不等同于专业奇门排盘。
2. 不生成完整九宫盘，不做专业排盘。
3. 不构成现实决策依据。
4. 内容仅用于传统文化学习与自我反思、局势梳理和行动节奏参考。
5. 不要声称可以精准预测未来或给出必成必败结论。
6. 不要使用“必然、一定、注定、百分百、保证、大凶、大吉”等绝对词。
7. 禁止生成：精准预测、必成必败、大凶大吉、必发财、必复合、改运、化灾、投资建议、医疗建议、法律建议、赌博建议、军事行动建议。
8. 不要恐吓用户，不要诱导付费改运。
9. 语气温和、克制、不玄乎。
10. 必须根据 question_profile 与 qimen_lens 写出差异化报告，不要套用固定模板；每个问题都要围绕 intent_type、risk_tone、pacing_theme 展开。

你必须只输出纯文本报告正文，不要输出 Markdown 代码块，不要输出 JSON，不要输出额外解释。`

	qimenUserPromptTemplate = `请基于以下简化奇门问事分析信息生成完整报告。

要求按以下 8 个部分输出，每部分用标题开头：
1. 方法说明与简化边界
2. 局势梳理展开
3. 风险观察展开
4. 行动节奏与节奏建议
5. 自我反思问题
6. 行动建议
7. 观察与延伸
8. 免责声明

方法说明：{{method_note}}
问事分类：{{category_label}}
时段参考：{{time_bucket_label}}
问事摘要：{{question_summary}}
安全问事特征：{{safe_question_summary}}
question_profile：{{question_profile}}
qimen_lens：{{qimen_lens}}
局势梳理：{{situation_overview}}
风险观察：{{risk_observations}}
行动节奏：{{action_pacing}}
自我反思问题：{{reflection_questions}}
行动建议参考：{{action_suggestions}}
规则限制：{{limits}}
免费解读摘要：{{free_content}}

写作要求：
- 必须围绕 question_profile 的 intent_type、risk_tone 与 qimen_lens 的 focus_theme、pacing_theme 展开，避免不同问题写出相同段落。
- 不要套用固定模板句，各章节内容需体现本次问事的差异。
- 保持传统文化学习与自我反思语气，不做精准预测与强吉凶判断。

必须再次强调：这是 qimen-simple-v1 简化学习版，不生成完整九宫盘，仅供传统文化学习与自我反思，不构成现实决策依据。`
)

var qimenForbiddenPhrases = []string{
	"精准预测", "必成必败", "大凶", "大吉", "必发财", "必复合",
	"改运", "化灾", "投资建议", "医疗建议", "法律建议", "赌博建议", "军事行动",
	"预测未来", "保证结果", "百分百", "注定", "必然", "一定会", "一定得",
	"精准算命", "婚姻财运预测", "保证发财", "看广告改运",
}

type deepSeekFullGenerator struct {
	cfg    *config.Config
	client *http.Client
}

func newDeepSeekFullGenerator(cfg *config.Config) *deepSeekFullGenerator {
	timeout := 60 * time.Second
	if cfg != nil && cfg.DeepSeekTimeoutSeconds > 0 {
		timeout = time.Duration(cfg.DeepSeekTimeoutSeconds) * time.Second
	}
	return &deepSeekFullGenerator{
		cfg: cfg,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (g *deepSeekFullGenerator) enabled() bool {
	if g == nil || g.cfg == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(g.cfg.AIProvider), "deepseek") &&
		strings.TrimSpace(g.cfg.DeepSeekAPIKey) != ""
}

func (g *deepSeekFullGenerator) generate(ctx context.Context, _ int64, input *fullReportPromptInput) (string, error) {
	if !g.enabled() {
		return "", fmt.Errorf("deepseek not configured")
	}
	if input == nil {
		return "", fmt.Errorf("invalid input")
	}

	userPrompt := buildQimenUserPrompt(input)
	content, err := g.callAPI(ctx, userPrompt)
	if err != nil {
		return "", err
	}
	if !isValidDeepSeekFullContent(content) {
		return "", fmt.Errorf("invalid model output")
	}
	return content, nil
}

func buildQimenUserPrompt(input *fullReportPromptInput) string {
	timeBucketLabel := nonEmpty(input.TimeBucketLabel, "（未指定）")
	risks := "（无）"
	if len(input.RiskObservations) > 0 {
		risks = strings.Join(input.RiskObservations, "；")
	}
	reflections := "（无）"
	if len(input.ReflectionQuestions) > 0 {
		reflections = strings.Join(input.ReflectionQuestions, "；")
	}
	suggestions := "（无）"
	if len(input.ActionSuggestions) > 0 {
		suggestions = strings.Join(input.ActionSuggestions, "；")
	}
	limits := "（无）"
	if len(input.Limits) > 0 {
		limits = strings.Join(input.Limits, "；")
	}

	replacer := strings.NewReplacer(
		"{{method_note}}", input.MethodNote,
		"{{category_label}}", input.CategoryLabel,
		"{{time_bucket_label}}", timeBucketLabel,
		"{{question_summary}}", input.QuestionSummary,
		"{{safe_question_summary}}", nonEmpty(input.SafeQuestionSummary, "（无）"),
		"{{question_profile}}", formatQuestionProfileForPrompt(input.QuestionProfile),
		"{{qimen_lens}}", formatQimenLensForPrompt(input.QimenLens),
		"{{situation_overview}}", input.SituationOverview,
		"{{risk_observations}}", risks,
		"{{action_pacing}}", nonEmpty(input.ActionPacing, "（无）"),
		"{{reflection_questions}}", reflections,
		"{{action_suggestions}}", suggestions,
		"{{limits}}", limits,
		"{{free_content}}", nonEmpty(input.FreeContent, "（无）"),
	)
	return replacer.Replace(qimenUserPromptTemplate)
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	Stream    bool          `json:"stream"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
}

func (g *deepSeekFullGenerator) callAPI(ctx context.Context, userPrompt string) (string, error) {
	reqBody := chatRequest{
		Model: g.cfg.DeepSeekModel,
		Messages: []chatMessage{
			{Role: "system", Content: qimenSystemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream:    false,
		MaxTokens: g.cfg.DeepSeekMaxOutputTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(g.cfg.DeepSeekBaseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.cfg.DeepSeekAPIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("parse response")
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return "", fmt.Errorf("api error")
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty content")
	}
	return content, nil
}

func isValidDeepSeekFullContent(content string) bool {
	content = strings.TrimSpace(content)
	if len([]rune(content)) < 120 {
		return false
	}
	for _, phrase := range qimenForbiddenPhrases {
		if strings.Contains(content, phrase) {
			return false
		}
	}
	if !strings.Contains(content, "免责声明") {
		return false
	}
	return true
}

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
