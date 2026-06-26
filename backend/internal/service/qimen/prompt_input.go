package qimen

import (
	"encoding/json"
	"fmt"
	"strings"
)

type timeContextPayload struct {
	TimeBucket string `json:"time_bucket"`
}

type calculationMetaPayload struct {
	Limits []string `json:"limits"`
}

type profilePayload struct {
	IntentType       string `json:"intent_type"`
	TimeHorizon      string `json:"time_horizon"`
	DecisionPressure string `json:"decision_pressure"`
	RelationScope    string `json:"relation_scope"`
	RiskTone         string `json:"risk_tone"`
}

type lensPayload struct {
	FocusTheme   string `json:"focus_theme"`
	SupportTheme string `json:"support_theme"`
	CautionTheme string `json:"caution_theme"`
	PacingTheme  string `json:"pacing_theme"`
}

type fullReportPromptInput struct {
	MethodNote          string
	QuestionSummary     string
	SafeQuestionSummary string
	Category            string
	CategoryLabel       string
	TimeBucket          string
	TimeBucketLabel     string
	QuestionProfile     QuestionProfile
	QimenLens           QimenLens
	SituationOverview   string
	RiskObservations    []string
	ActionPacing        string
	ReflectionQuestions []string
	ActionSuggestions   []string
	Limits              []string
	FreeContent         string
}

func buildFullReportPromptInput(resultPayload json.RawMessage, freeContent string) (*fullReportPromptInput, error) {
	var parsed struct {
		MethodNote          string                  `json:"method_note"`
		QuestionSummary     string                  `json:"question_summary"`
		SafeQuestionSummary string                  `json:"safe_question_summary"`
		Category            string                  `json:"category"`
		TimeContext         *timeContextPayload     `json:"time_context"`
		QuestionProfile     *profilePayload         `json:"question_profile"`
		QimenLens           *lensPayload            `json:"qimen_lens"`
		SituationOverview   string                  `json:"situation_overview"`
		RiskObservations    []string                `json:"risk_observations"`
		ActionPacing        string                  `json:"action_pacing"`
		ReflectionQuestions []string                `json:"reflection_questions"`
		ActionSuggestions   []string                `json:"action_suggestions"`
		CalculationMeta     *calculationMetaPayload `json:"calculation_meta"`
	}
	if err := json.Unmarshal(resultPayload, &parsed); err != nil {
		return nil, fmt.Errorf("invalid result_payload")
	}
	if strings.TrimSpace(parsed.SituationOverview) == "" {
		return nil, fmt.Errorf("invalid result_payload")
	}

	category := NormalizeCategory(parsed.Category)
	bucket := ""
	if parsed.TimeContext != nil {
		bucket = strings.TrimSpace(parsed.TimeContext.TimeBucket)
	}

	profile := profileFromPayload(parsed.QuestionProfile, category)
	lens := lensFromPayload(parsed.QimenLens, profile, category)

	summary := QuestionSummary
	safeSummary := strings.TrimSpace(parsed.SafeQuestionSummary)
	if safeSummary == "" {
		safeSummary = BuildSafeQuestionSummary(profile)
	}

	input := &fullReportPromptInput{
		MethodNote:          strings.TrimSpace(parsed.MethodNote),
		QuestionSummary:     summary,
		SafeQuestionSummary: safeSummary,
		Category:            category,
		CategoryLabel:       categoryLabel(category),
		TimeBucket:          bucket,
		TimeBucketLabel:     timeBucketLabel(bucket),
		QuestionProfile:     profile,
		QimenLens:           lens,
		SituationOverview:   strings.TrimSpace(parsed.SituationOverview),
		RiskObservations:    append([]string{}, parsed.RiskObservations...),
		ActionPacing:        strings.TrimSpace(parsed.ActionPacing),
		ReflectionQuestions: append([]string{}, parsed.ReflectionQuestions...),
		ActionSuggestions:   append([]string{}, parsed.ActionSuggestions...),
		FreeContent:         summarizeFreeContentForPrompt(freeContent),
	}
	if parsed.CalculationMeta != nil {
		input.Limits = append([]string{}, parsed.CalculationMeta.Limits...)
	}
	if input.MethodNote == "" {
		input.MethodNote = MethodNote
	}
	if len(input.Limits) == 0 {
		input.Limits = append([]string{}, calculationLimits...)
	}
	return input, nil
}

func profileFromPayload(p *profilePayload, category string) QuestionProfile {
	if p == nil {
		return QuestionProfile{
			IntentType:       defaultIntentForCategory(category),
			TimeHorizon:      "未明确",
			DecisionPressure: "中",
			RelationScope:    defaultRelationScope(category),
			RiskTone:         "平衡",
		}
	}
	profile := QuestionProfile{
		IntentType:       strings.TrimSpace(p.IntentType),
		TimeHorizon:      strings.TrimSpace(p.TimeHorizon),
		DecisionPressure: strings.TrimSpace(p.DecisionPressure),
		RelationScope:    strings.TrimSpace(p.RelationScope),
		RiskTone:         strings.TrimSpace(p.RiskTone),
	}
	if profile.IntentType == "" {
		profile.IntentType = defaultIntentForCategory(category)
	}
	if profile.TimeHorizon == "" {
		profile.TimeHorizon = "未明确"
	}
	if profile.DecisionPressure == "" {
		profile.DecisionPressure = "中"
	}
	if profile.RelationScope == "" {
		profile.RelationScope = defaultRelationScope(category)
	}
	if profile.RiskTone == "" {
		profile.RiskTone = "平衡"
	}
	return profile
}

func lensFromPayload(p *lensPayload, profile QuestionProfile, category string) QimenLens {
	if p == nil {
		return BuildQimenLens(profile, category)
	}
	lens := QimenLens{
		FocusTheme:   strings.TrimSpace(p.FocusTheme),
		SupportTheme: strings.TrimSpace(p.SupportTheme),
		CautionTheme: strings.TrimSpace(p.CautionTheme),
		PacingTheme:  strings.TrimSpace(p.PacingTheme),
	}
	if lens.FocusTheme == "" {
		return BuildQimenLens(profile, category)
	}
	if lens.SupportTheme == "" {
		lens.SupportTheme = supportThemeFor(profile, category)
	}
	if lens.CautionTheme == "" {
		lens.CautionTheme = cautionThemeFor(profile, category)
	}
	if lens.PacingTheme == "" {
		lens.PacingTheme = pacingThemeFor(profile, category)
	}
	return lens
}

func formatQuestionProfileForPrompt(profile QuestionProfile) string {
	return strings.Join([]string{
		"intent_type=" + profile.IntentType,
		"time_horizon=" + profile.TimeHorizon,
		"decision_pressure=" + profile.DecisionPressure,
		"relation_scope=" + profile.RelationScope,
		"risk_tone=" + profile.RiskTone,
	}, "；")
}

func formatQimenLensForPrompt(lens QimenLens) string {
	return strings.Join([]string{
		"focus_theme=" + lens.FocusTheme,
		"support_theme=" + lens.SupportTheme,
		"caution_theme=" + lens.CautionTheme,
		"pacing_theme=" + lens.PacingTheme,
	}, "；")
}

func categoryLabel(category string) string {
	labels := map[string]string{
		"career":       "事业/计划",
		"relationship": "人际/关系",
		"study":        "学习/成长",
		"decision":     "决策/选择",
		"general":      "综合问题",
	}
	if label, ok := labels[category]; ok {
		return label
	}
	return labels["general"]
}

func timeBucketLabel(bucket string) string {
	labels := map[string]string{
		"morning": "上午",
		"day":     "白天",
		"evening": "傍晚",
		"night":   "夜间",
	}
	if label, ok := labels[bucket]; ok {
		return label
	}
	return ""
}

const maxFreeContentPromptRunes = 320

func summarizeFreeContentForPrompt(freeContent string) string {
	freeContent = strings.TrimSpace(freeContent)
	if freeContent == "" {
		return "（无）"
	}
	runes := []rune(freeContent)
	if len(runes) <= maxFreeContentPromptRunes {
		return freeContent
	}
	return string(runes[:maxFreeContentPromptRunes]) + "…（已截断，完整内容见结构化字段）"
}
