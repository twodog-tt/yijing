package qimen

import (
	"encoding/json"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
)

const (
	MethodNote       = "本功能采用简化奇门文化规则，仅用于传统文化学习与自我反思，不等同于专业奇门排盘。"
	QuestionSummary  = "用户问题已用于本次局势梳理"
	MinQuestionRunes = 4
	MaxQuestionRunes = 120
)

var (
	validCategories = map[string]struct{}{
		"career":       {},
		"relationship": {},
		"study":        {},
		"decision":     {},
		"general":      {},
	}

	calculationLimits = []string{
		"简化学习版，不生成完整九宫盘",
		"不做专业排盘",
		"不构成现实决策依据",
	}
)

type CreateInput struct {
	SessionKey string
	Question   string
	Category   string
	ClientInfo string
}

type TimeContext struct {
	CreatedAt  string `json:"created_at"`
	TimeBucket string `json:"time_bucket"`
}

type CalculationResult struct {
	Question            string
	Category            string
	TimeContext         TimeContext
	QuestionProfile     QuestionProfile
	QimenLens           QimenLens
	DifferentiationSeed DifferentiationSeed
	SafeQuestionSummary string
	SituationOverview   string
	RiskObservations    []string
	ActionPacing        string
	ReflectionQuestions []string
	ActionSuggestions   []string
	MethodNote          string
	Limits              []string
}

func (c CalculationResult) InputPayload() (json.RawMessage, error) {
	payload := map[string]any{
		"question":           c.Question,
		"category":           c.Category,
		"confirm_disclaimer": true,
	}
	return json.Marshal(payload)
}

func (c CalculationResult) ResultPayload() (json.RawMessage, error) {
	payload := map[string]any{
		"algorithm_version":    model.AlgorithmVersionQimenSimpleV1,
		"method_note":          c.MethodNote,
		"question_summary":     QuestionSummary,
		"category":             c.Category,
		"time_context":         c.TimeContext,
		"question_profile":     c.QuestionProfile,
		"qimen_lens":           c.QimenLens,
		"differentiation_seed": c.DifferentiationSeed,
		"safe_question_summary": c.SafeQuestionSummary,
		"situation_overview":   c.SituationOverview,
		"risk_observations":    c.RiskObservations,
		"action_pacing":        c.ActionPacing,
		"reflection_questions": c.ReflectionQuestions,
		"action_suggestions":   c.ActionSuggestions,
		"calculation_meta": map[string]any{
			"limits": c.Limits,
		},
	}
	return json.Marshal(payload)
}

func NormalizeCategory(category string) string {
	category = trimLower(category)
	if category == "" {
		return "general"
	}
	return category
}

func ValidateCategory(category string) bool {
	_, ok := validCategories[category]
	return ok
}

func timeBucketFor(now time.Time) string {
	hour := now.Hour()
	switch {
	case hour >= 5 && hour < 11:
		return "morning"
	case hour >= 11 && hour < 17:
		return "day"
	case hour >= 17 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func SanitizeListItem(item *model.AnalysisListItem) {
	if item == nil || item.ModuleType != model.ModuleTypeQimen {
		return
	}
	summary := QuestionSummary
	item.Question = &summary
}
