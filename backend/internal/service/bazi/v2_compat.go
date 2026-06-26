package bazi

import (
	"encoding/json"
	"time"
)

// CalculationResultFromV2 maps v2 pillars into the shared CalculationResult used by free/full content builders.
func CalculationResultFromV2(v2 CalculationResultV2) CalculationResult {
	birthMonth := 1
	if t, err := time.ParseInLocation("2006-01-02", v2.BirthDate, time.Local); err == nil {
		birthMonth = int(t.Month())
	}

	pillars := v2.PillarsV2
	profile := BuildBaziProfile(birthMonth, v2.DayMaster, v2.FiveElements, v2.BirthHourUnknown, pillars)
	lens := BuildInterpretationLens(profile, v2.FiveElements, v2.DayMaster, v2.BirthHourUnknown)

	return CalculationResult{
		BirthDate:           v2.BirthDate,
		BirthHourBranch:     v2.BirthHourBranch,
		BirthHourUnknown:    v2.BirthHourUnknown,
		Pillars:             pillars,
		DayMaster:           v2.DayMaster,
		FiveElements:        v2.FiveElements,
		BaziProfile:         profile,
		InterpretationLens:  lens,
		DifferentiationSeed: BuildDifferentiationSeed(!v2.BirthHourUnknown && pillars.Hour != ""),
		ReflectionFocus:     BuildReflectionFocus(profile, v2.DayMaster),
		ActionSuggestions:   BuildActionSuggestions(profile, lens, v2.DayMaster, v2.FiveElements, v2.BirthHourUnknown),
		MethodNote:          v2.MethodNote,
		Limits:              append([]string{}, v2.Limits...),
	}
}

// BuildV2APIResultPayload produces a client-compatible result_payload for bazi-v2-poc.
func BuildV2APIResultPayload(v2 CalculationResultV2, calc CalculationResult) (json.RawMessage, error) {
	pillars := map[string]string{
		"year":  calc.Pillars.Year,
		"month": calc.Pillars.Month,
		"day":   calc.Pillars.Day,
	}
	pillarsV2 := map[string]string{
		"year":  v2.PillarsV2.Year,
		"month": v2.PillarsV2.Month,
		"day":   v2.PillarsV2.Day,
	}
	if !calc.BirthHourUnknown && calc.Pillars.Hour != "" {
		pillars["hour"] = calc.Pillars.Hour
		pillarsV2["hour"] = v2.PillarsV2.Hour
	}

	payload := map[string]any{
		"algorithm_version":    AlgorithmVersionBaziV2POC,
		"method_note":          calc.MethodNote,
		"calendar_basis":       v2.CalendarBasis,
		"pillars":              pillars,
		"pillars_v2":           pillarsV2,
		"compatibility_note":   v2.CompatibilityNote,
		"day_master":           calc.DayMaster,
		"five_elements":        calc.FiveElements,
		"bazi_profile":         calc.BaziProfile,
		"interpretation_lens":  calc.InterpretationLens,
		"differentiation_seed": calc.DifferentiationSeed,
		"reflection_focus":     calc.ReflectionFocus,
		"action_suggestions":   calc.ActionSuggestions,
	}
	if len(calc.Limits) > 0 {
		payload["calculation_meta"] = map[string]any{"limits": calc.Limits}
	}
	return json.Marshal(payload)
}
