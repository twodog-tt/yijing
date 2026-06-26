package qimen

import "encoding/json"

// MergeV1InterpretationWithProfessional keeps v1 interpretation fields and overlays professional method/limits.
func MergeV1InterpretationWithProfessional(v1 CalculationResult, pro *CalculationResultV2Professional) CalculationResult {
	if pro == nil {
		return v1
	}
	merged := v1
	merged.MethodNote = pro.MethodNote
	merged.Limits = append([]string(nil), pro.Limits...)
	return merged
}

// BuildProfessionalAPIResultPayload produces a client-compatible result_payload for qimen-v2-professional.
// It merges v1 interpretation fields with professional nine-palace structure for existing result pages.
func BuildProfessionalAPIResultPayload(v1 CalculationResult, pro *CalculationResultV2Professional) (json.RawMessage, error) {
	if pro == nil {
		return nil, ErrInvalidParams
	}
	payload := map[string]any{
		"algorithm_version":     AlgorithmVersionQimenV2Professional,
		"layout_version":        pro.LayoutVersion,
		"layout_basis":          pro.LayoutBasis,
		"method_note":           pro.MethodNote,
		"question_summary":      QuestionSummary,
		"category":              v1.Category,
		"time_context":          TimeContext{CreatedAt: v1.TimeContext.CreatedAt, TimeBucket: v1.TimeContext.TimeBucket},
		"question_profile":      v1.QuestionProfile,
		"qimen_lens":            v1.QimenLens,
		"differentiation_seed":  v1.DifferentiationSeed,
		"safe_question_summary": v1.SafeQuestionSummary,
		"situation_overview":    v1.SituationOverview,
		"risk_observations":     v1.RiskObservations,
		"action_pacing":         v1.ActionPacing,
		"reflection_questions":  v1.ReflectionQuestions,
		"action_suggestions":    v1.ActionSuggestions,
		"calendar_basis":        pro.CalendarBasis,
		"dun":                   pro.Dun,
		"ganzhi":                pro.Ganzhi,
		"xun":                   pro.Xun,
		"chief":                 pro.Chief,
		"palaces":               pro.Palaces,
		"limits":                pro.Limits,
		"calculation_meta": map[string]any{
			"limits": pro.Limits,
		},
	}
	return json.Marshal(payload)
}

func professionalPalacesToPalaces(proPalaces []ProfessionalPalace) []Palace {
	if len(proPalaces) == 0 {
		return nil
	}
	out := make([]Palace, len(proPalaces))
	for i, p := range proPalaces {
		out[i] = Palace{
			Index:           p.Index,
			Name:            p.Name,
			EarthPlateStem:  p.EarthPlateStem,
			HeavenPlateStem: p.HeavenPlateStem,
			Star:            p.Star,
			Door:            p.Door,
			Deity:           p.Deity,
			Summary:         p.Summary,
		}
	}
	return out
}

func professionalChiefToChief(chief ProfessionalChief) Chief {
	return Chief{
		ZhiFu:  chief.ZhiFu,
		ZhiShi: chief.ZhiShi,
	}
}
