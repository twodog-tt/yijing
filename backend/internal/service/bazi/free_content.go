package bazi

import (
	"fmt"
	"strings"
)

func BuildFreeContent(calc CalculationResult) string {
	lensSection := buildLensSection(calc.BaziProfile, calc.InterpretationLens)
	pillarSection := buildPillarObservation(calc)
	elementSection := buildElementLearning(calc)
	reflectionSection := buildReflectionSection(calc)
	disclaimer := "免责声明：本内容为基于简化干支文化规则的学习参考，可以作为自我观察和行动整理的一个角度，不构成现实决策依据；不用于命运论断、姻缘财运类预测或健康寿命判断。"

	sections := []string{
		lensSection,
		"【简化干支观察】\n" + pillarSection,
		"【五行倾向】\n" + elementSection,
		"【自我反思重点】\n" + reflectionSection,
		"【免责声明】\n" + disclaimer,
	}
	return strings.Join(sections, "\n\n")
}

func buildLensSection(profile BaziProfile, lens InterpretationLens) string {
	lines := []string{
		"【解读视角】",
		"· 日主观察：" + profile.DayMasterObservation,
		"· 季节倾向：" + profile.SeasonTendency + "（简化参考）",
		"· 五行倾向：" + profile.ElementBalanceType,
		"· 行动风格：" + profile.ActionStyle,
		"· 反思主题：" + profile.ReflectionTheme,
		"· 可借助：" + lens.StrengthHint,
		"· 需留意：" + lens.CautionHint,
		"· 节奏建议：" + lens.PacingHint,
	}
	return strings.Join(lines, "\n")
}

func buildPillarObservation(calc CalculationResult) string {
	parts := []string{calc.Pillars.Year, calc.Pillars.Month, calc.Pillars.Day}
	if calc.Pillars.Hour != "" {
		parts = append(parts, calc.Pillars.Hour)
	}
	pillarText := strings.Join(parts, " ")
	if calc.BirthHourUnknown {
		return fmt.Sprintf(
			"简化干支示意为 %s，时辰未知，本次不生成时柱，日主为「%s」。%s",
			pillarText,
			calc.DayMaster,
			calc.BaziProfile.DayMasterObservation,
		)
	}
	return fmt.Sprintf(
		"简化干支示意为 %s，日主为「%s」。%s",
		pillarText,
		calc.DayMaster,
		calc.BaziProfile.DayMasterObservation,
	)
}

func buildElementLearning(calc CalculationResult) string {
	e := calc.FiveElements
	base := fmt.Sprintf(
		"五行计数（学习参考）：木 %d、火 %d、土 %d、金 %d、水 %d。整体呈%s。",
		e.Wood, e.Fire, e.Earth, e.Metal, e.Water,
		calc.BaziProfile.ElementBalanceType,
	)
	return base + calc.InterpretationLens.RelationshipWithElements
}

func buildReflectionSection(calc CalculationResult) string {
	lines := []string{calc.ReflectionFocus}
	lines = append(lines, calc.ActionSuggestions...)
	return strings.Join(lines, "\n")
}
