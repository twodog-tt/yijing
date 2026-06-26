package qimen

import (
	"strings"
	"testing"
	"time"
)

func TestExtractQuestionProfileAdvanceVsCommunication(t *testing.T) {
	advance := ExtractQuestionProfile("我最近适合推进这个项目吗？", "career")
	communication := ExtractQuestionProfile("我和合作伙伴沟通总是不顺，应该怎么调整？", "relationship")

	if advance.IntentType != "推进计划" {
		t.Fatalf("expected 推进计划, got %q", advance.IntentType)
	}
	if communication.IntentType != "关系沟通" {
		t.Fatalf("expected 关系沟通, got %q", communication.IntentType)
	}
	if advance.IntentType == communication.IntentType {
		t.Fatalf("intent types must differ")
	}
}

func TestExtractQuestionProfileStudyAndDecision(t *testing.T) {
	study := ExtractQuestionProfile("我最近学习状态不好，怎么安排节奏？", "study")
	decision := ExtractQuestionProfile("我是否应该现在换一个方向发展？", "decision")

	if study.IntentType != "学习节奏" {
		t.Fatalf("expected 学习节奏, got %q", study.IntentType)
	}
	if decision.IntentType != "决策选择" {
		t.Fatalf("expected 决策选择, got %q", decision.IntentType)
	}
	if study.TimeHorizon != "短期" {
		t.Fatalf("expected short horizon, got %q", study.TimeHorizon)
	}
}

func TestBuildQimenLensDiffersByProfile(t *testing.T) {
	advanceProfile := ExtractQuestionProfile("我最近适合推进这个项目吗？", "career")
	commProfile := ExtractQuestionProfile("我和合作伙伴沟通总是不顺，应该怎么调整？", "relationship")

	advanceLens := BuildQimenLens(advanceProfile, "career")
	commLens := BuildQimenLens(commProfile, "relationship")

	if advanceLens.FocusTheme == commLens.FocusTheme {
		t.Fatalf("focus themes should differ")
	}
	if advanceLens.PacingTheme == commLens.PacingTheme && advanceProfile.IntentType == commProfile.IntentType {
		t.Fatalf("pacing themes should differ for different intents")
	}
}

func TestCalculateDifferentCategoriesProduceDifferentPayloads(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	cases := []struct {
		question string
		category string
	}{
		{"我最近适合推进这个项目吗？", "career"},
		{"我和合作伙伴沟通总是不顺，应该怎么调整？", "relationship"},
		{"我最近学习状态不好，怎么安排节奏？", "study"},
	}

	overviews := make(map[string]string)
	freeContents := make(map[string]string)
	for _, c := range cases {
		calc, err := Calculate(c.question, c.category, fixed)
		if err != nil {
			t.Fatalf("Calculate: %v", err)
		}
		if calc.QuestionProfile.IntentType == "" || calc.QimenLens.FocusTheme == "" {
			t.Fatalf("expected profile/lens populated")
		}
		overviews[c.category] = calc.SituationOverview
		freeContents[c.category] = BuildFreeContent(calc)
	}

	if len(overviews) != len(cases) {
		t.Fatalf("expected unique category outputs")
	}
	for a, b := range overviews {
		for c, d := range overviews {
			if a == c {
				continue
			}
			if b == d {
				t.Fatalf("overview for %s and %s must differ", a, c)
			}
		}
	}
}

func TestSameCategoryDifferentQuestionsDiffer(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	c1, err := Calculate("我最近适合推进这个项目吗？", "career", fixed)
	if err != nil {
		t.Fatalf("Calculate c1: %v", err)
	}
	c2, err := Calculate("团队资源协调总是卡住，下一步怎么安排？", "career", fixed)
	if err != nil {
		t.Fatalf("Calculate c2: %v", err)
	}

	if c1.SituationOverview == c2.SituationOverview {
		t.Fatalf("same category questions should produce different overviews")
	}
	if c1.ActionPacing == c2.ActionPacing {
		t.Fatalf("same category questions should produce different pacing")
	}
	if BuildFreeContent(c1) == BuildFreeContent(c2) {
		t.Fatalf("free_content should differ for different questions")
	}
}

func TestDecisionCategorySuggestsInformationGathering(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	calc, err := Calculate("我是否应该现在换一个方向发展？", "decision", fixed)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	combined := strings.Join(calc.ActionSuggestions, " ") + calc.ActionPacing
	if !strings.Contains(combined, "选项") && !strings.Contains(combined, "试探") {
		t.Fatalf("decision suggestions should mention options or trial steps, got %q", combined)
	}
}

func TestStudyCategorySuggestsLearningActions(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	calc, err := Calculate("我最近学习状态不好，怎么安排节奏？", "study", fixed)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	combined := strings.Join(calc.ActionSuggestions, " ")
	if !strings.Contains(combined, "学习") {
		t.Fatalf("study suggestions should mention learning, got %q", combined)
	}
}

func TestBuildFreeContentIncludesLensSection(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	calc, err := Calculate("我最近做决定容易犹豫，应该先看哪些风险？", "general", fixed)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	free := BuildFreeContent(calc)
	for _, section := range []string{"【关注主题】", "【局势梳理】", "【风险观察】", "【行动节奏】"} {
		if !strings.Contains(free, section) {
			t.Fatalf("free_content missing %s", section)
		}
	}
}

func TestBuildQimenUserPromptIncludesProfileAndLens(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	calc, err := Calculate("我最近适合推进这个项目吗？", "career", fixed)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	raw, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	input, err := buildFullReportPromptInput(raw, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	prompt := buildQimenUserPrompt(input)
	for _, required := range []string{
		"question_profile",
		"qimen_lens",
		"问事特征",
		"intent_type=推进计划",
		"focus_theme=",
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("prompt missing %q", required)
		}
	}
}

func TestBuildQimenUserPromptExcludesSensitiveFields(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	calc, err := Calculate("我最近适合推进这个项目吗？", "career", fixed)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	raw, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	prompt := buildQimenUserPrompt(mustPromptInput(t, raw, BuildFreeContent(calc)))
	for _, forbidden := range []string{
		"session_key",
		"input_payload",
		"result_payload",
		`{"question"`,
		"differentiation_seed",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt must not contain %s", forbidden)
		}
	}
}

func mustPromptInput(t *testing.T, raw []byte, free string) *fullReportPromptInput {
	t.Helper()
	input, err := buildFullReportPromptInput(raw, free)
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	return input
}

func TestResultPayloadIncludesProfileAndLens(t *testing.T) {
	fixed := time.Date(2025, 6, 23, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))
	calc, err := Calculate("我最近适合推进这个项目吗？", "career", fixed)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	raw, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	body := string(raw)
	for _, key := range []string{
		"question_profile",
		"qimen_lens",
		"differentiation_seed",
		"safe_question_summary",
	} {
		if !strings.Contains(body, key) {
			t.Fatalf("result_payload missing %s", key)
		}
	}
	if strings.Contains(body, calc.Question) {
		t.Fatalf("result_payload must not contain raw question")
	}
}
