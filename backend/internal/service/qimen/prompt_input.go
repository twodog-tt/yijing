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

type fullReportPromptInput struct {
	MethodNote          string
	QuestionSummary     string
	Category            string
	CategoryLabel       string
	TimeBucket          string
	TimeBucketLabel     string
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
		Category            string                  `json:"category"`
		TimeContext         *timeContextPayload     `json:"time_context"`
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

	summary := strings.TrimSpace(parsed.QuestionSummary)
	if summary == "" {
		summary = QuestionSummary
	}
	// Always use the safe constant; never trust stored question_summary for AI prompts.
	summary = QuestionSummary

	input := &fullReportPromptInput{
		MethodNote:          strings.TrimSpace(parsed.MethodNote),
		QuestionSummary:     summary,
		Category:            category,
		CategoryLabel:       categoryLabel(category),
		TimeBucket:          bucket,
		TimeBucketLabel:     timeBucketLabel(bucket),
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
