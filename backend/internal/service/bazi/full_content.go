package bazi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
)

const fullReportDisclaimer = "本报告基于 bazi-simple-v1 简化干支文化规则生成，仅用于传统文化学习与自我反思，不等同于专业八字排盘，不构成现实决策依据。"

const (
	sectionBrief         = "【一、简要说明】"
	sectionPillars       = "【二、四柱与五行观察】"
	sectionTendency      = "【三、个人倾向与行动风格】"
	sectionPacing        = "【四、需要留意的节奏】"
	sectionReflection    = "【五、适合的自我反思问题】"
	sectionActions       = "【六、近期行动建议】"
	sectionBoundary      = "【七、边界声明】"
)

// BuildFullContent generates a structured template full report from stored analysis payloads.
func BuildFullContent(resultPayload json.RawMessage, freeContent string) (string, error) {
	parsed, err := parseStoredResultPayload(resultPayload)
	if err != nil {
		return "", err
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

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
