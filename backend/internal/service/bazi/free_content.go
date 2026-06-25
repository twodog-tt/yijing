package bazi

import (
	"fmt"
	"strings"
)

func BuildFreeContent(calc CalculationResult) string {
	summary := buildSummary(calc)
	elementLearning := buildElementLearning(calc)
	tendency := buildTendency(calc)
	reflection := buildReflectionSection(calc)
	disclaimer := "免责声明：本内容为基于简化干支文化规则的学习参考，可以作为自我观察和行动整理的一个角度，不构成现实决策依据；不用于命运论断、姻缘财运类预测或健康寿命判断。"

	sections := []string{
		"【一句话总结】\n" + summary,
		"【五行倾向学习】\n" + elementLearning,
		"【性格/行动倾向】\n" + tendency,
		"【自我反思与行动建议】\n" + reflection,
		"【免责声明】\n" + disclaimer,
	}
	return strings.Join(sections, "\n\n")
}

func buildSummary(calc CalculationResult) string {
	parts := []string{calc.Pillars.Year, calc.Pillars.Month, calc.Pillars.Day}
	if calc.Pillars.Hour != "" {
		parts = append(parts, calc.Pillars.Hour)
	}
	pillarText := strings.Join(parts, " ")
	if calc.BirthHourUnknown {
		return fmt.Sprintf("基于简化干支文化规则，简化干支示意为 %s，时辰未知，本次不生成时柱，日主为「%s」。", pillarText, calc.DayMaster)
	}
	return fmt.Sprintf("基于简化干支文化规则，简化干支示意为 %s，日主为「%s」。", pillarText, calc.DayMaster)
}

func buildElementLearning(calc CalculationResult) string {
	e := calc.FiveElements
	return fmt.Sprintf(
		"五行计数（学习参考）：木 %d、火 %d、土 %d、金 %d、水 %d。该分布来自简化规则下的干支示意，不等同于专业旺衰判断。",
		e.Wood, e.Fire, e.Earth, e.Metal, e.Water,
	)
}

func buildTendency(calc CalculationResult) string {
	return fmt.Sprintf(
		"日主「%s」可作为一个自我观察切入点：%s",
		calc.DayMaster,
		calc.ReflectionFocus,
	)
}

func buildReflectionSection(calc CalculationResult) string {
	lines := append([]string{}, calc.ActionSuggestions...)
	if calc.BirthHourUnknown {
		lines = append([]string{"时辰未知，本次不生成时柱；可从年月日三柱示意出发做自我观察。"}, lines...)
	}
	return strings.Join(lines, "\n")
}
