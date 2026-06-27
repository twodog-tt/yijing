package bazi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
)

const fullReportDisclaimer = "本报告基于 bazi-simple-v1 简化干支文化规则生成，仅用于传统文化学习与自我反思，不等同于专业八字排盘，不构成现实决策依据。"

const (
	sectionBrief      = "【一、简要说明】"
	sectionPillars    = "【二、四柱与五行观察】"
	sectionTendency   = "【三、个人倾向与行动风格】"
	sectionPacing     = "【四、需要留意的节奏】"
	sectionReflection = "【五、适合的自我反思问题】"
	sectionActions    = "【六、近期行动建议】"
	sectionBoundary   = "【七、边界声明】"
)

const (
	sectionV2Overall  = "【一、整体结构摘要】"
	sectionV2Basis    = "【二、排盘口径说明】"
	sectionV2Pillars  = "【三、四柱结构观察】"
	sectionV2Elements = "【四、五行分布观察】"
	sectionV2Strength = "【五、可借助的倾向】"
	sectionV2Caution  = "【六、需要留意的倾向】"
	sectionV2Pacing   = "【七、行动节奏建议】"
	sectionV2Boundary = "【八、边界声明】"
)

// BuildFullContent generates a structured template full report from stored analysis payloads.
func BuildFullContent(resultPayload json.RawMessage, freeContent string) (string, error) {
	parsed, err := parseStoredResultPayload(resultPayload)
	if err != nil {
		return "", err
	}
	if parsed.AlgorithmVersion == AlgorithmVersionBaziV2POC {
		return buildBaziV2FullContent(parsed, freeContent), nil
	}

	hourUnknown := strings.TrimSpace(parsed.Pillars.Hour) == ""
	profile := profileFromPayload(parsed.BaziProfile, parsed.DayMaster, parsed.FiveElements, hourUnknown)
	lens := lensFromPayload(parsed.InterpretationLens, profile, parsed.DayMaster, parsed.FiveElements, hourUnknown)

	methodNote := strings.TrimSpace(parsed.MethodNote)
	if methodNote == "" {
		if parsed.AlgorithmVersion == AlgorithmVersionBaziV2POC {
			methodNote = MethodNoteV2
		} else {
			methodNote = MethodNote
		}
	}
	disclaimer := fullReportDisclaimerFor(parsed.AlgorithmVersion)

	sections := []string{
		sectionBrief + "\n" + buildBaziBriefSection(parsed, profile, methodNote, disclaimer),
		sectionPillars + "\n" + buildBaziPillarsSection(parsed, profile, lens, hourUnknown),
		sectionTendency + "\n" + buildBaziTendencySection(profile, parsed.DayMaster),
		sectionPacing + "\n" + buildBaziPacingSection(profile, lens, hourUnknown, parsed.AlgorithmVersion),
		sectionReflection + "\n" + buildBaziFullReflectionQuestions(profile, lens, hourUnknown),
		sectionActions + "\n" + buildBaziFullActionSuggestions(parsed.ActionSuggestions, profile, lens, parsed.DayMaster),
		sectionBoundary + "\n" + buildBaziBoundarySection(disclaimer, methodNote, limitsFromStored(parsed)),
	}

	if snippet := strings.TrimSpace(freeContent); snippet != "" {
		sections[5] = sectionActions + "\n" + buildBaziFullActionSuggestions(parsed.ActionSuggestions, profile, lens, parsed.DayMaster) +
			"\n\n可参考免费解读中的要点继续延伸记录，但不把简化示意当作确定结论。"
	}

	return strings.Join(sections, "\n\n"), nil
}

func buildBaziV2FullContent(parsed storedResultPayload, freeContent string) string {
	hourUnknown := strings.TrimSpace(parsed.Pillars.Hour) == ""
	profile := profileFromPayload(parsed.BaziProfile, parsed.DayMaster, parsed.FiveElements, hourUnknown)
	lens := lensFromPayload(parsed.InterpretationLens, profile, parsed.DayMaster, parsed.FiveElements, hourUnknown)
	methodNote := strings.TrimSpace(parsed.MethodNote)
	if methodNote == "" {
		methodNote = MethodNoteV2
	}
	disclaimer := fullReportDisclaimerFor(parsed.AlgorithmVersion)

	sections := []string{
		sectionV2Overall + "\n" + buildBaziV2OverallSection(parsed, profile, methodNote),
		sectionV2Basis + "\n" + buildBaziV2BasisSection(parsed, methodNote),
		sectionV2Pillars + "\n" + buildBaziV2PillarsSection(parsed, profile, hourUnknown),
		sectionV2Elements + "\n" + buildBaziV2ElementsSection(parsed, profile, lens),
		sectionV2Strength + "\n" + buildBaziV2StrengthSection(profile, lens, parsed.DayMaster),
		sectionV2Caution + "\n" + buildBaziV2CautionSection(profile, lens, hourUnknown),
		sectionV2Pacing + "\n" + buildBaziV2PacingSection(parsed, profile, lens, freeContent),
		sectionV2Boundary + "\n" + buildBaziBoundarySection(disclaimer, methodNote, limitsFromStored(parsed)),
	}
	return strings.Join(sections, "\n\n")
}

func buildBaziV2OverallSection(parsed storedResultPayload, profile BaziProfile, methodNote string) string {
	return strings.Join([]string{
		"本报告基于八字 v2 / bazi-v2-poc 的结构化结果生成，用于传统文化学习、自我观察与行动节奏整理。",
		fmt.Sprintf("algorithm_version：%s", AlgorithmVersionBaziV2POC),
		methodNote,
		fmt.Sprintf("本次结构摘要：日主「%s」，五行倾向「%s」，行动风格「%s」，反思主题「%s」。",
			parsed.DayMaster, profile.ElementBalanceType, profile.ActionStyle, profile.ReflectionTheme),
		"以下内容只讨论观察角度与可执行节奏，不作为现实决策依据。",
	}, "\n")
}

func buildBaziV2BasisSection(parsed storedResultPayload, methodNote string) string {
	lines := []string{
		"排盘口径：" + formatCalendarBasisForReport(parsed.CalendarBasis),
		"年柱按立春换年，月柱按十二节令切换月令；节令时刻为公式近似，非天文台精确时刻。",
		"真太阳时：未实现；日柱口径：" + nonEmpty(parsed.CalendarBasis.DayPillarBasis, "fixed_epoch_v1") + "。",
	}
	if methodNote != "" {
		lines = append(lines, "方法说明："+methodNote)
	}
	return strings.Join(lines, "\n")
}

func buildBaziV2PillarsSection(parsed storedResultPayload, profile BaziProfile, hourUnknown bool) string {
	lines := []string{
		fmt.Sprintf("年柱：%s（立春换年口径）", nonEmpty(parsed.Pillars.Year, "—")),
		fmt.Sprintf("月柱：%s（十二节令月柱口径）", nonEmpty(parsed.Pillars.Month, "—")),
		fmt.Sprintf("日柱：%s（日主为「%s」）", nonEmpty(parsed.Pillars.Day, "—"), nonEmpty(parsed.DayMaster, "—")),
	}
	if hourUnknown {
		lines = append(lines, "时柱：时辰未知，本次未生成时柱，也不做时柱相关展开。")
	} else {
		lines = append(lines, fmt.Sprintf("时柱：%s（仅作结构观察，不作强判断）", nonEmpty(parsed.Pillars.Hour, "—")))
	}
	lines = append(lines, profile.DayMasterObservation)
	return strings.Join(lines, "\n")
}

func buildBaziV2ElementsSection(parsed storedResultPayload, profile BaziProfile, lens InterpretationLens) string {
	e := parsed.FiveElements
	dominant, weak := dominantElements(e)
	lines := []string{
		fmt.Sprintf("五行结构：木 %d、火 %d、土 %d、金 %d、水 %d。", e.Wood, e.Fire, e.Earth, e.Metal, e.Water),
		fmt.Sprintf("整体观察：%s。", profile.ElementBalanceType),
		elementObservationFor("木", e.Wood),
		elementObservationFor("火", e.Fire),
		elementObservationFor("土", e.Earth),
		elementObservationFor("金", e.Metal),
		elementObservationFor("水", e.Water),
		lens.RelationshipWithElements,
	}
	if dominant != "" {
		lines = append(lines, fmt.Sprintf("相对突出：%s；相对偏少：%s。这里的多少只作观察维度，不代表好坏。", dominant, nonEmpty(weak, "无明显偏少")))
	}
	return strings.Join(lines, "\n")
}

func buildBaziV2StrengthSection(profile BaziProfile, lens InterpretationLens, dayMaster string) string {
	return strings.Join([]string{
		fmt.Sprintf("可借助的主题：%s。", profile.ReflectionTheme),
		fmt.Sprintf("可借助的行动风格：%s。", profile.ActionStyle),
		lens.StrengthHint + "。",
		fmt.Sprintf("从%s元素相关的「%s」入手，更适合作为自我观察入口，而不是结论。", stemElementName(dayMaster), profile.ReflectionTheme),
	}, "\n")
}

func buildBaziV2CautionSection(profile BaziProfile, lens InterpretationLens, hourUnknown bool) string {
	lines := []string{
		lens.CautionHint + "。",
		"需要留意把「偏多 / 偏少」理解成确定好坏；更适合记录哪些情境让自己更稳定或更消耗。",
	}
	if hourUnknown {
		lines = append(lines, "时辰未知时，先不展开时柱细节，可把观察重点放在年月日三柱与五行结构。")
	}
	return strings.Join(lines, "\n")
}

func buildBaziV2PacingSection(parsed storedResultPayload, profile BaziProfile, lens InterpretationLens, freeContent string) string {
	suggestions := buildBaziFullActionSuggestions(parsed.ActionSuggestions, profile, lens, parsed.DayMaster)
	lines := []string{
		"节奏建议：" + lens.PacingHint + "。",
		"近期可以选择一件低成本、可复盘的小事验证观察，而不是急于给自己下定义。",
		"行动参考：\n" + suggestions,
	}
	if strings.TrimSpace(freeContent) != "" {
		lines = append(lines, "免费解读可作为补充摘要，但本报告仍以结构化字段为主。")
	}
	return strings.Join(lines, "\n")
}

func buildBaziBriefSection(parsed storedResultPayload, profile BaziProfile, methodNote, disclaimer string) string {
	lines := []string{
		disclaimer,
		methodNote,
		fmt.Sprintf("algorithm_version：%s", nonEmpty(parsed.AlgorithmVersion, model.AlgorithmVersionBaziSimpleV1)),
		fmt.Sprintf("本次简析侧重：五行倾向「%s」、行动风格「%s」、反思主题「%s」。",
			profile.ElementBalanceType, profile.ActionStyle, profile.ReflectionTheme),
	}
	if parsed.AlgorithmVersion == AlgorithmVersionBaziV2POC {
		lines = append(lines, "年柱按立春换年、月柱按十二节令切换；节令时刻为公式近似，非天文台精确时刻。")
	}
	return strings.Join(lines, "\n")
}

func buildBaziPillarsSection(parsed storedResultPayload, profile BaziProfile, lens InterpretationLens, hourUnknown bool) string {
	pillarParts := []string{
		fmt.Sprintf("年柱：%s", nonEmpty(parsed.Pillars.Year, "—")),
		fmt.Sprintf("月柱：%s", nonEmpty(parsed.Pillars.Month, "—")),
		fmt.Sprintf("日柱：%s", nonEmpty(parsed.Pillars.Day, "—")),
	}
	if hourUnknown {
		pillarParts = append(pillarParts, "时柱：时辰未知，本次不生成时柱，也不做时柱相关推断。")
	} else {
		pillarParts = append(pillarParts, fmt.Sprintf("时柱：%s", parsed.Pillars.Hour))
	}

	e := parsed.FiveElements
	elementBlock := fmt.Sprintf(
		"五行计数（学习参考）：木 %d、火 %d、土 %d、金 %d、水 %d。\n整体呈「%s」。%s",
		e.Wood, e.Fire, e.Earth, e.Metal, e.Water,
		profile.ElementBalanceType,
		lens.RelationshipWithElements,
	)
	observation := fmt.Sprintf(
		"日主「%s」：%s\n季节倾向（简化）：%s。",
		parsed.DayMaster,
		profile.DayMasterObservation,
		profile.SeasonTendency,
	)
	return strings.Join(pillarParts, "\n") + "\n\n" + elementBlock + "\n\n" + observation
}

func buildBaziTendencySection(profile BaziProfile, dayMaster string) string {
	return fmt.Sprintf(
		"行动风格偏向「%s」，适合从「%s」相关情境入手做自我观察。\n"+
			"反思主题可关注「%s」：当遇到与%s元素相关的节奏变化时，可先记录感受，再安排小动作验证。\n"+
			"可借助：%s",
		profile.ActionStyle,
		profile.ReflectionTheme,
		profile.ReflectionTheme,
		stemElementName(dayMaster),
		strengthHintFor(profile, dayMaster, ""),
	)
}

func buildBaziPacingSection(profile BaziProfile, lens InterpretationLens, hourUnknown bool, algorithmVersion string) string {
	lines := []string{
		fmt.Sprintf("节奏建议：%s", lens.PacingHint),
		fmt.Sprintf("需留意：%s", lens.CautionHint),
		fmt.Sprintf("与五行相关的观察角度：%s", strings.TrimSuffix(lens.RelationshipWithElements, "。")),
	}
	if hourUnknown {
		lines = append([]string{
			"时辰未知：暂缓对时柱相关细节的推断，先从年月日三柱与五行倾向做整理。",
		}, lines...)
	}
	if algorithmVersion == AlgorithmVersionBaziV2POC {
		lines = append(lines, "v2 口径下月令按十二节令切换，边界日前后宜多观察、少定论。")
	}
	return strings.Join(lines, "\n")
}

func buildBaziFullReflectionQuestions(profile BaziProfile, lens InterpretationLens, hourUnknown bool) string {
	questions := []string{
		fmt.Sprintf("围绕「%s」，最近哪些情境让我更容易进入稳定状态？", profile.ReflectionTheme),
		fmt.Sprintf("当五行倾向呈「%s」时，我在哪些场景下更容易过度投入或过度保守？", profile.ElementBalanceType),
		fmt.Sprintf("如果按「%s」的方式推进，本周最小可验证的一步是什么？", profile.ActionStyle),
		lens.CautionHint + "——我可以如何提前做一点缓冲？",
	}
	if hourUnknown {
		questions = append(questions, "在缺少时柱信息时，我仍可以从哪些已知信息出发做自我观察？")
	}
	return strings.Join(questions, "\n")
}

func buildBaziFullActionSuggestions(stored []string, profile BaziProfile, lens InterpretationLens, dayMaster string) string {
	suggestions := append([]string{}, stored...)
	if len(suggestions) == 0 {
		suggestions = BuildActionSuggestions(profile, lens, dayMaster, FiveElements{}, false)
	}
	suggestions = append(suggestions,
		fmt.Sprintf("结合「%s」主题，安排一件本周可完成的小事并记录感受。", profile.ReflectionTheme),
		lens.StrengthHint+"，可作为本周观察切入点。",
	)
	return strings.Join(dedupeLimit(suggestions, 5), "\n")
}

func limitsFromStored(parsed storedResultPayload) []string {
	if parsed.CalculationMeta != nil && len(parsed.CalculationMeta.Limits) > 0 {
		return parsed.CalculationMeta.Limits
	}
	return nil
}

func buildBaziBoundarySection(disclaimer, methodNote string, limits []string) string {
	lines := []string{
		disclaimer,
		methodNote,
		"本报告不构成现实决策依据，不做精准预测、强吉凶判断、改运化解，也不提供投资/医疗/法律/赌博/军事建议。",
	}
	if len(limits) > 0 {
		lines = append(lines, "规则限制："+strings.Join(limits, "；"))
	}
	return strings.Join(lines, "\n")
}

func formatCalendarBasisForReport(basis CalendarBasis) string {
	parts := []string{
		"year_boundary=" + nonEmpty(basis.YearBoundary, "—"),
		"month_boundary=" + nonEmpty(basis.MonthBoundary, "—"),
		"true_solar_time=" + fmt.Sprintf("%t", basis.TrueSolarTime),
		"day_pillar_basis=" + nonEmpty(basis.DayPillarBasis, "—"),
	}
	if note := strings.TrimSpace(basis.Note); note != "" {
		parts = append(parts, "note="+note)
	}
	return strings.Join(parts, "；")
}

func elementObservationFor(name string, count int) string {
	quality := map[string]string{
		"木": "生发、规划、学习、扩展",
		"火": "表达、推动、显化、热情",
		"土": "承载、稳定、整合、节奏",
		"金": "规则、边界、判断、收敛",
		"水": "流动、信息、适应、反思",
	}[name]
	level := "可观察"
	if count >= 3 {
		level = "相对突出"
	} else if count <= 0 {
		level = "相对偏少"
	}
	return fmt.Sprintf("%s：%s，计数 %d，可从「%s」角度做轻量观察。", name, level, count, quality)
}

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
