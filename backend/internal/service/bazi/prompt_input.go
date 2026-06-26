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

type fullReportPromptInput struct {
	MethodNote          string
	Pillars             Pillars
	HourUnknown         bool
	DayMaster           string
	FiveElements        FiveElements
	BaziProfile         BaziProfile
	InterpretationLens  InterpretationLens
	ReflectionFocus     string
	ActionSuggestions   []string
	Limits              []string
	FreeContent         string
}

func buildFullReportPromptInput(resultPayload json.RawMessage, freeContent string) (*fullReportPromptInput, error) {
	var parsed struct {
		MethodNote          string           `json:"method_note"`
		Pillars             Pillars          `json:"pillars"`
		DayMaster           string           `json:"day_master"`
		FiveElements        FiveElements     `json:"five_elements"`
		BaziProfile         *profilePayload  `json:"bazi_profile"`
		InterpretationLens  *lensPayload     `json:"interpretation_lens"`
		ReflectionFocus     string           `json:"reflection_focus"`
		ActionSuggestions   []string         `json:"action_suggestions"`
		CalculationMeta     *calculationMeta `json:"calculation_meta"`
	}
	if err := json.Unmarshal(resultPayload, &parsed); err != nil {
		return nil, fmt.Errorf("invalid result_payload")
	}
	if strings.TrimSpace(parsed.DayMaster) == "" {
		return nil, fmt.Errorf("invalid result_payload")
	}

	hourUnknown := strings.TrimSpace(parsed.Pillars.Hour) == ""
	profile := profileFromPayload(parsed.BaziProfile, parsed.DayMaster, parsed.FiveElements, hourUnknown)
	lens := lensFromPayload(parsed.InterpretationLens, profile, parsed.DayMaster, parsed.FiveElements, hourUnknown)

	input := &fullReportPromptInput{
		MethodNote:          strings.TrimSpace(parsed.MethodNote),
		Pillars:             parsed.Pillars,
		HourUnknown:         hourUnknown,
		DayMaster:           strings.TrimSpace(parsed.DayMaster),
		FiveElements:        parsed.FiveElements,
		BaziProfile:         profile,
		InterpretationLens:  lens,
		ReflectionFocus:     strings.TrimSpace(parsed.ReflectionFocus),
		ActionSuggestions:   append([]string{}, parsed.ActionSuggestions...),
		FreeContent:         summarizeFreeContentForPrompt(freeContent),
	}
	if parsed.CalculationMeta != nil {
		input.Limits = append([]string{}, parsed.CalculationMeta.Limits...)
	}
	if input.MethodNote == "" {
		input.MethodNote = MethodNote
	}
	return input, nil
}

func profileFromPayload(p *profilePayload, dayMaster string, elements FiveElements, hourUnknown bool) BaziProfile {
	if p == nil {
		return BuildBaziProfile(1, dayMaster, elements, hourUnknown, Pillars{})
	}
	profile := BaziProfile{
		DayMasterObservation: strings.TrimSpace(p.DayMasterObservation),
		SeasonTendency:       strings.TrimSpace(p.SeasonTendency),
		ElementBalanceType:   strings.TrimSpace(p.ElementBalanceType),
		ActionStyle:          strings.TrimSpace(p.ActionStyle),
		ReflectionTheme:      strings.TrimSpace(p.ReflectionTheme),
	}
	if profile.DayMasterObservation == "" {
		return BuildBaziProfile(1, dayMaster, elements, hourUnknown, Pillars{})
	}
	return profile
}

func lensFromPayload(p *lensPayload, profile BaziProfile, dayMaster string, elements FiveElements, hourUnknown bool) InterpretationLens {
	if p == nil {
		return BuildInterpretationLens(profile, elements, dayMaster, hourUnknown)
	}
	lens := InterpretationLens{
		StrengthHint:             strings.TrimSpace(p.StrengthHint),
		CautionHint:              strings.TrimSpace(p.CautionHint),
		PacingHint:               strings.TrimSpace(p.PacingHint),
		RelationshipWithElements: strings.TrimSpace(p.RelationshipWithElements),
	}
	if lens.StrengthHint == "" {
		return BuildInterpretationLens(profile, elements, dayMaster, hourUnknown)
	}
	if lens.CautionHint == "" {
		lens.CautionHint = cautionHintFor(profile, dayMaster, "")
	}
	if lens.PacingHint == "" {
		lens.PacingHint = pacingHintFor(profile, hourUnknown)
	}
	if lens.RelationshipWithElements == "" {
		dominant, weak := dominantElements(elements)
		lens.RelationshipWithElements = elementRelationshipHint(elements, dominant, weak)
	}
	return lens
}

func formatBaziProfileForPrompt(profile BaziProfile) string {
	return strings.Join([]string{
		"day_master_observation=" + profile.DayMasterObservation,
		"season_tendency=" + profile.SeasonTendency,
		"element_balance_type=" + profile.ElementBalanceType,
		"action_style=" + profile.ActionStyle,
		"reflection_theme=" + profile.ReflectionTheme,
	}, "；")
}

func formatInterpretationLensForPrompt(lens InterpretationLens) string {
	return strings.Join([]string{
		"strength_hint=" + lens.StrengthHint,
		"caution_hint=" + lens.CautionHint,
		"pacing_hint=" + lens.PacingHint,
		"relationship_with_elements=" + lens.RelationshipWithElements,
	}, "；")
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
