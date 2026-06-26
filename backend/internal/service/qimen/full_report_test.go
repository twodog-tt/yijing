package qimen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/config"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func sampleResultPayload(category string) json.RawMessage {
	if category == "" {
		category = "career"
	}
	profile := ExtractQuestionProfile("示例问事问题用于测试", category)
	lens := BuildQimenLens(profile, category)
	payload := map[string]any{
		"algorithm_version":  model.AlgorithmVersionQimenSimpleV1,
		"method_note":        MethodNote,
		"question_summary":   QuestionSummary,
		"safe_question_summary": BuildSafeQuestionSummary(profile),
		"category":           category,
		"time_context":       map[string]string{"time_bucket": "day"},
		"question_profile":   profile,
		"qimen_lens":         lens,
		"differentiation_seed": BuildDifferentiationSeed(category, "day"),
		"situation_overview": "当前局势更像是在整理方向与节奏，适合先观察再推进。",
		"risk_observations":  []string{"过度依赖单一结论，可能忽略现实细节。"},
		"action_pacing":      "建议分三步：先整理现状，再安排小动作，最后复盘。",
		"reflection_questions": []string{
			"我真正想推进的核心目标是什么？",
		},
		"action_suggestions": []string{"用一页纸写下现状与目标。"},
		"calculation_meta": map[string]any{
			"limits": calculationLimits,
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return raw
}

func TestFullReportGeneratorUsesTemplateWhenDeepSeekDisabled(t *testing.T) {
	gen := NewFullReportGenerator(nil)
	content, provider, err := gen.Generate(context.Background(), 1, sampleResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【一、问题局势摘要】") {
		t.Fatalf("expected template sections, got %q", content)
	}
}

func TestBuildFullContentRejectsMalformedPayload(t *testing.T) {
	_, err := BuildFullContent(json.RawMessage(`{"category":"career"}`), "")
	if err == nil {
		t.Fatalf("expected error for malformed payload")
	}
}

func TestBuildFullReportPromptInputRejectsMalformedPayload(t *testing.T) {
	_, err := buildFullReportPromptInput(json.RawMessage(`not-json`), "")
	if err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestBuildQimenUserPromptPrivacy(t *testing.T) {
	raw := sampleResultPayload("career")
	input, err := buildFullReportPromptInput(raw, "免费解读")
	if err != nil {
		t.Fatalf("build prompt input: %v", err)
	}
	prompt := buildQimenUserPrompt(input)
	if !strings.Contains(prompt, QuestionSummary) {
		t.Fatalf("expected sanitized question summary in prompt")
	}
	if !strings.Contains(prompt, "question_profile") {
		t.Fatalf("expected question_profile in prompt")
	}
	if !strings.Contains(prompt, "qimen_lens") {
		t.Fatalf("expected qimen_lens in prompt")
	}
	for _, forbidden := range []string{
		"session_key",
		"input_payload",
		"result_payload",
		`{"question"`,
		"我最近适合推进这个计划吗",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt must not contain %s", forbidden)
		}
	}
}

func TestBuildFullReportPromptInputForcesSafeQuestionSummary(t *testing.T) {
	payload := sampleResultPayload("career")
	var obj map[string]any
	if err := json.Unmarshal(payload, &obj); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	obj["question_summary"] = "用户原问题全文不应进入 Prompt"
	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	input, err := buildFullReportPromptInput(raw, "免费解读")
	if err != nil {
		t.Fatalf("build prompt input: %v", err)
	}
	if input.QuestionSummary != QuestionSummary {
		t.Fatalf("expected forced question summary, got %q", input.QuestionSummary)
	}
	prompt := buildQimenUserPrompt(input)
	if strings.Contains(prompt, "用户原问题全文不应进入 Prompt") {
		t.Fatalf("prompt must not contain tampered question_summary")
	}
}

func TestSummarizeFreeContentForPromptTruncatesLongText(t *testing.T) {
	long := strings.Repeat("解读内容。", 100)
	summary := summarizeFreeContentForPrompt(long)
	if len([]rune(summary)) > maxFreeContentPromptRunes+20 {
		t.Fatalf("expected truncated summary, got length %d", len([]rune(summary)))
	}
	if !strings.Contains(summary, "已截断") {
		t.Fatalf("expected truncation marker")
	}
}

func TestIsValidDeepSeekFullContentRejectsForbiddenPhrase(t *testing.T) {
	content := strings.Repeat("完整报告。", 40) + "必成必败。免责声明：仅供学习。"
	if isValidDeepSeekFullContent(content) {
		t.Fatalf("expected forbidden phrase rejection")
	}
}

func TestIsValidDeepSeekFullContentRejectsEmptyLikeOutput(t *testing.T) {
	if isValidDeepSeekFullContent("免责声明") {
		t.Fatalf("expected short content rejection")
	}
}

func TestFullReportGeneratorUsesDeepSeekWhenConfigured(t *testing.T) {
	server := newDeepSeekTestServer(t, validDeepSeekReport())
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
	content, provider, err := gen.Generate(context.Background(), 9, sampleResultPayload("career"), "免费解读")
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
	content, provider, err := gen.Generate(context.Background(), 9, sampleResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【一、问题局势摘要】") {
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
	content, provider, err := gen.Generate(context.Background(), 9, sampleResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【一、问题局势摘要】") {
		t.Fatalf("expected template fallback content")
	}
}

func TestFullReportGeneratorFallsBackWhenDeepSeekFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
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
	_, provider, err := gen.Generate(context.Background(), 9, sampleResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
}

func TestBuildFullContentDoesNotContainForbiddenWords(t *testing.T) {
	content, err := BuildFullContent(sampleResultPayload("general"), "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	body := reportBodyExcludingBoundary(content)
	for _, phrase := range []string{"精准预测", "必成", "必败", "改运", "投资建议"} {
		if strings.Contains(body, phrase) {
			t.Fatalf("template content body must not contain %q", phrase)
		}
	}
}

func validDeepSeekReport() string {
	content := strings.Repeat("完整报告展开内容。", 20) +
		"\n8. 免责声明：基于 qimen-simple-v1 简化规则，仅供传统文化学习与自我反思，不构成现实决策依据。"
	raw, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{
			{"message": map[string]string{"content": content}},
		},
	})
	return string(raw)
}

func sampleV2ResultPayload(category string) json.RawMessage {
	raw := sampleResultPayload(category)
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		panic(err)
	}
	palaces := make([]Palace, 9)
	for i := range palaces {
		palaces[i] = Palace{
			Index:           i + 1,
			Name:            palaceNames[i],
			EarthPlateStem:  "戊",
			HeavenPlateStem: "甲",
			Star:            palaceStars[i],
			Door:            palaceDoors[i],
			Deity:           palaceDeities[i],
			Summary:         palaceSummaryPOC,
		}
	}
	obj["algorithm_version"] = AlgorithmVersionQimenV2POC
	obj["method_note"] = MethodNoteV2
	obj["calendar_basis"] = CalendarBasis{
		SolarTerm:  "惊蛰",
		JieqiBasis: "formula_approximation",
		TimeBasis:  "local_time",
		Note:       calendarNotePOC,
	}
	obj["dun"] = Dun{Type: "yang", Ju: 2, Source: "poc_formula"}
	obj["xun"] = Xun{XunShou: "甲子", EmptyBranches: []string{"戌", "亥"}}
	obj["chief"] = Chief{ZhiFu: "天冲", ZhiShi: "生门"}
	obj["palaces"] = palaces
	obj["limits"] = calculationLimitsV2
	obj["calculation_meta"] = map[string]any{"limits": calculationLimitsV2}
	out, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return out
}

func TestBuildFullContentSupportsV2Payload(t *testing.T) {
	content, err := BuildFullContent(sampleV2ResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent v2: %v", err)
	}
	if !strings.Contains(content, fullReportDisclaimerV2) {
		t.Fatalf("expected v2 disclaimer in full content")
	}
	if !strings.Contains(content, v2SectionPalaces) {
		t.Fatalf("expected v2 palace section")
	}
	if !strings.Contains(content, "POC") {
		t.Fatalf("expected POC note")
	}
}

func TestBuildFullReportPromptInputIncludesV2Fields(t *testing.T) {
	input, err := buildFullReportPromptInput(sampleV2ResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("buildFullReportPromptInput v2: %v", err)
	}
	if input.AlgorithmVersion != AlgorithmVersionQimenV2POC {
		t.Fatalf("algorithm_version=%q", input.AlgorithmVersion)
	}
	prompt := buildQimenUserPrompt(input)
	for _, want := range []string{"calendar_basis", "palaces_summary", "focus_palaces_summary", "qimen-v2-poc"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}

func TestBuildFullReportPromptInputIncludesProfessionalFields(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	v1, err := Calculate("test question here", "career", when)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	pro, err := CalculateProfessionalPreview(CalculateInputProfessional{Category: "career", Now: when})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	raw, err := BuildProfessionalAPIResultPayload(v1, pro)
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	input, err := buildFullReportPromptInput(raw, "免费解读")
	if err != nil {
		t.Fatalf("buildFullReportPromptInput professional: %v", err)
	}
	if input.AlgorithmVersion != AlgorithmVersionQimenV2Professional {
		t.Fatalf("algorithm_version=%q", input.AlgorithmVersion)
	}
	prompt := buildQimenUserPrompt(input)
	for _, want := range []string{"layout_version", "ganzhi", "palaces_summary", "focus_palaces_summary", "qimen-v2-professional"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
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
