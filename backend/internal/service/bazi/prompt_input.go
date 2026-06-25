package bazi

import (
	"encoding/json"
	"fmt"
	"strings"
)

type calculationMeta struct {
	Limits []string `json:"limits"`
}

type fullReportPromptInput struct {
	MethodNote        string
	Pillars           Pillars
	HourUnknown       bool
	DayMaster         string
	FiveElements      FiveElements
	ReflectionFocus   string
	ActionSuggestions []string
	Limits            []string
	FreeContent       string
}

func buildFullReportPromptInput(resultPayload json.RawMessage, freeContent string) (*fullReportPromptInput, error) {
	var parsed struct {
		MethodNote        string           `json:"method_note"`
		Pillars           Pillars          `json:"pillars"`
		DayMaster         string           `json:"day_master"`
		FiveElements      FiveElements     `json:"five_elements"`
		ReflectionFocus   string           `json:"reflection_focus"`
		ActionSuggestions []string         `json:"action_suggestions"`
		CalculationMeta   *calculationMeta `json:"calculation_meta"`
	}
	if err := json.Unmarshal(resultPayload, &parsed); err != nil {
		return nil, fmt.Errorf("invalid result_payload")
	}
	if strings.TrimSpace(parsed.DayMaster) == "" {
		return nil, fmt.Errorf("invalid result_payload")
	}

	input := &fullReportPromptInput{
		MethodNote:        strings.TrimSpace(parsed.MethodNote),
		Pillars:           parsed.Pillars,
		HourUnknown:       strings.TrimSpace(parsed.Pillars.Hour) == "",
		DayMaster:         strings.TrimSpace(parsed.DayMaster),
		FiveElements:      parsed.FiveElements,
		ReflectionFocus:   strings.TrimSpace(parsed.ReflectionFocus),
		ActionSuggestions: append([]string{}, parsed.ActionSuggestions...),
		FreeContent:       strings.TrimSpace(freeContent),
	}
	if parsed.CalculationMeta != nil {
		input.Limits = append([]string{}, parsed.CalculationMeta.Limits...)
	}
	if input.MethodNote == "" {
		input.MethodNote = MethodNote
	}
	return input, nil
}
