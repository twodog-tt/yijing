package bazi

import (
	"encoding/json"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
)

type fullReportPromptInput struct {
	AlgorithmVersion  string
	MethodNote        string
	Pillars           Pillars
	CalendarBasis     CalendarBasis
	HourUnknown       bool
	DayMaster         string
	FiveElements      FiveElements
	BaziProfile       BaziProfile
	InterpretationLens InterpretationLens
	ReflectionFocus   string
	ActionSuggestions []string
	Limits            []string
	FreeContent       string
}

func buildFullReportPromptInput(resultPayload json.RawMessage, freeContent string) (*fullReportPromptInput, error) {
	parsed, err := parseStoredResultPayload(resultPayload)
	if err != nil {
		return nil, err
	}

	hourUnknown := strings.TrimSpace(parsed.Pillars.Hour) == ""
	profile := profileFromPayload(parsed.BaziProfile, parsed.DayMaster, parsed.FiveElements, hourUnknown)
	lens := lensFromPayload(parsed.InterpretationLens, profile, parsed.DayMaster, parsed.FiveElements, hourUnknown)

	input := &fullReportPromptInput{
		AlgorithmVersion:   parsed.AlgorithmVersion,
		MethodNote:         strings.TrimSpace(parsed.MethodNote),
		Pillars:            parsed.Pillars,
		CalendarBasis:      parsed.CalendarBasis,
		HourUnknown:        hourUnknown,
		DayMaster:          parsed.DayMaster,
		FiveElements:       parsed.FiveElements,
		BaziProfile:        profile,
		InterpretationLens: lens,
		ReflectionFocus:    strings.TrimSpace(parsed.ReflectionFocus),
		ActionSuggestions:  append([]string{}, parsed.ActionSuggestions...),
		FreeContent:        summarizeFreeContentForPrompt(freeContent),
	}
	if parsed.CalculationMeta != nil {
		input.Limits = append([]string{}, parsed.CalculationMeta.Limits...)
	}
	if input.MethodNote == "" {
		if input.AlgorithmVersion == AlgorithmVersionBaziV2POC {
			input.MethodNote = MethodNoteV2
		} else {
			input.MethodNote = MethodNote
		}
	}
	if input.AlgorithmVersion == "" {
		input.AlgorithmVersion = model.AlgorithmVersionBaziSimpleV1
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
