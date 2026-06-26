package qimen

import (
	"fmt"
	"strings"
)

const (
	v2SectionSummary       = "【一、局势摘要】"
	v2SectionBasis         = "【二、排盘口径说明】"
	v2SectionPalaces       = "【三、九宫结构观察】"
	v2SectionFocusPalaces  = "【四、重点宫位提示】"
	v2SectionSupport       = "【五、可借助的条件】"
	v2SectionRisks         = "【六、需要留意的阻力】"
	v2SectionPacing        = "【七、行动节奏建议】"
	v2SectionReflection    = "【八、自我反思问题】"
	v2SectionBoundary      = "【九、边界声明】"
)

// summarizeQimenV2Palaces compresses palace fields for prompt or report text.
func summarizeQimenV2Palaces(palaces []Palace) string {
	if len(palaces) == 0 {
		return "（无）"
	}
	lines := make([]string, 0, len(palaces))
	for _, p := range palaces {
		lines = append(lines, fmt.Sprintf(
			"%s：星=%s，门=%s，神=%s，天盘=%s，地盘=%s；%s",
			p.Name, p.Star, p.Door, p.Deity, p.HeavenPlateStem, p.EarthPlateStem, p.Summary,
		))
	}
	return strings.Join(lines, "\n")
}

func categoryFocusPalaceNames(category string) []string {
	switch NormalizeCategory(category) {
	case "career":
		return []string{"乾六宫", "离九宫"}
	case "relationship":
		return []string{"兑七宫", "巽四宫"}
	case "study":
		return []string{"坤二宫", "离九宫"}
	case "decision":
		return []string{"艮八宫", "坎一宫"}
	default:
		return []string{"中五宫", "坤二宫"}
	}
}

func findPalaceByStar(palaces []Palace, star string) (Palace, bool) {
	star = strings.TrimSpace(star)
	if star == "" {
		return Palace{}, false
	}
	for _, p := range palaces {
		if p.Star == star {
			return p, true
		}
	}
	return Palace{}, false
}

func findPalaceByDoor(palaces []Palace, door string) (Palace, bool) {
	door = strings.TrimSpace(door)
	if door == "" || door == "—" {
		return Palace{}, false
	}
	for _, p := range palaces {
		if p.Door == door {
			return p, true
		}
	}
	return Palace{}, false
}

func findPalaceByName(palaces []Palace, name string) (Palace, bool) {
	for _, p := range palaces {
		if p.Name == name {
			return p, true
		}
	}
	return Palace{}, false
}

// pickQimenV2FocusPalaces selects up to 3 stable focus palaces for v2 reports.
func pickQimenV2FocusPalaces(palaces []Palace, chief Chief, category string) []Palace {
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

	if p, ok := findPalaceByStar(palaces, chief.ZhiFu); ok {
		add(p)
	}
	if p, ok := findPalaceByDoor(palaces, chief.ZhiShi); ok {
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
	if len(picked) > 3 {
		picked = picked[:3]
	}
	if len(picked) == 0 {
		picked = append(picked, palaces[0])
	}
	return picked
}

func formatPalacesSummaryForPrompt(palaces []Palace) string {
	return summarizeQimenV2Palaces(palaces)
}

func formatFocusPalacesSummaryForPrompt(focus []Palace) string {
	if len(focus) == 0 {
		return "（无）"
	}
	return summarizeQimenV2Palaces(focus)
}

func buildQimenV2FallbackFullContent(
	parsed parsedResultPayload,
	profile QuestionProfile,
	lens QimenLens,
	category, categoryText, methodNote, disclaimer string,
	freeContent string,
) string {
	focus := pickQimenV2FocusPalaces(parsed.Palaces, chiefFromPayload(parsed.Chief), category)
	sections := []string{
		v2SectionSummary + "\n" + buildQimenV2SummarySection(parsed, profile, lens, category, categoryText, methodNote, disclaimer, freeContent),
		v2SectionBasis + "\n" + buildQimenV2BasisSection(parsed),
		v2SectionPalaces + "\n" + buildQimenV2PalacesOverviewSection(parsed.Palaces),
		v2SectionFocusPalaces + "\n" + buildQimenV2FocusPalacesSection(focus, chiefFromPayload(parsed.Chief), dunFromPayload(parsed.Dun), category),
		v2SectionSupport + "\n" + buildQimenV2SupportSection(profile, lens, focus, category),
		v2SectionRisks + "\n" + buildQimenV2RiskSection(parsed.RiskObservations, profile, lens, focus, category),
		v2SectionPacing + "\n" + buildQimenV2PacingSection(parsed.ActionPacing, profile, lens, category),
		v2SectionReflection + "\n" + buildQimenReflectionSection(parsed.ReflectionQuestions, profile, category),
		v2SectionBoundary + "\n" + buildQimenBoundarySection(methodNote, parsed.CalculationMeta, parsed.AlgorithmVersion, parsed.Palaces),
	}
	return strings.Join(sections, "\n\n")
}

func chiefFromPayload(chief *Chief) Chief {
	if chief == nil {
		return Chief{}
	}
	return *chief
}

func dunFromPayload(dun *Dun) Dun {
	if dun == nil {
		return Dun{}
	}
	return *dun
}

func calendarFromPayload(basis *CalendarBasis) CalendarBasis {
	if basis == nil {
		return CalendarBasis{}
	}
	return *basis
}

func buildQimenV2SummarySection(
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
		categorySummaryHint(category, profile),
		fmt.Sprintf("关注主题：%s；行动节奏：%s。", lens.FocusTheme, lens.PacingTheme),
	}
	if strings.TrimSpace(freeContent) != "" {
		lines = append(lines, "可参考免费解读中的局势梳理要点，继续延伸记录与复盘。")
	}
	return strings.Join(lines, "\n")
}

func buildQimenV2BasisSection(parsed parsedResultPayload) string {
	basis := calendarFromPayload(parsed.CalendarBasis)
	dun := dunFromPayload(parsed.Dun)
	xun := parsed.Xun
	lines := []string{
		"algorithm_version：qimen-v2-poc（POC 近似排盘，非专业完整起局）。",
		fmt.Sprintf("节令参考：%s（口径：%s，时间基准：%s）。", fallbackString(basis.SolarTerm, "未指定"), fallbackString(basis.JieqiBasis, "formula_approximation"), fallbackString(basis.TimeBasis, "local_time")),
	}
	if strings.TrimSpace(basis.Note) != "" {
		lines = append(lines, basis.Note)
	}
	if dun.Type != "" || dun.Ju > 0 {
		lines = append(lines, fmt.Sprintf("阴阳遁：%s；局数：%d（来源：%s）。", dun.Type, dun.Ju, fallbackString(dun.Source, "poc_formula")))
	}
	if xun != nil && xun.XunShou != "" {
		lines = append(lines, fmt.Sprintf("旬首：%s；空亡：%s。", xun.XunShou, strings.Join(xun.EmptyBranches, "、")))
	}
	lines = append(lines, "星、门、神、天盘干、地盘干均为 POC 占位/近似口径，仅供结构化观察。")
	return strings.Join(lines, "\n")
}

func buildQimenV2PalacesOverviewSection(palaces []Palace) string {
	if len(palaces) == 0 {
		return "当前记录未包含九宫结构，无法展开宫位观察。"
	}
	lines := []string{
		fmt.Sprintf("本次 POC 九宫共 %d 宫，以下为压缩摘要（非原始 JSON）：", len(palaces)),
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

func buildQimenV2FocusPalacesSection(focus []Palace, chief Chief, dun Dun, category string) string {
	if len(focus) == 0 {
		return "暂无可引用的重点宫位，请先回到局势摘要整理问题。"
	}
	lines := []string{
		fmt.Sprintf("值符参考：%s；值使参考：%s；当前局数：%d。", chief.ZhiFu, chief.ZhiShi, dun.Ju),
		categoryFocusIntro(category),
	}
	for _, p := range focus {
		lines = append(lines, fmt.Sprintf(
			"· %s：星=%s，门=%s，神=%s；%s",
			p.Name, p.Star, p.Door, p.Deity, categoryPalaceObservationHint(category, p),
		))
	}
	return strings.Join(lines, "\n")
}

func categoryFocusIntro(category string) string {
	switch NormalizeCategory(category) {
	case "career":
		return "就事业/计划类问事，以下宫位仅作推进顺序与资源协调的结构化观察："
	case "relationship":
		return "就人际/关系类问事，以下宫位仅作沟通边界与节奏的结构化观察："
	case "study":
		return "就学习/成长类问事，以下宫位仅作复盘节奏与专注的结构化观察："
	case "decision":
		return "就决策/选择类问事，以下宫位仅作信息补齐与小步试探的结构化观察："
	default:
		return "就综合问事，以下宫位仅作问题整理与风险收敛的结构化观察："
	}
}

func categoryPalaceObservationHint(category string, p Palace) string {
	switch NormalizeCategory(category) {
	case "career":
		if p.Name == "乾六宫" || p.Name == "离九宫" {
			return "可观察推进主线与对外表达节奏，避免一次承担过多线程。"
		}
		return "可观察当前线程是否需先整理顺序，再安排小动作验证。"
	case "relationship":
		if p.Name == "兑七宫" || p.Name == "巽四宫" {
			return "可观察沟通方式与边界感，避免把一次互动放大成整体判断。"
		}
		return "可观察互动中的误解点，优先记录感受再调整节奏。"
	case "study":
		if p.Name == "坤二宫" || p.Name == "离九宫" {
			return "可观察复盘方法与阶段目标，安排可完成的专注块。"
		}
		return "可观察学习状态与方法匹配度，先小步验证再调整。"
	case "decision":
		if p.Name == "艮八宫" || p.Name == "坎一宫" {
			return "可观察选项约束与信息缺口，优先补齐再小步试探。"
		}
		return "可观察决策压力来源，列出约束与可接受范围。"
	default:
		if p.Name == "中五宫" || p.Name == "坤二宫" {
			return "可观察问题核心与承载能力，先整理再行动。"
		}
		return "可观察当前局势中的可验证一步，避免过度解读。"
	}
}

func buildQimenV2SupportSection(profile QuestionProfile, lens QimenLens, focus []Palace, category string) string {
	lines := []string{
		lens.SupportTheme + "。",
		fmt.Sprintf("可借助「%s」相关已有资源，先明确最小可验证一步。", profile.IntentType),
	}
	if len(focus) > 0 {
		p := focus[0]
		lines = append(lines, fmt.Sprintf("从 %s（%s / %s）出发，先记录一个今天能完成的小动作。", p.Name, p.Star, p.Door))
	}
	switch NormalizeCategory(category) {
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

func buildQimenV2RiskSection(risks []string, profile QuestionProfile, lens QimenLens, focus []Palace, category string) string {
	items := append([]string{lens.CautionTheme + "。"}, risks...)
	switch NormalizeCategory(category) {
	case "career":
		items = append(items, "执行分散、节奏过快时，可能一次承担过多线程。")
	case "relationship":
		items = append(items, "把一次互动放大成整体判断，可能增加误解。")
	case "study":
		items = append(items, "只比较结果而忽略方法，可能让学习节奏失衡。")
	case "decision":
		items = append(items, "在选项未写清前急于选择，可能遗漏关键约束。")
	default:
		items = append(items, "把 POC 九宫结构当作确定答案，可能偏离自我反思初衷。")
	}
	if len(focus) > 1 {
		p := focus[1]
		items = append(items, fmt.Sprintf("关注 %s（%s）时，宜先观察再行动，避免过度解读。", p.Name, p.Door))
	}
	if profile.DecisionPressure == "高" {
		items = append(items, "犹豫与担心反复切换，可能让行动迟迟无法启动。")
	}
	if len(items) > 5 {
		items = items[:5]
	}
	return strings.Join(items, "\n")
}

func buildQimenV2PacingSection(pacing string, profile QuestionProfile, lens QimenLens, category string) string {
	if strings.TrimSpace(pacing) == "" {
		pacing = lens.PacingTheme + "：先观察再行动。"
	}
	return pacing + "\n\n" + categoryPacingHint(category, profile)
}

func fallbackString(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
