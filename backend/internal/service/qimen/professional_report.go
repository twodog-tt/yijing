package qimen

import (
	"fmt"
	"strings"
)

const (
	fullReportDisclaimerProfessional = "本报告基于 qimen-v2-professional 第一版九宫落盘生成，仅用于传统文化学习与结构化观察，不等同于最终权威专业排盘，不构成现实决策依据。"
	freeDisclaimerProfessional       = "本解读基于 qimen-v2-professional 第一版专业口径，仅供传统文化学习与结构化观察，置闰法、寄宫流派校准仍未完成，不构成现实决策依据。"
)

func buildQimenProfessionalFallbackFullContent(
	parsed parsedResultPayload,
	profile QuestionProfile,
	lens QimenLens,
	category, categoryText, methodNote, disclaimer string,
	freeContent string,
) string {
	palaces := parsed.Palaces
	chief := chiefFromPayload(parsed.Chief)
	dun := dunFromPayload(parsed.Dun)
	focus := pickQimenV2FocusPalaces(palaces, chief, category)
	sections := []string{
		v2SectionSummary + "\n" + buildQimenV2SummarySection(parsed, profile, lens, category, categoryText, methodNote, disclaimer, freeContent),
		v2SectionBasis + "\n" + buildQimenProfessionalBasisSection(parsed),
		v2SectionPalaces + "\n" + buildQimenProfessionalPalacesOverviewSection(palaces),
		v2SectionFocusPalaces + "\n" + buildQimenV2FocusPalacesSection(focus, chief, dun, category),
		v2SectionSupport + "\n" + buildQimenV2SupportSection(profile, lens, focus, category),
		v2SectionRisks + "\n" + buildQimenV2RiskSection(parsed.RiskObservations, profile, lens, focus, category),
		v2SectionPacing + "\n" + buildQimenV2PacingSection(parsed.ActionPacing, profile, lens, category),
		v2SectionReflection + "\n" + buildQimenReflectionSection(parsed.ReflectionQuestions, profile, category),
		v2SectionBoundary + "\n" + buildQimenProfessionalBoundarySection(methodNote, parsed),
	}
	return strings.Join(sections, "\n\n")
}

func buildQimenProfessionalBasisSection(parsed parsedResultPayload) string {
	basis := calendarFromPayload(parsed.CalendarBasis)
	dun := dunFromPayload(parsed.Dun)
	xun := parsed.Xun
	layoutVersion := strings.TrimSpace(parsed.LayoutVersion)
	if layoutVersion == "" {
		layoutVersion = ProfessionalLayoutVersionV1
	}
	lines := []string{
		"algorithm_version：qimen-v2-professional（第一版落盘口径，非最终权威排盘）。",
		fmt.Sprintf("layout_version：%s。", layoutVersion),
		fmt.Sprintf("节令参考：%s（口径：%s，时间基准：%s）。", fallbackString(basis.SolarTerm, "未指定"), fallbackString(basis.JieqiBasis, "formula_approximation"), fallbackString(basis.TimeBasis, "local_time")),
	}
	if strings.TrimSpace(basis.Note) != "" {
		lines = append(lines, basis.Note)
	}
	if parsed.Ganzhi != nil {
		gz := parsed.Ganzhi
		lines = append(lines, fmt.Sprintf("四柱：年=%s，月=%s，日=%s，时=%s。", gz.Year, gz.Month, gz.Day, gz.Hour))
	}
	if dun.Type != "" || dun.Ju > 0 {
		method := dun.Source
		if method == "" {
			method = DunMethodChaiBu
		}
		lines = append(lines, fmt.Sprintf("阴阳遁：%s；局数：%d（拆补/节气口径：%s）。", dun.Type, dun.Ju, method))
	}
	if xun != nil && xun.XunShou != "" {
		lines = append(lines, fmt.Sprintf("旬首：%s；空亡：%s。", xun.XunShou, strings.Join(xun.EmptyBranches, "、")))
	}
	lines = append(lines,
		"当前默认天禽留中五宫；置闰法、坤二/艮八寄宫流派校准仍未完成。",
		"星、门、神、天盘干、地盘干为第一版 professional 落盘，仅供结构化观察。",
	)
	return strings.Join(lines, "\n")
}

func buildQimenProfessionalPalacesOverviewSection(palaces []Palace) string {
	if len(palaces) == 0 {
		return "当前记录未包含九宫结构，无法展开宫位观察。"
	}
	lines := []string{
		fmt.Sprintf("本次 professional 九宫共 %d 宫，以下为压缩摘要（非原始 JSON）：", len(palaces)),
	}
	for _, p := range palaces {
		lines = append(lines, fmt.Sprintf(
			"· %s：%s / %s / %s（天盘%s，地盘%s）",
			p.Name, p.Star, p.Door, p.Deity, p.HeavenPlateStem, p.EarthPlateStem,
		))
	}
	lines = append(lines, "以上结构用于学习观察，不作吉凶强断。")
	return strings.Join(lines, "\n")
}

func buildQimenProfessionalBoundarySection(methodNote string, parsed parsedResultPayload) string {
	limits := calculationLimits
	if parsed.CalculationMeta != nil && len(parsed.CalculationMeta.Limits) > 0 {
		limits = parsed.CalculationMeta.Limits
	}
	lines := []string{
		fullReportDisclaimerProfessional,
		methodNote,
		"当前仍为 professional 第一版，不等同于最终权威排盘；置闰法、寄宫流派校准仍未完成；不构成现实决策依据。",
		"本报告不构成现实决策依据，不做精准预测、强吉凶判断、改运化解，也不提供投资/医疗/法律/赌博/军事建议。",
		"规则限制：" + strings.Join(limits, "；"),
	}
	if len(parsed.Palaces) > 0 {
		lines = append(lines, fmt.Sprintf("九宫落盘 layout_version=%s（共 %d 宫），仅供结构化观察。",
			fallbackString(parsed.LayoutVersion, ProfessionalLayoutVersionV1), len(parsed.Palaces)))
	}
	return strings.Join(lines, "\n")
}
