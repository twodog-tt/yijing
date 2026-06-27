package bazi

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

func sampleResultPayload(hour string) json.RawMessage {
	payload := map[string]any{
		"day_master": "甲",
		"pillars": map[string]string{
			"year":  "乙亥",
			"month": "己丑",
			"day":   "甲子",
		},
		"five_elements": map[string]int{
			"wood": 2, "fire": 1, "earth": 1, "metal": 1, "water": 1,
		},
		"reflection_focus":   "观察节奏",
		"action_suggestions": []string{"记录一周状态"},
		"method_note":        MethodNote,
		"calculation_meta": map[string]any{
			"limits": []string{"简化规则，不等同专业排盘"},
		},
	}
	if hour != "" {
		payload["pillars"].(map[string]string)["hour"] = hour
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return raw
}

func TestFullReportGeneratorUsesTemplateWhenDeepSeekDisabled(t *testing.T) {
	gen := NewFullReportGenerator(nil)
	content, provider, err := gen.Generate(context.Background(), 1, sampleResultPayload("甲子"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【一、简要说明】") {
		t.Fatalf("expected template sections, got %q", content)
	}
}

func TestFullReportGeneratorTemplateHourUnknownNote(t *testing.T) {
	gen := NewFullReportGenerator(nil)
	content, _, err := gen.Generate(context.Background(), 1, sampleResultPayload(""), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(content, "时辰未知，本次不生成时柱") {
		t.Fatalf("expected hour unknown note in template content")
	}
}

func TestBuildBaziUserPromptHourUnknown(t *testing.T) {
	input, err := buildFullReportPromptInput(sampleResultPayload(""), "免费解读")
	if err != nil {
		t.Fatalf("build prompt input: %v", err)
	}
	prompt := buildBaziUserPrompt(input)
	if !strings.Contains(prompt, "时辰未知，本次不生成时柱") {
		t.Fatalf("expected hour unknown instruction in prompt: %q", prompt)
	}
	for _, forbidden := range []string{"birth_date", "birth_hour_branch", "session_key"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt must not contain %s", forbidden)
		}
	}
}

func TestBuildBaziUserPromptV2UsesStructuredSummariesAndHidesSensitiveInput(t *testing.T) {
	v2, err := CalculateV2("1995-03-12", "mao", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	calc := CalculationResultFromV2(v2)
	payload, err := BuildV2APIResultPayload(v2, calc)
	if err != nil {
		t.Fatalf("BuildV2APIResultPayload: %v", err)
	}
	input, err := buildFullReportPromptInput(payload, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("build prompt input: %v", err)
	}
	prompt := buildBaziUserPrompt(input)

	for _, snippet := range []string{
		"algorithm_version：bazi-v2-poc",
		"calendar_basis：",
		"year_boundary=lichun",
		"month_boundary=solar_terms_jie",
		"true_solar_time=false",
		"pillars_v2_summary：",
		"year=" + v2.PillarsV2.Year,
		"month=" + v2.PillarsV2.Month,
		"day=" + v2.PillarsV2.Day,
		"five_elements_summary：",
		"wood=",
		"fire=",
		"earth=",
		"metal=",
		"water=",
		"bazi_profile：",
		"interpretation_lens：",
		"必须按以下 8 个部分输出",
	} {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("expected v2 prompt to contain %q\n%s", snippet, prompt)
		}
	}
	for _, forbidden := range []string{
		"1995-03-12",
		"session_key",
		"input_payload",
		"result_payload",
		"raw",
		"DeepSeek 原始",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("v2 prompt must not contain %q", forbidden)
		}
	}
}

func TestBuildBaziUserPromptV2HourUnknownDoesNotInventHourPillar(t *testing.T) {
	v2, err := CalculateV2("1988-05-18", "", true)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	calc := CalculationResultFromV2(v2)
	payload, err := BuildV2APIResultPayload(v2, calc)
	if err != nil {
		t.Fatalf("BuildV2APIResultPayload: %v", err)
	}
	input, err := buildFullReportPromptInput(payload, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("build prompt input: %v", err)
	}
	prompt := buildBaziUserPrompt(input)
	if !strings.Contains(prompt, "hour=时辰未知，本次不生成时柱") {
		t.Fatalf("expected unknown-hour v2 prompt")
	}
	if strings.Contains(prompt, "1988-05-18") {
		t.Fatalf("v2 prompt must not contain full birth date")
	}
}

func TestIsValidDeepSeekFullContentRejectsStrongPredictionWords(t *testing.T) {
	cases := []string{
		strings.Repeat("完整报告。", 40) + "未来必然发生变局。免责声明：仅供学习。",
		strings.Repeat("完整报告。", 40) + "结果一定会成功。免责声明：仅供学习。",
		strings.Repeat("完整报告。", 40) + "命运注定如此。免责声明：仅供学习。",
		strings.Repeat("完整报告。", 40) + "百分百准确。免责声明：仅供学习。",
		strings.Repeat("完整报告。", 40) + "可以预测未来走势。免责声明：仅供学习。",
		strings.Repeat("完整报告。", 40) + "保证结果无误。免责声明：仅供学习。",
	}
	for _, content := range cases {
		if isValidDeepSeekFullContent(content, false) {
			t.Fatalf("expected rejection for content containing forbidden prediction wording")
		}
	}
}

func TestIsValidDeepSeekFullContentRejectsForbiddenPhrase(t *testing.T) {
	content := strings.Repeat("完整报告。", 40) + "保证发财。免责声明：仅供学习。"
	if isValidDeepSeekFullContent(content, false) {
		t.Fatalf("expected forbidden phrase rejection")
	}
}

func TestIsValidDeepSeekFullContentRejectsEmptyLikeOutput(t *testing.T) {
	if isValidDeepSeekFullContent("免责声明", false) {
		t.Fatalf("expected short content rejection")
	}
}

func TestFullReportGeneratorUsesDeepSeekWhenConfigured(t *testing.T) {
	server := newDeepSeekTestServer(t, validDeepSeekReport(false))
	defer server.Close()

	cfg := &config.Config{
		AIProvider:              model.AIProviderDeepSeek,
		DeepSeekAPIKey:          "test-key",
		DeepSeekBaseURL:         server.URL,
		DeepSeekModel:           "deepseek-chat",
		DeepSeekMaxOutputTokens: 800,
		DeepSeekTimeoutSeconds:  5,
	}
	gen := NewFullReportGenerator(cfg)
	content, provider, err := gen.Generate(context.Background(), 9, sampleResultPayload("甲子"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderDeepSeek {
		t.Fatalf("expected deepseek provider, got %q", provider)
	}
	if !strings.Contains(content, "免责声明") {
		t.Fatalf("expected disclaimer in AI content")
	}
}

func TestFullReportGeneratorFallsBackWhenDeepSeekReturnsEmpty(t *testing.T) {
	server := newDeepSeekTestServer(t, `{"choices":[{"message":{"content":""}}]}`)
	defer server.Close()

	cfg := &config.Config{
		AIProvider:              model.AIProviderDeepSeek,
		DeepSeekAPIKey:          "test-key",
		DeepSeekBaseURL:         server.URL,
		DeepSeekModel:           "deepseek-chat",
		DeepSeekMaxOutputTokens: 800,
		DeepSeekTimeoutSeconds:  5,
	}
	gen := NewFullReportGenerator(cfg)
	content, provider, err := gen.Generate(context.Background(), 9, sampleResultPayload("甲子"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【一、简要说明】") {
		t.Fatalf("expected template fallback content")
	}
}

func TestFullReportGeneratorFallsBackWhenDeepSeekReturnsForbiddenPhrase(t *testing.T) {
	badContent := strings.Repeat("完整报告展开内容。", 20) + "未来一定会发生。免责声明：仅供学习参考。"
	raw, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{
			{"message": map[string]string{"content": badContent}},
		},
	})
	server := newDeepSeekTestServer(t, string(raw))
	defer server.Close()

	cfg := &config.Config{
		AIProvider:              model.AIProviderDeepSeek,
		DeepSeekAPIKey:          "test-key",
		DeepSeekBaseURL:         server.URL,
		DeepSeekModel:           "deepseek-chat",
		DeepSeekMaxOutputTokens: 800,
		DeepSeekTimeoutSeconds:  5,
	}
	gen := NewFullReportGenerator(cfg)
	content, provider, err := gen.Generate(context.Background(), 9, sampleResultPayload("甲子"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【一、简要说明】") {
		t.Fatalf("expected template fallback content")
	}
}

func validDeepSeekReport(hourUnknown bool) string {
	hourNote := ""
	if hourUnknown {
		hourNote = "时辰未知，本次不生成时柱，相关内容仅基于已知信息进行简化分析。"
	}
	content := strings.Repeat("完整报告展开内容。", 20) + hourNote + "\n7. 免责声明：基于简化干支文化规则，仅供传统文化学习与自我反思，不构成现实决策依据。"
	raw, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{
			{"message": map[string]string{"content": content}},
		},
	})
	return string(raw)
}

func newDeepSeekTestServer(t *testing.T, responseBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
}
