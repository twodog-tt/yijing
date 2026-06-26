package qimen

import (
	"encoding/json"
	"fmt"
	"strings"
)

const fullReportDisclaimer = "本报告基于 qimen-simple-v1 简化奇门文化规则生成，仅用于传统文化学习与自我反思，不等同于专业奇门排盘，不生成完整九宫盘，不构成现实决策依据。"

type parsedResultPayload struct {
	MethodNote          string                  `json:"method_note"`
	QuestionSummary     string                  `json:"question_summary"`
	SafeQuestionSummary string                  `json:"safe_question_summary"`
	Category            string                  `json:"category"`
	TimeContext         *timeContextPayload     `json:"time_context"`
	QuestionProfile     *profilePayload         `json:"question_profile"`
	QimenLens           *lensPayload            `json:"qimen_lens"`
	SituationOverview   string                  `json:"situation_overview"`
	RiskObservations    []string                `json:"risk_observations"`
	ActionPacing        string                  `json:"action_pacing"`
	ReflectionQuestions []string                `json:"reflection_questions"`
	ActionSuggestions   []string                `json:"action_suggestions"`
	CalculationMeta     *calculationMetaPayload `json:"calculation_meta"`
}

// BuildFullContent generates a template full report from stored analysis payloads.
func BuildFullContent(resultPayload json.RawMessage, freeContent string) (string, error) {
	var parsed parsedResultPayload
	if err := json.Unmarshal(resultPayload, &parsed); err != nil {
		return "", fmt.Errorf("invalid result_payload")
	}
	if strings.TrimSpace(parsed.SituationOverview) == "" {
		return "", fmt.Errorf("invalid result_payload")
	}

	methodNote := strings.TrimSpace(parsed.MethodNote)
	if methodNote == "" {
		methodNote = MethodNote
	}

	category := NormalizeCategory(parsed.Category)
	categoryText := categoryLabel(category)
	timeNote := ""
	if parsed.TimeContext != nil {
		if label := timeBucketLabel(strings.TrimSpace(parsed.TimeContext.TimeBucket)); label != "" {
			timeNote = fmt.Sprintf("时段参考：%s。", label)
		}
	}

	summary := QuestionSummary

	risks := parsed.RiskObservations
	if len(risks) == 0 {
		risks = []string{
			"过度依赖单一结论，可能忽略现实细节与变化。",
			"在信息不完整时仓促行动，容易放大情绪波动。",
		}
	}

	pacing := strings.TrimSpace(parsed.ActionPacing)
	if pacing == "" {
		pacing = "建议分步整理：先观察现状，再安排一件可执行的小事，最后记录感受与下一步。"
	}

	reflections := parsed.ReflectionQuestions
	if len(reflections) == 0 {
		reflections = []string{
			"此刻我最需要整理的是情绪、信息还是行动？",
			"如果把问题拆小，第一步可以是什么？",
		}
	}

	suggestions := parsed.ActionSuggestions
	if len(suggestions) == 0 {
		suggestions = []string{
			"安排一件今天能完成的小事，建立可控感。",
			"把问题改写成「我想观察什么」而非「结果会怎样」。",
		}
	}

	limits := calculationLimits
	if parsed.CalculationMeta != nil && len(parsed.CalculationMeta.Limits) > 0 {
		limits = parsed.CalculationMeta.Limits
	}

	profile := profileFromPayload(parsed.QuestionProfile, category)
	lens := lensFromPayload(parsed.QimenLens, profile, category)
	safeSummary := strings.TrimSpace(parsed.SafeQuestionSummary)
	if safeSummary == "" {
		safeSummary = BuildSafeQuestionSummary(profile)
	}

	methodSection := fmt.Sprintf(
		"方法说明：%s\n问事分类：%s\n%s问事摘要：%s\n问事特征：%s\n关注主题：%s\n行动节奏倾向：%s\n规则限制：%s",
		methodNote,
		categoryText,
		timeNote,
		summary,
		safeSummary,
		lens.FocusTheme,
		lens.PacingTheme,
		strings.Join(limits, "；"),
	)

	situationSection := parsed.SituationOverview + "\n\n" +
		"可把上述局势理解为一面观察镜：重点不是给出确定结果，而是帮助你在当前阶段做更清晰的自我整理。"

	riskSection := strings.Join(risks, "\n") + "\n\n" +
		"以上风险观察来自简化规则下的通用提醒，请结合你的现实情境自行判断。"

	pacingSection := pacing + "\n\n" +
		"行动节奏建议保持温和、可执行，避免一次性做满所有决定。"

	freeSnippet := strings.TrimSpace(freeContent)
	observationNote := "可参考免费解读中的局势梳理与反思要点，继续延伸记录。"
	if freeSnippet != "" {
		observationNote = "可参考免费解读中的要点，继续延伸记录与复盘。"
	}

	sections := []string{
		"【完整报告说明】\n" + fullReportDisclaimer,
		"【1. 方法说明与简化边界】\n" + methodSection,
		"【2. 局势梳理展开】\n" + situationSection,
		"【3. 风险观察展开】\n" + riskSection,
		"【4. 行动节奏与节奏建议】\n" + pacingSection,
		"【5. 自我反思问题】\n" + strings.Join(reflections, "\n"),
		"【6. 行动建议】\n" + strings.Join(suggestions, "\n"),
		"【7. 观察与延伸】\n" + observationNote,
		"【8. 免责声明】\n" + fullReportDisclaimer + "\n" + methodNote,
	}
	return strings.Join(sections, "\n\n"), nil
}
