package qimen

import "encoding/json"

// MergeV1InterpretationWithV2 keeps v1 interpretation fields and overlays v2 method/limits.
func MergeV1InterpretationWithV2(v1 CalculationResult, v2 CalculationResultV2) CalculationResult {
	merged := v1
	merged.MethodNote = v2.MethodNote
	merged.Limits = append([]string(nil), v2.Limits...)
	return merged
}

// BuildV2APIResultPayload produces a client-compatible result_payload for qimen-v2-poc.
// It merges v1 interpretation fields with v2 nine-palace structure for existing result pages.
func BuildV2APIResultPayload(v1 CalculationResult, v2 CalculationResultV2) (json.RawMessage, error) {
	payload := map[string]any{
		"algorithm_version":     AlgorithmVersionQimenV2POC,
		"method_note":           v2.MethodNote,
		"question_summary":      QuestionSummary,
		"category":              v1.Category,
		"time_context":          v2.TimeContext,
		"question_profile":      v1.QuestionProfile,
		"qimen_lens":            v1.QimenLens,
		"differentiation_seed":  v1.DifferentiationSeed,
		"safe_question_summary": v1.SafeQuestionSummary,
		"situation_overview":    v1.SituationOverview,
		"risk_observations":     v1.RiskObservations,
		"action_pacing":         v1.ActionPacing,
		"reflection_questions":  v1.ReflectionQuestions,
		"action_suggestions":    v1.ActionSuggestions,
		"calendar_basis":        v2.CalendarBasis,
		"dun":                   v2.Dun,
		"xun":                   v2.Xun,
		"chief":                 v2.Chief,
		"palaces":               v2.Palaces,
		"limits":                v2.Limits,
		"calculation_meta": map[string]any{
			"limits": v2.Limits,
		},
	}
	return json.Marshal(payload)
}
