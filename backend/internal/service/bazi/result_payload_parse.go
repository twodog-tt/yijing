package bazi

import (
	"encoding/json"
	"fmt"
	"strings"
)

type calculationMeta struct {
	Limits []string `json:"limits"`
}

type profilePayload struct {
	DayMasterObservation string `json:"day_master_observation"`
	SeasonTendency       string `json:"season_tendency"`
	ElementBalanceType   string `json:"element_balance_type"`
	ActionStyle          string `json:"action_style"`
	ReflectionTheme      string `json:"reflection_theme"`
}

type lensPayload struct {
	StrengthHint             string `json:"strength_hint"`
	CautionHint              string `json:"caution_hint"`
	PacingHint               string `json:"pacing_hint"`
	RelationshipWithElements string `json:"relationship_with_elements"`
}

type storedResultPayload struct {
	AlgorithmVersion   string
	MethodNote         string
	Pillars            Pillars
	CalendarBasis      CalendarBasis
	BaziProfile        *profilePayload
	InterpretationLens *lensPayload
	ReflectionFocus    string
	ActionSuggestions  []string
	CalculationMeta    *calculationMeta
	DayMaster          string
	FiveElements       FiveElements
}

func parseStoredResultPayload(raw json.RawMessage) (storedResultPayload, error) {
	var parsed struct {
		AlgorithmVersion   string           `json:"algorithm_version"`
		MethodNote         string           `json:"method_note"`
		Pillars            Pillars          `json:"pillars"`
		PillarsV2          Pillars          `json:"pillars_v2"`
		CalendarBasis      CalendarBasis    `json:"calendar_basis"`
		BaziProfile        *profilePayload  `json:"bazi_profile"`
		InterpretationLens *lensPayload     `json:"interpretation_lens"`
		ReflectionFocus    string           `json:"reflection_focus"`
		ActionSuggestions  []string         `json:"action_suggestions"`
		CalculationMeta    *calculationMeta `json:"calculation_meta"`
		DayMaster          string           `json:"day_master"`
		FiveElements       FiveElements     `json:"five_elements"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return storedResultPayload{}, fmt.Errorf("invalid result_payload")
	}

	pillars := parsed.Pillars
	if strings.TrimSpace(pillars.Year) == "" && strings.TrimSpace(parsed.PillarsV2.Year) != "" {
		pillars = parsed.PillarsV2
	}
	if strings.TrimSpace(parsed.DayMaster) == "" {
		return storedResultPayload{}, fmt.Errorf("invalid result_payload")
	}

	return storedResultPayload{
		AlgorithmVersion:   strings.TrimSpace(parsed.AlgorithmVersion),
		MethodNote:         parsed.MethodNote,
		Pillars:            pillars,
		CalendarBasis:      parsed.CalendarBasis,
		BaziProfile:        parsed.BaziProfile,
		InterpretationLens: parsed.InterpretationLens,
		ReflectionFocus:    parsed.ReflectionFocus,
		ActionSuggestions:  append([]string{}, parsed.ActionSuggestions...),
		CalculationMeta:    parsed.CalculationMeta,
		DayMaster:          strings.TrimSpace(parsed.DayMaster),
		FiveElements:       parsed.FiveElements,
	}, nil
}

func fullReportDisclaimerFor(algorithmVersion string) string {
	if algorithmVersion == AlgorithmVersionBaziV2POC {
		return "本报告基于 bazi-v2-poc 节气与立春换年规则生成，仅用于传统文化学习与自我反思，不等同于专业八字排盘，节令时刻为公式近似，不构成现实决策依据。"
	}
	return fullReportDisclaimer
}

func formatCalendarBasisForPrompt(basis CalendarBasis) string {
	if strings.TrimSpace(basis.YearBoundary) == "" && strings.TrimSpace(basis.MonthBoundary) == "" {
		return "（无）"
	}
	parts := []string{
		"year_boundary=" + basis.YearBoundary,
		"month_boundary=" + basis.MonthBoundary,
		"true_solar_time=" + fmt.Sprintf("%t", basis.TrueSolarTime),
		"day_pillar_basis=" + basis.DayPillarBasis,
	}
	if note := strings.TrimSpace(basis.Note); note != "" {
		parts = append(parts, "note="+note)
	}
	return strings.Join(parts, "；")
}
