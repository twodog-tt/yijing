package qimen

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	fullReportDisclaimerV1 = "本报告基于 qimen-simple-v1 简化奇门文化规则生成，仅用于传统文化学习与自我反思，不等同于专业奇门排盘，不生成完整九宫盘，不构成现实决策依据。"
	fullReportDisclaimerV2 = "本报告基于 qimen-v2-poc 九宫结构 POC 生成，仅用于传统文化学习与结构化观察，不等同于专业奇门排盘，不构成现实决策依据。"
)

type parsedResultPayload struct {
	AlgorithmVersion    string                  `json:"algorithm_version"`
	MethodNote          string                  `json:"method_note"`
	Category            string                  `json:"category"`
	TimeContext         *timeContextPayload     `json:"time_context"`
	QuestionProfile     *profilePayload         `json:"question_profile"`
	QimenLens           *lensPayload            `json:"qimen_lens"`
	SituationOverview   string                  `json:"situation_overview"`
	RiskObservations    []string                `json:"risk_observations"`
	ActionPacing        string                  `json:"action_pacing"`
	ReflectionQuestions []string                `json:"reflection_questions"`
	CalculationMeta     *calculationMetaPayload `json:"calculation_meta"`
	SafeQuestionSummary string                  `json:"safe_question_summary"`
	CalendarBasis       *CalendarBasis          `json:"calendar_basis"`
	Dun                 *Dun                    `json:"dun"`
	Xun                 *Xun                    `json:"xun"`
	Chief               *Chief                  `json:"chief"`
	Palaces             []Palace                `json:"palaces"`
}

const (
	sectionSummary    = "【一、问题局势摘要】"
	sectionFocus      = "【二、关注主题】"
	sectionSupport    = "【三、可借助的条件】"
	sectionRisks      = "【四、主要风险与阻力】"
	sectionPacing     = "【五、行动节奏建议】"
	sectionReflection = "【六、自我反思问题】"
	sectionBoundary   = "【七、边界声明】"
)

// BuildFullContent generates a structured template full report from stored analysis payloads.
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
	profile := profileFromPayload(parsed.QuestionProfile, category)
	lens := lensFromPayload(parsed.QimenLens, profile, category)
	categoryText := categoryLabel(category)
	disclaimer := fullReportDisclaimerFor(parsed.AlgorithmVersion)

	if parsed.AlgorithmVersion == AlgorithmVersionQimenV2POC {
		return buildQimenV2FallbackFullContent(parsed, profile, lens, category, categoryText, methodNote, disclaimer, freeContent), nil
	}

	sections := []string{
		sectionSummary + "\n" + buildQimenSummarySection(parsed, profile, lens, category, categoryText, methodNote, disclaimer),
		sectionFocus + "\n" + buildQimenFocusSection(profile, lens, category),
		sectionSupport + "\n" + buildQimenSupportSection(profile, lens, category),
		sectionRisks + "\n" + buildQimenRiskSection(parsed.RiskObservations, profile, lens, category),
		sectionPacing + "\n" + buildQimenPacingSection(parsed.ActionPacing, profile, lens, category),
		sectionReflection + "\n" + buildQimenReflectionSection(parsed.ReflectionQuestions, profile, category),
		sectionBoundary + "\n" + buildQimenBoundarySection(methodNote, parsed.CalculationMeta, parsed.AlgorithmVersion, parsed.Palaces),
	}

	if snippet := strings.TrimSpace(freeContent); snippet != "" {
		sections[0] = sectionSummary + "\n" + buildQimenSummarySection(parsed, profile, lens, category, categoryText, methodNote, disclaimer) +
			"\n\n可参考免费解读中的局势梳理要点，继续延伸记录与复盘。"
	}

	return strings.Join(sections, "\n\n"), nil
}

func fullReportDisclaimerFor(algorithmVersion string) string {
	if algorithmVersion == AlgorithmVersionQimenV2POC {
		return fullReportDisclaimerV2
	}
	return fullReportDisclaimerV1
}

func buildQimenSummarySection(parsed parsedResultPayload, profile QuestionProfile, lens QimenLens, category, categoryText, methodNote, disclaimer string) string {
	timeNote := ""
	if parsed.TimeContext != nil {
		if label := timeBucketLabel(strings.TrimSpace(parsed.TimeContext.TimeBucket)); label != "" {
			timeNote = fmt.Sprintf("时段参考：%s。", label)
		}
	}
	safeSummary := strings.TrimSpace(parsed.SafeQuestionSummary)
	if safeSummary == "" {
		safeSummary = BuildSafeQuestionSummary(profile)
	}
	return strings.Join([]string{
		disclaimer,
		methodNote,
		fmt.Sprintf("问事分类：%s。%s", categoryText, timeNote),
		fmt.Sprintf("问事特征：%s", safeSummary),
		fmt.Sprintf("问事摘要：%s", QuestionSummary),
		parsed.SituationOverview,
		categorySummaryHint(category, profile),
	}, "\n")
}

func categorySummaryHint(category string, profile QuestionProfile) string {
	switch category {
	case "career":
		return fmt.Sprintf("就事业/计划而言，当前更适合先整理推进顺序与资源协调，再围绕「%s」安排小动作。", profile.IntentType)
	case "relationship":
		return fmt.Sprintf("就人际/关系而言，当前更适合先梳理与「%s」相关的沟通节奏与边界。", profile.IntentType)
	case "study":
		return "就学习/成长而言，当前更适合先复盘方法与精力，再设定阶段目标。"
	case "decision":
		return "就决策/选择而言，当前更适合先补齐信息与选项约束，再小步试探。"
	default:
		return fmt.Sprintf("就综合问题而言，当前更适合先整理问题与风险，再安排与「%s」相关的小步行动。", profile.IntentType)
	}
}

func buildQimenFocusSection(profile QuestionProfile, lens QimenLens, category string) string {
	focusLine := fmt.Sprintf(
		"关注主题：%s\n问事侧重：%s（%s，决策压力%s）\n风险倾向：%s",
		lens.FocusTheme,
		profile.IntentType,
		profile.TimeHorizon,
		profile.DecisionPressure,
		profile.RiskTone,
	)
	switch category {
	case "career":
		focusLine += "\n本类问事重点：推进顺序、资源协调、执行边界。"
	case "relationship":
		focusLine += "\n本类问事重点：沟通节奏、关系边界、误解修复。"
	case "study":
		focusLine += "\n本类问事重点：复盘方法、专注节奏、阶段目标。"
	case "decision":
		focusLine += "\n本类问事重点：信息补齐、小步试探、备用方案。"
	default:
		focusLine += "\n本类问事重点：问题整理、风险收敛、小步行动。"
	}
	return focusLine
}

func buildQimenSupportSection(profile QuestionProfile, lens QimenLens, category string) string {
	lines := []string{
		lens.SupportTheme + "。",
		fmt.Sprintf("可借助「%s」相关已有资源，先明确最小可验证一步。", profile.IntentType),
	}
	switch category {
	case "career":
		lines = append(lines, "把目标拆成「本周一件可完成的小事」，先验证推进顺序是否合理。")
	case "relationship":
		lines = append(lines, "先记录一次互动感受，再决定是否调整沟通方式或边界。")
	case "study":
		lines = append(lines, "设定一段 25–40 分钟的专注块，并记录收获与卡点。")
	case "decision":
		lines = append(lines, "列出选项、约束与可接受的最坏情况，再选低成本试探。")
	default:
		lines = append(lines, "安排一件今天能完成的小事，建立可控感。")
	}
	return strings.Join(lines, "\n")
}

func buildQimenRiskSection(risks []string, profile QuestionProfile, lens QimenLens, category string) string {
	items := append([]string{lens.CautionTheme + "。"}, risks...)
	switch category {
	case "career":
		items = append(items, "执行分散、节奏过快时，可能一次承担过多线程。")
	case "relationship":
		items = append(items, "把一次互动放大成整体判断，可能增加误解。")
	case "study":
		items = append(items, "只比较结果而忽略方法，可能让学习节奏失衡。")
	case "decision":
		items = append(items, "在选项未写清前急于选择，可能遗漏关键约束。")
	default:
		items = append(items, "把简化规则当作确定答案，可能偏离自我反思初衷。")
	}
	if profile.DecisionPressure == "高" {
		items = append(items, "犹豫与担心反复切换，可能让行动迟迟无法启动。")
	}
	if len(items) > 4 {
		items = items[:4]
	}
	return strings.Join(items, "\n")
}

func buildQimenPacingSection(pacing string, profile QuestionProfile, lens QimenLens, category string) string {
	if strings.TrimSpace(pacing) == "" {
		pacing = lens.PacingTheme + "：先观察再行动。"
	}
	extra := categoryPacingHint(category, profile)
	return pacing + "\n\n" + extra
}

func categoryPacingHint(category string, profile QuestionProfile) string {
	switch category {
	case "career":
		return fmt.Sprintf("节奏提示：先整理现状与目标，再列出与「%s」相关的一件小动作，最后定期复盘。", profile.IntentType)
	case "relationship":
		return fmt.Sprintf("节奏提示：先记录感受，再明确一次与「%s」相关的沟通边界，最后观察变化。", profile.IntentType)
	case "study":
		return "节奏提示：先复盘状态，再设定可完成的学习块，最后记录收获。"
	case "decision":
		return "节奏提示：先写下选项与约束，再选一个小步验证，最后比较反馈。"
	default:
		return fmt.Sprintf("节奏提示：先观察，再安排与「%s」相关的一件小事。", profile.IntentType)
	}
}

func buildQimenReflectionSection(questions []string, profile QuestionProfile, category string) string {
	if len(questions) == 0 {
		questions = []string{
			"此刻我最需要整理的是情绪、信息还是行动？",
			"如果把问题拆小，第一步可以是什么？",
		}
	}
	extra := reflectionExtraForCategory(category, profile)
	return strings.Join(append(questions, extra), "\n")
}

func reflectionExtraForCategory(category string, profile QuestionProfile) string {
	switch category {
	case "career":
		return fmt.Sprintf("我真正想推进的「%s」核心目标是什么？", profile.IntentType)
	case "relationship":
		return "我期待的关系互动方式是什么？"
	case "study":
		return "当前学习方法是否匹配我的精力状态？"
	case "decision":
		return "我还缺哪一条信息，才能做更稳妥的比较？"
	default:
		return "我是否把简化解读当作确定结果，而忽视了现实验证？"
	}
}

func buildQimenBoundarySection(methodNote string, meta *calculationMetaPayload, algorithmVersion string, palaces []Palace) string {
	limits := calculationLimits
	if meta != nil && len(meta.Limits) > 0 {
		limits = meta.Limits
	}
	lines := []string{
		fullReportDisclaimerFor(algorithmVersion),
		methodNote,
		"本报告不构成现实决策依据，不做精准预测、强吉凶判断、改运化解，也不提供投资/医疗/法律/赌博/军事建议。",
		"规则限制：" + strings.Join(limits, "；"),
	}
	if algorithmVersion == AlgorithmVersionQimenV2POC && len(palaces) > 0 {
		lines = append(lines, fmt.Sprintf("九宫结构为 POC 近似排盘（共 %d 宫），仅供结构化观察。", len(palaces)))
	}
	return strings.Join(lines, "\n")
}
