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
	focus := pickProfessionalFocusPalaces(palaces, chief, category)
	sections := []string{
		v2SectionSummary + "\n" + buildQimenProfessionalSummarySection(parsed, profile, lens, category, categoryText, methodNote, disclaimer, freeContent),
		v2SectionBasis + "\n" + buildQimenProfessionalBasisSection(parsed),
		v2SectionPalaces + "\n" + buildQimenProfessionalPalacesOverviewSection(palaces),
		v2SectionFocusPalaces + "\n" + buildProfessionalPalaceReportSections(focus, chief, dun, category),
		v2SectionSupport + "\n" + buildQimenProfessionalSupportSection(profile, lens, focus, category),
		v2SectionRisks + "\n" + buildQimenProfessionalRiskSection(parsed.RiskObservations, profile, lens, focus, category),
		v2SectionPacing + "\n" + buildQimenProfessionalPacingSection(parsed.ActionPacing, profile, lens, category),
		v2SectionReflection + "\n" + buildQimenReflectionSection(parsed.ReflectionQuestions, profile, category),
		v2SectionBoundary + "\n" + buildQimenProfessionalBoundarySection(methodNote, parsed),
	}
	return strings.Join(sections, "\n\n")
}

// pickProfessionalFocusPalaces selects 2–3 stable focus palaces for professional reports.
// Category only affects report focus, not layout calculation.
func pickProfessionalFocusPalaces(palaces []Palace, chief Chief, category string) []Palace {
	if len(palaces) == 0 {
		return nil
	}

	picked := make([]Palace, 0, 3)
	add := func(p Palace) {
		if strings.TrimSpace(p.Name) == "" {
			return
		}
		for _, existing := range picked {
			if existing.Name == p.Name {
				return
			}
		}
		picked = append(picked, p)
	}

	if name := strings.TrimSpace(chief.ZhiFuPalace); name != "" {
		if p, ok := findPalaceByName(palaces, name); ok {
			add(p)
		}
	} else if p, ok := findPalaceByStar(palaces, chief.ZhiFu); ok {
		add(p)
	}

	if name := strings.TrimSpace(chief.ZhiShiPalace); name != "" {
		if p, ok := findPalaceByName(palaces, name); ok {
			add(p)
		}
	} else if p, ok := findPalaceByDoor(palaces, chief.ZhiShi); ok {
		add(p)
	}

	for _, name := range categoryFocusPalaceNames(category) {
		if p, ok := findPalaceByName(palaces, name); ok {
			add(p)
		}
		if len(picked) >= 3 {
			break
		}
	}

	if len(picked) < 2 {
		for _, p := range palaces {
			add(p)
			if len(picked) >= 2 {
				break
			}
		}
	}
	if len(picked) > 3 {
		picked = picked[:3]
	}
	if len(picked) == 0 {
		picked = append(picked, palaces[0])
	}
	return picked
}

func summarizeProfessionalFocusPalaces(focus []Palace) string {
	if len(focus) == 0 {
		return "（无）"
	}
	lines := make([]string, 0, len(focus))
	for _, p := range focus {
		lines = append(lines, formatProfessionalPalaceDetail(p))
	}
	return strings.Join(lines, "\n")
}

func formatProfessionalPalaceDetail(p Palace) string {
	door := fallbackString(p.Door, "—")
	deity := fallbackString(p.Deity, "—")
	summary := strings.TrimSpace(p.Summary)
	if summary == "" {
		summary = "可用于结构化观察，建议结合现实情况判断。"
	}
	return fmt.Sprintf(
		"%s：星=%s，门=%s，神=%s，天盘干=%s，地盘干=%s；%s",
		p.Name, p.Star, door, deity, p.HeavenPlateStem, p.EarthPlateStem, summary,
	)
}

func buildProfessionalPalaceReportSections(focus []Palace, chief Chief, dun Dun, category string) string {
	if len(focus) == 0 {
		return "暂无可引用的重点宫位，请先回到局势摘要整理问题。"
	}

	dunTypeLabel := "—"
	switch dun.Type {
	case "yang":
		dunTypeLabel = "阳遁"
	case "yin":
		dunTypeLabel = "阴遁"
	case "":
		dunTypeLabel = "—"
	default:
		dunTypeLabel = dun.Type
	}
	yuanLabel := professionalYuanLabel(dun)

	lines := []string{
		fmt.Sprintf(
			"值符：%s（%s）；值使：%s（%s）；%s第%d局·%s。",
			fallbackString(chief.ZhiFu, "—"),
			fallbackString(chief.ZhiFuPalace, "—"),
			fallbackString(chief.ZhiShi, "—"),
			fallbackString(chief.ZhiShiPalace, "—"),
			dunTypeLabel,
			dun.Ju,
			yuanLabel,
		),
		professionalCategoryFocusIntro(category),
		"以下 2–3 个宫位为可优先关注的结构化观察点，不作强预测：",
	}
	for i, p := range focus {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, formatProfessionalPalaceDetail(p)))
		lines = append(lines, "   "+professionalPalaceObservationHint(category, p))
	}
	return strings.Join(lines, "\n")
}

func professionalYuanLabel(dun Dun) string {
	switch strings.TrimSpace(dun.Yuan) {
	case "upper":
		return "上元"
	case "middle":
		return "中元"
	case "lower":
		return "下元"
	default:
		if y := strings.TrimSpace(dun.Yuan); y != "" {
			return y
		}
		return "—"
	}
}

func professionalCategoryFocusIntro(category string) string {
	switch NormalizeCategory(category) {
	case "career":
		return "就事业/计划类问事，可优先关注推进顺序、资源协调与执行阻力，建议先小范围验证："
	case "relationship":
		return "就人际/关系类问事，可优先关注沟通边界、情绪节奏与误解修复，不急于定性："
	case "study":
		return "就学习/成长类问事，可优先关注复盘节奏、专注力与阶段目标，先搭框架再补细节："
	case "decision":
		return "就决策/选择类问事，可优先关注信息补齐、小步试探与备用方案，避免一次性押注："
	default:
		return "就综合问事，可优先关注问题整理、风险收敛与行动节奏，观察反馈后再调整："
	}
}

func professionalPalaceObservationHint(category string, p Palace) string {
	base := categoryPalaceObservationHint(category, p)
	return "结构化观察：" + base + " 建议结合现实情况判断，不构成现实决策依据。"
}

func buildQimenProfessionalSummarySection(
	parsed parsedResultPayload,
	profile QuestionProfile,
	lens QimenLens,
	category, categoryText, methodNote, disclaimer, freeContent string,
) string {
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
	lines := []string{
		disclaimer,
		methodNote,
		fmt.Sprintf("问事分类：%s。%s", categoryText, timeNote),
		fmt.Sprintf("问事特征：%s", safeSummary),
		fmt.Sprintf("问事摘要：%s", QuestionSummary),
		parsed.SituationOverview,
		professionalCategorySummaryHint(category, profile),
		fmt.Sprintf("关注主题：%s；行动节奏整理：%s。", lens.FocusTheme, lens.PacingTheme),
	}
	if strings.TrimSpace(freeContent) != "" {
		lines = append(lines, "可参考免费解读中的局势梳理要点，继续延伸记录与复盘。")
	}
	return strings.Join(lines, "\n")
}

func professionalCategorySummaryHint(category string, profile QuestionProfile) string {
	switch NormalizeCategory(category) {
	case "career":
		return fmt.Sprintf("就事业/计划而言，当前更适合先整理推进顺序与资源协调，围绕「%s」安排一件可先验证的小动作。", profile.IntentType)
	case "relationship":
		return fmt.Sprintf("就人际/关系而言，当前更适合先梳理与「%s」相关的沟通边界与情绪节奏，不急于定性。", profile.IntentType)
	case "study":
		return "就学习/成长而言，当前更适合先复盘方法与专注状态，再设定可完成的阶段目标。"
	case "decision":
		return "就决策/选择而言，当前更适合先补齐信息与选项约束，再小步试探并准备备用方案。"
	default:
		return fmt.Sprintf("就综合问题而言，当前更适合先整理问题与风险，再安排与「%s」相关的小步行动并观察反馈。", profile.IntentType)
	}
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
		lines = append(lines, fmt.Sprintf(
			"阴阳遁：%s；局数：%d；三元：%s（拆补/节气口径：%s）。",
			dun.Type, dun.Ju, professionalYuanLabel(dun), method,
		))
	}
	if xun != nil && xun.XunShou != "" {
		lines = append(lines, fmt.Sprintf("旬首：%s；空亡：%s。", xun.XunShou, strings.Join(xun.EmptyBranches, "、")))
	}
	lines = append(lines,
		"当前为第一版排盘口径：默认天禽留中五宫；置闰法、坤二/艮八寄宫流派校准仍未完成。",
		"星、门、神、天盘干、地盘干为第一版 professional 落盘，仅供结构化观察，不作强预测。",
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
			"· %s：星=%s，门=%s，神=%s，天盘干=%s，地盘干=%s",
			p.Name, p.Star, fallbackString(p.Door, "—"), fallbackString(p.Deity, "—"), p.HeavenPlateStem, p.EarthPlateStem,
		))
	}
	lines = append(lines, "以上九宫结构用于传统文化学习中的结构化观察，不作吉凶强断。")
	return strings.Join(lines, "\n")
}

func buildQimenProfessionalSupportSection(profile QuestionProfile, lens QimenLens, focus []Palace, category string) string {
	lines := []string{
		lens.SupportTheme + "。",
		fmt.Sprintf("可借助「%s」相关已有资源，先明确最小可验证一步。", profile.IntentType),
	}
	if len(focus) > 0 {
		p := focus[0]
		lines = append(lines, fmt.Sprintf(
			"从 %s（星=%s，门=%s，天盘干=%s，地盘干=%s）出发，可以先验证一件今天能完成的小事。",
			p.Name, p.Star, fallbackString(p.Door, "—"), p.HeavenPlateStem, p.EarthPlateStem,
		))
	}
	switch NormalizeCategory(category) {
	case "career":
		lines = append(lines, "把目标拆成「本周一件可完成的小事」，先验证推进顺序是否合理，再决定是否加码。")
	case "relationship":
		lines = append(lines, "先记录一次互动感受，再决定是否调整沟通方式或边界，不急于定性。")
	case "study":
		lines = append(lines, "设定一段 25–40 分钟的专注块，先搭框架再补细节，并记录收获与卡点。")
	case "decision":
		lines = append(lines, "列出选项、约束与可接受的最坏情况，再选低成本试探并保留备用方案。")
	default:
		lines = append(lines, "安排一件今天能完成的小事，建立可控感并观察反馈。")
	}
	return strings.Join(lines, "\n")
}

func buildQimenProfessionalRiskSection(risks []string, profile QuestionProfile, lens QimenLens, focus []Palace, category string) string {
	items := append([]string{lens.CautionTheme + "。"}, risks...)
	switch NormalizeCategory(category) {
	case "career":
		items = append(items, "执行分散、节奏过快时，可能一次承担过多线程，宜先收敛再推进。")
	case "relationship":
		items = append(items, "把一次互动放大成整体判断，可能增加误解，宜先观察情绪节奏。")
	case "study":
		items = append(items, "只比较结果而忽略方法，可能让学习节奏失衡，宜先复盘再调整。")
	case "decision":
		items = append(items, "在选项未写清前急于选择，可能遗漏关键约束，宜先补齐信息。")
	default:
		items = append(items, "把第一版九宫结构当作确定答案，可能偏离自我反思初衷。")
	}
	if len(focus) > 1 {
		p := focus[1]
		items = append(items, fmt.Sprintf(
			"关注 %s（星=%s，门=%s）时，宜先观察再行动，避免过度解读。",
			p.Name, p.Star, fallbackString(p.Door, "—"),
		))
	}
	if profile.DecisionPressure == "高" {
		items = append(items, "犹豫与担心反复切换，可能让行动迟迟无法启动，可先安排最小验证步。")
	}
	if len(items) > 5 {
		items = items[:5]
	}
	return strings.Join(items, "\n")
}

func buildQimenProfessionalPacingSection(pacing string, profile QuestionProfile, lens QimenLens, category string) string {
	if strings.TrimSpace(pacing) == "" {
		pacing = lens.PacingTheme + "：先观察再行动，建议结合现实情况判断。"
	}
	return pacing + "\n\n" + professionalCategoryPacingHint(category, profile)
}

func professionalCategoryPacingHint(category string, profile QuestionProfile) string {
	switch NormalizeCategory(category) {
	case "career":
		return fmt.Sprintf("行动节奏整理：先整理现状与目标，再列出与「%s」相关的一件小动作，最后定期复盘推进顺序。", profile.IntentType)
	case "relationship":
		return fmt.Sprintf("行动节奏整理：先记录感受，再明确一次与「%s」相关的沟通边界，最后观察变化后再调整。", profile.IntentType)
	case "study":
		return "行动节奏整理：先复盘状态，再设定可完成的学习块，最后记录收获并调整阶段目标。"
	case "decision":
		return "行动节奏整理：先写下选项与约束，再选一个小步验证，最后比较反馈并更新备用方案。"
	default:
		return fmt.Sprintf("行动节奏整理：先观察，再安排与「%s」相关的一件小事，并根据反馈收敛风险。", profile.IntentType)
	}
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
