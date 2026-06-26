package bazi

import (
	"encoding/json"

	"github.com/wangxintong/yijing/backend/internal/model"
)

const (
	MethodNote = "本功能采用简化干支文化规则，不等同于专业八字排盘。"
	Timezone   = "Asia/Shanghai"
	Calendar   = "gregorian"
)

var validHourBranches = map[string]int{
	"zi": 0, "chou": 1, "yin": 2, "mao": 3, "chen": 4, "si": 5,
	"wu": 6, "wei": 7, "shen": 8, "you": 9, "xu": 10, "hai": 11,
}

type CreateInput struct {
	SessionKey       string
	BirthDate        string
	BirthHourBranch  string
	BirthHourUnknown bool
	ClientInfo       string
}

type CalculationResult struct {
	BirthDate           string
	BirthHourBranch     string
	BirthHourUnknown    bool
	Pillars             Pillars
	DayMaster           string
	FiveElements        FiveElements
	BaziProfile         BaziProfile
	InterpretationLens  InterpretationLens
	DifferentiationSeed DifferentiationSeed
	ReflectionFocus     string
	ActionSuggestions   []string
	MethodNote          string
	Limits              []string
}

type Pillars struct {
	Year  string `json:"year"`
	Month string `json:"month"`
	Day   string `json:"day"`
	Hour  string `json:"hour"`
}

type FiveElements struct {
	Wood  int `json:"wood"`
	Fire  int `json:"fire"`
	Earth int `json:"earth"`
	Metal int `json:"metal"`
	Water int `json:"water"`
}

func (c CalculationResult) InputPayload() (json.RawMessage, error) {
	payload := map[string]any{
		"birth_date":         c.BirthDate,
		"birth_hour_unknown": c.BirthHourUnknown,
		"calendar":           Calendar,
		"timezone":           Timezone,
	}
	if !c.BirthHourUnknown {
		payload["birth_hour_branch"] = c.BirthHourBranch
	}
	return json.Marshal(payload)
}

func (c CalculationResult) ResultPayload() (json.RawMessage, error) {
	pillars := map[string]string{
		"year":  c.Pillars.Year,
		"month": c.Pillars.Month,
		"day":   c.Pillars.Day,
	}
	if !c.BirthHourUnknown && c.Pillars.Hour != "" {
		pillars["hour"] = c.Pillars.Hour
	}

	payload := map[string]any{
		"algorithm_version":    model.AlgorithmVersionBaziSimpleV1,
		"method_note":          c.MethodNote,
		"pillars":              pillars,
		"day_master":           c.DayMaster,
		"five_elements":        c.FiveElements,
		"bazi_profile":         c.BaziProfile,
		"interpretation_lens":  c.InterpretationLens,
		"differentiation_seed": c.DifferentiationSeed,
		"reflection_focus":     c.ReflectionFocus,
		"action_suggestions":   c.ActionSuggestions,
	}
	if len(c.Limits) > 0 {
		payload["calculation_meta"] = map[string]any{"limits": c.Limits}
	}
	return json.Marshal(payload)
}
