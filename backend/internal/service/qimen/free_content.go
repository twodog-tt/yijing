package qimen

import "strings"

const (
	freeDisclaimerV1 = "本解读基于 qimen-simple-v1 简化规则，仅供传统文化学习与自我反思参考，不等同于专业奇门排盘，不构成现实决策依据。"
	freeDisclaimerV2 = "本解读基于 qimen-v2-poc 九宫结构 POC，仅供传统文化学习与结构化观察，不等同于专业奇门排盘，不构成现实决策依据。"
)

func BuildFreeContent(calc CalculationResult) string {
	return buildFreeContent(calc, "")
}

func BuildFreeContentForVersion(calc CalculationResult, algorithmVersion string) string {
	return buildFreeContent(calc, algorithmVersion)
}

func buildFreeContent(calc CalculationResult, algorithmVersion string) string {
	disclaimer := freeDisclaimerV1
	switch algorithmVersion {
	case AlgorithmVersionQimenV2POC:
		disclaimer = freeDisclaimerV2
	case AlgorithmVersionQimenV2Professional:
		disclaimer = freeDisclaimerProfessional
	}
	lensSection := buildLensSection(calc.QimenLens, calc.QuestionProfile)
	sections := []string{
		lensSection,
		"【局势梳理】\n" + calc.SituationOverview,
		"【风险观察】\n" + strings.Join(calc.RiskObservations, "\n"),
		"【行动节奏】\n" + calc.ActionPacing,
		"【自我反思问题】\n" + formatBulletLines(calc.ReflectionQuestions),
		"【行动建议】\n" + formatBulletLines(calc.ActionSuggestions),
		"【免责声明】\n" + disclaimer,
	}
	return strings.Join(sections, "\n\n")
}

func buildLensSection(lens QimenLens, profile QuestionProfile) string {
	lines := []string{
		"【关注主题】",
		"· 关注主题：" + lens.FocusTheme,
		"· 问事侧重：" + profile.IntentType + "（" + profile.TimeHorizon + "，决策压力" + profile.DecisionPressure + "）",
		"· 可借助：" + lens.SupportTheme,
		"· 需留意：" + lens.CautionTheme,
		"· 行动节奏：" + lens.PacingTheme,
	}
	return strings.Join(lines, "\n")
}

func formatBulletLines(items []string) string {
	if len(items) == 0 {
		return "—"
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "·") {
			line = "· " + line
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return "—"
	}
	return strings.Join(lines, "\n")
}
