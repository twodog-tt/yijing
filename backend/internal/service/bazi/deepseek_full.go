package bazi

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
	baziSystemPrompt = `你是“传统文化八字简析助手”。

你的任务是根据简化干支文化规则下的结构化信息，生成一份温和、理性、适合普通用户阅读的完整报告。

重要边界：
1. 这是基于 bazi-simple-v1 简化干支文化规则的学习参考。
2. 不等同于专业八字排盘。
3. 不构成现实决策依据。
4. 内容仅用于传统文化学习与自我反思。
5. 不要声称可以预测未来或一生命运。
6. 不要使用“必然、一定、注定、百分百、保证”等绝对词。
7. 禁止生成：精准算命、婚姻财运预测、保证发财、保证复合、疾病寿命、改运化解、投资建议、医疗建议、法律建议、赌博建议。
8. 不要恐吓用户，不要诱导付费改运。
9. 语气温和、克制、不玄乎。

你必须只输出纯文本报告正文，不要输出 Markdown 代码块，不要输出 JSON，不要输出额外解释。`

	baziUserPromptTemplate = `请基于以下简化八字分析信息生成完整报告。

要求按以下 7 个部分输出，每部分用标题开头：
1. 简化干支示意
2. 五行倾向展开
3. 日主与行动风格
4. 自我反思问题
5. 行动建议
6. 观察方向
7. 免责声明

方法说明：{{method_note}}
{{hour_unknown_note}}
简化干支示意：
- 年柱：{{year_pillar}}
- 月柱：{{month_pillar}}
- 日柱：{{day_pillar}}
{{hour_pillar_line}}
日主：{{day_master}}
五行倾向（简化计数）：木 {{wood}}、火 {{fire}}、土 {{earth}}、金 {{metal}}、水 {{water}}
反思焦点：{{reflection_focus}}
行动建议参考：{{action_suggestions}}
规则限制：{{limits}}
免费解读摘要：{{free_content}}

若时辰未知，必须在相关部分明确写出：时辰未知，本次不生成时柱，相关内容仅基于已知信息进行简化分析。
免责声明必须再次强调：基于简化干支文化规则，仅供传统文化学习与自我反思，不构成现实决策依据。`
)

var baziForbiddenPhrases = []string{
	"精准算命", "一生命运", "婚姻财运预测", "保证发财", "保证复合",
	"疾病寿命", "改运化解", "看广告改运", "投资建议", "医疗建议", "法律建议", "赌博建议",
	"预测未来", "保证结果", "百分百", "注定", "必然", "一定会", "一定得",
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

	userPrompt := buildBaziUserPrompt(input)
	content, err := g.callAPI(ctx, userPrompt)
	if err != nil {
		return "", err
	}
	if !isValidDeepSeekFullContent(content, input.HourUnknown) {
		return "", fmt.Errorf("invalid model output")
	}
	return content, nil
}

func buildBaziUserPrompt(input *fullReportPromptInput) string {
	hourUnknownNote := ""
	hourPillarLine := fmt.Sprintf("- 时柱：%s", nonEmpty(input.Pillars.Hour, "—"))
	if input.HourUnknown {
		hourUnknownNote = "时辰未知：本次不生成时柱，相关内容仅基于已知信息进行简化分析。"
		hourPillarLine = "- 时柱：时辰未知，本次不生成时柱"
	}

	suggestions := "（无）"
	if len(input.ActionSuggestions) > 0 {
		suggestions = strings.Join(input.ActionSuggestions, "；")
	}
	limits := "（无）"
	if len(input.Limits) > 0 {
		limits = strings.Join(input.Limits, "；")
	}
	freeContent := nonEmpty(input.FreeContent, "（无）")

	replacer := strings.NewReplacer(
		"{{method_note}}", input.MethodNote,
		"{{hour_unknown_note}}", hourUnknownNote,
		"{{year_pillar}}", nonEmpty(input.Pillars.Year, "—"),
		"{{month_pillar}}", nonEmpty(input.Pillars.Month, "—"),
		"{{day_pillar}}", nonEmpty(input.Pillars.Day, "—"),
		"{{hour_pillar_line}}", hourPillarLine,
		"{{day_master}}", input.DayMaster,
		"{{wood}}", fmt.Sprintf("%d", input.FiveElements.Wood),
		"{{fire}}", fmt.Sprintf("%d", input.FiveElements.Fire),
		"{{earth}}", fmt.Sprintf("%d", input.FiveElements.Earth),
		"{{metal}}", fmt.Sprintf("%d", input.FiveElements.Metal),
		"{{water}}", fmt.Sprintf("%d", input.FiveElements.Water),
		"{{reflection_focus}}", nonEmpty(input.ReflectionFocus, "（无）"),
		"{{action_suggestions}}", suggestions,
		"{{limits}}", limits,
		"{{free_content}}", freeContent,
	)
	return replacer.Replace(baziUserPromptTemplate)
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
			{Role: "system", Content: baziSystemPrompt},
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

func isValidDeepSeekFullContent(content string, hourUnknown bool) bool {
	content = strings.TrimSpace(content)
	if len([]rune(content)) < 120 {
		return false
	}
	for _, phrase := range baziForbiddenPhrases {
		if strings.Contains(content, phrase) {
			return false
		}
	}
	if !strings.Contains(content, "免责声明") {
		return false
	}
	if hourUnknown && !strings.Contains(content, "时辰未知") {
		return false
	}
	return true
}
