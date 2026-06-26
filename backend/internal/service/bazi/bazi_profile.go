package bazi

import (
	"fmt"
	"strings"
)

type BaziProfile struct {
	DayMasterObservation string `json:"day_master_observation"`
	SeasonTendency       string `json:"season_tendency"`
	ElementBalanceType   string `json:"element_balance_type"`
	ActionStyle          string `json:"action_style"`
	ReflectionTheme      string `json:"reflection_theme"`
}

type InterpretationLens struct {
	StrengthHint             string `json:"strength_hint"`
	CautionHint              string `json:"caution_hint"`
	PacingHint               string `json:"pacing_hint"`
	RelationshipWithElements string `json:"relationship_with_elements"`
}

type DifferentiationSeed struct {
	Source string `json:"source"`
	Note   string `json:"note"`
}

const differentiationSeedNote = "仅用于生成差异化学习解读，不构成预测依据"

func BuildBaziProfile(birthMonth int, dayMaster string, elements FiveElements, hourUnknown bool, pillars Pillars) BaziProfile {
	balanceType, _, _ := classifyElementBalance(elements)
	season := seasonTendency(birthMonth)
	actionStyle := actionStyleFor(dayMaster, balanceType, hourUnknown)
	dominant, _ := dominantElements(elements)
	reflectionTheme := reflectionThemeFor(dayMaster, dominant, balanceType)
	observation := dayMasterObservation(dayMaster, balanceType, season, hourUnknown)

	return BaziProfile{
		DayMasterObservation: observation,
		SeasonTendency:       season,
		ElementBalanceType:   balanceType,
		ActionStyle:          actionStyle,
		ReflectionTheme:      reflectionTheme,
	}
}

func BuildInterpretationLens(profile BaziProfile, elements FiveElements, dayMaster string, hourUnknown bool) InterpretationLens {
	dominant, weak := dominantElements(elements)
	return InterpretationLens{
		StrengthHint:             strengthHintFor(profile, dayMaster, dominant),
		CautionHint:              cautionHintFor(profile, dayMaster, weak),
		PacingHint:               pacingHintFor(profile, hourUnknown),
		RelationshipWithElements: elementRelationshipHint(elements, dominant, weak),
	}
}

func BuildDifferentiationSeed(hourKnown bool) DifferentiationSeed {
	source := "pillars + five_elements + day_master + hour_known"
	if !hourKnown {
		source = "pillars + five_elements + day_master + hour_unknown"
	}
	return DifferentiationSeed{
		Source: source,
		Note:   differentiationSeedNote,
	}
}

func BuildReflectionFocus(profile BaziProfile, dayMaster string) string {
	elementName := stemElementName(dayMaster)
	return fmt.Sprintf(
		"基于简化干支文化规则的学习参考：日主为「%s」，可从%s相关的「%s」入手；当前五行倾向为%s，行动风格偏向「%s」。",
		dayMaster,
		elementName,
		profile.ReflectionTheme,
		profile.ElementBalanceType,
		profile.ActionStyle,
	)
}

func BuildActionSuggestions(profile BaziProfile, lens InterpretationLens, dayMaster string, elements FiveElements, hourUnknown bool) []string {
	items := []string{
		actionForTheme(profile.ReflectionTheme),
		actionForStyle(profile.ActionStyle),
	}
	if dominant, weak := dominantElements(elements); dominant != "" {
		items = append(items, fmt.Sprintf(
			"五行分布中%s相对较多、%s相对较少，可作为学习参考观察节奏与投入方式，不构成现实决策依据。",
			dominant, weak,
		))
	}
	items = append(items, lens.StrengthHint+"。")
	if hourUnknown {
		items = append([]string{"时辰未知，本次不生成时柱；可从年月日三柱示意出发做自我观察。"}, items...)
	}
	return dedupeLimit(items, 4)
}

func seasonTendency(month int) string {
	switch month {
	case 3, 4, 5:
		return "春"
	case 6, 7:
		return "夏"
	case 8:
		return "长夏"
	case 9, 10, 11:
		return "秋"
	default:
		return "冬"
	}
}

func classifyElementBalance(elements FiveElements) (balanceType, dominant, weak string) {
	dominant, weak = dominantElements(elements)
	if dominant == "" {
		return "相对均衡", "", ""
	}
	items := []struct {
		name  string
		count int
	}{
		{"木", elements.Wood},
		{"火", elements.Fire},
		{"土", elements.Earth},
		{"金", elements.Metal},
		{"水", elements.Water},
	}
	maxCount, minCount := items[0].count, items[0].count
	for _, item := range items {
		if item.count > maxCount {
			maxCount = item.count
		}
		if item.count < minCount {
			minCount = item.count
		}
	}
	if maxCount-minCount <= 1 {
		return "相对均衡", dominant, weak
	}
	return "偏" + dominant, dominant, weak
}

func actionStyleFor(dayMaster, balanceType string, hourUnknown bool) string {
	if hourUnknown {
		return "先整理信息"
	}
	switch stemElementName(dayMaster) {
	case "木":
		if strings.Contains(balanceType, "偏木") {
			return "稳步推进"
		}
		return "先整理信息"
	case "火":
		return "适合小步试探"
	case "土":
		return "稳步推进"
	case "金":
		return "先整理信息"
	case "水":
		return "适合复盘调整"
	default:
		return "稳步推进"
	}
}

func reflectionThemeFor(dayMaster, dominant, balanceType string) string {
	element := stemElementName(dayMaster)
	switch element {
	case "木":
		return "执行节奏"
	case "火":
		return "沟通方式"
	case "土":
		if strings.Contains(balanceType, "偏土") {
			return "边界感"
		}
		return "执行节奏"
	case "金":
		return "边界感"
	case "水":
		return "情绪整理"
	default:
		if dominant == "木" || dominant == "火" {
			return "学习沉淀"
		}
		return "执行节奏"
	}
}

func dayMasterObservation(dayMaster, balanceType, season string, hourUnknown bool) string {
	elementName := stemElementName(dayMaster)
	text := fmt.Sprintf(
		"日主「%s」属%s，在简化规则下可作为一个自我观察切入点；出生季节倾向为%s，五行分布呈%s。",
		dayMaster, elementName, season, balanceType,
	)
	if hourUnknown {
		text += " 时辰未知，时柱未纳入本次示意。"
	}
	text += " 以上仅作传统文化学习参考，不构成命运判断。"
	return text
}

func strengthHintFor(profile BaziProfile, dayMaster, dominant string) string {
	elementName := stemElementName(dayMaster)
	switch profile.ReflectionTheme {
	case "执行节奏":
		return fmt.Sprintf("可借助%s日主带来的「%s」特点，先明确一件本周可完成的小事", elementName, profile.ActionStyle)
	case "沟通方式":
		return "可借助一次清晰、克制的表达，观察沟通后的感受变化"
	case "边界感":
		return "可借助明确一条个人边界或节奏规则，减少同时处理过多事项"
	case "学习沉淀":
		return "可借助短周期复盘，把收获与卡点分别记录下来"
	case "情绪整理":
		return "可借助情绪记录，区分「感受」与「下一步行动」"
	default:
		return fmt.Sprintf("可借助%s元素相对明显的倾向，作为观察自身节奏的一个角度", dominant)
	}
}

func cautionHintFor(profile BaziProfile, dayMaster, weak string) string {
	switch profile.ElementBalanceType {
	case "相对均衡":
		return "需要留意把简化示意当作确定结论，而忽视现实验证"
	case "偏木":
		return "需要留意节奏过快、同时推进过多线程而分散精力"
	case "偏火":
		return "需要留意情绪波动影响判断，宜先冷却再行动"
	case "偏土":
		return "需要留意过度保守或拖延，错过小步尝试的时机"
	case "偏金":
		return "需要留意过度强调标准与边界，忽略灵活调整"
	case "偏水":
		return "需要留意思虑过多而行动不足，宜先安排一件可完成的小事"
	default:
		_ = dayMaster
		if weak != "" {
			return fmt.Sprintf("需要留意%s元素相对偏弱带来的节奏失衡，宜温和补齐而非强行改变", weak)
		}
		return "需要留意把简化解读当作确定结果"
	}
}

func pacingHintFor(profile BaziProfile, hourUnknown bool) string {
	if hourUnknown {
		return "先整理已知信息，暂缓对时辰相关细节的推断"
	}
	switch profile.ActionStyle {
	case "稳步推进":
		return "保持可执行的连续感，分步完成而非一次定案"
	case "先整理信息":
		return "先记录现状与选项，再安排小动作验证"
	case "适合小步试探":
		return "用低成本动作试探，观察反馈后再决定是否加码"
	case "适合复盘调整":
		return "定期回看记录，根据实际感受微调节奏"
	default:
		return "先观察，再小步行动"
	}
}

func elementRelationshipHint(elements FiveElements, dominant, weak string) string {
	if dominant == "" {
		return "五行计数较为接近，可把各元素理解为不同生活面向的观察维度，而非强弱定论"
	}
	parts := []string{
		fmt.Sprintf("简化计数中%s相对突出", dominant),
	}
	if weak != "" && weak != dominant {
		parts = append(parts, fmt.Sprintf("%s相对偏弱", weak))
	}
	parts = append(parts, "可把差异理解为「哪些面向更容易被注意到、哪些需要主动补齐观察」")
	return strings.Join(parts, "；") + "。"
}

func actionForTheme(theme string) string {
	switch theme {
	case "边界感":
		return "本周明确一条可执行的边界或节奏规则，并记录遵守后的感受。"
	case "执行节奏":
		return "选择一项可执行的小行动，按固定节奏完成并复盘，而非追求结论。"
	case "沟通方式":
		return "记录一次近期沟通的感受：什么表达方式让你更稳定、更被理解。"
	case "学习沉淀":
		return "设定一段 25–40 分钟的专注学习或复盘块，并记录收获与卡点。"
	case "情绪整理":
		return "把最近一件让你有感受的小事写下来，区分情绪与可行动项。"
	default:
		return "记录近期一件让你有感受的小事，尝试从性格倾向角度做自我观察。"
	}
}

func actionForStyle(style string) string {
	switch style {
	case "稳步推进":
		return "把目标拆成「本周一件可完成的小事」，完成后再决定是否继续。"
	case "先整理信息":
		return "用一页纸写下现状、选项与约束，再安排下一步小动作。"
	case "适合小步试探":
		return "先做一个低成本的试探动作，观察反馈后再决定是否加码。"
	case "适合复盘调整":
		return "每周固定一次回看记录，检查行动是否仍符合当前节奏。"
	default:
		return "选择一项可执行的小行动，先完成再复盘。"
	}
}

func dedupeLimit(items []string, limit int) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, limit)
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}
