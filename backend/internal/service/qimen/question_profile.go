package qimen

import (
	"strings"
	"unicode/utf8"
)

type QuestionProfile struct {
	IntentType       string `json:"intent_type"`
	TimeHorizon      string `json:"time_horizon"`
	DecisionPressure string `json:"decision_pressure"`
	RelationScope    string `json:"relation_scope"`
	RiskTone         string `json:"risk_tone"`
}

type QimenLens struct {
	FocusTheme   string `json:"focus_theme"`
	SupportTheme string `json:"support_theme"`
	CautionTheme string `json:"caution_theme"`
	PacingTheme  string `json:"pacing_theme"`
}

type DifferentiationSeed struct {
	Source string `json:"source"`
	Note   string `json:"note"`
}

const differentiationSeedNote = "仅用于生成差异化学习解读，不构成预测依据"

var keywordGroups = map[string][]string{
	"advance":      {"推进", "开始", "上线", "启动", "执行", "落地", "项目", "计划", "推进这个"},
	"choice":       {"是否", "要不要", "该不该", "换方向", "选择", "方向", "应该", "该不该"},
	"communication": {"沟通", "合作", "关系", "误会", "冲突", "同事", "伙伴", "不顺", "调整", "协调"},
	"study":        {"学习", "考试", "复习", "成长", "训练", "状态", "节奏", "专注"},
	"rhythm":       {"最近", "现在", "马上", "长期", "阶段", "节奏", "安排"},
	"risk":         {"担心", "风险", "不顺", "犹豫", "压力", "阻碍", "卡点", "困难"},
}

func ExtractQuestionProfile(question, category string) QuestionProfile {
	q := strings.TrimSpace(question)
	category = NormalizeCategory(category)

	profile := QuestionProfile{
		IntentType:       defaultIntentForCategory(category),
		TimeHorizon:      "未明确",
		DecisionPressure: "中",
		RelationScope:    defaultRelationScope(category),
		RiskTone:         "平衡",
	}

	profile.IntentType = resolveIntentType(q, category)

	if containsAny(q, []string{"最近", "现在", "马上", "本周", "今天"}) {
		profile.TimeHorizon = "短期"
	} else if containsAny(q, []string{"长期", "未来", "以后", "长远"}) {
		profile.TimeHorizon = "长期"
	} else if containsAny(q, []string{"阶段", "这段时间", "近期"}) {
		profile.TimeHorizon = "中期"
	}

	riskHits := countKeywordHits(q, keywordGroups["risk"])
	choiceHits := countKeywordHits(q, keywordGroups["choice"])
	switch {
	case riskHits >= 2 || containsAny(q, []string{"容易犹豫", "很担心", "压力很大"}):
		profile.DecisionPressure = "高"
	case choiceHits >= 1 && riskHits == 0:
		profile.DecisionPressure = "中"
	case containsAny(q, []string{"随便", "看看", "了解一下"}):
		profile.DecisionPressure = "低"
	}

	if containsAny(q, []string{"伙伴", "合作", "同事", "双方", "我们", "沟通"}) {
		profile.RelationScope = "双方"
	} else if containsAny(q, []string{"团队", "部门", "组织"}) {
		profile.RelationScope = "团队"
	} else if containsAny(q, []string{"环境", "外部", "市场", "机会"}) {
		profile.RelationScope = "外部环境"
	}

	switch {
	case riskHits >= 2:
		profile.RiskTone = "保守"
	case containsAny(q, keywordGroups["advance"]) && riskHits == 0:
		profile.RiskTone = "积极"
	default:
		profile.RiskTone = "平衡"
	}

	return profile
}

func BuildQimenLens(profile QuestionProfile, category string) QimenLens {
	category = NormalizeCategory(category)
	lens := QimenLens{
		FocusTheme:   focusThemeForIntent(profile.IntentType, category),
		SupportTheme: supportThemeFor(profile, category),
		CautionTheme: cautionThemeFor(profile, category),
		PacingTheme:  pacingThemeFor(profile, category),
	}
	return lens
}

func BuildDifferentiationSeed(category, bucket string) DifferentiationSeed {
	return DifferentiationSeed{
		Source: "category + safe question features + time bucket",
		Note:   differentiationSeedNote,
	}
}

func BuildSafeQuestionSummary(profile QuestionProfile) string {
	parts := []string{
		"问事侧重：" + profile.IntentType,
		"时间范围：" + profile.TimeHorizon,
		"决策压力：" + profile.DecisionPressure,
		"关系范围：" + profile.RelationScope,
		"风险倾向：" + profile.RiskTone,
	}
	return strings.Join(parts, "；")
}

func resolveIntentType(question, category string) string {
	type intentScore struct {
		intent string
		score  int
	}
	candidates := []intentScore{
		{"推进计划", countKeywordHits(question, keywordGroups["advance"])},
		{"关系沟通", countKeywordHits(question, keywordGroups["communication"])},
		{"学习节奏", countKeywordHits(question, keywordGroups["study"])},
		{"决策选择", countKeywordHits(question, keywordGroups["choice"])},
	}
	best := defaultIntentForCategory(category)
	bestScore := 0
	for _, item := range candidates {
		if item.score > bestScore {
			bestScore = item.score
			best = item.intent
		}
	}
	if bestScore == 0 {
		return defaultIntentForCategory(category)
	}
	return best
}

func defaultIntentForCategory(category string) string {
	switch category {
	case "career":
		return "推进计划"
	case "relationship":
		return "关系沟通"
	case "study":
		return "学习节奏"
	case "decision":
		return "决策选择"
	default:
		return "综合整理"
	}
}

func defaultRelationScope(category string) string {
	switch category {
	case "relationship":
		return "双方"
	case "career":
		return "团队"
	default:
		return "个人"
	}
}

func focusThemeForIntent(intent, category string) string {
	themes := map[string]string{
		"推进计划": "行动推进",
		"关系沟通": "沟通协调",
		"学习节奏": "节奏调整",
		"决策选择": "信息整理",
		"综合整理": "风险收敛",
	}
	if theme, ok := themes[intent]; ok {
		return theme
	}
	return themes[defaultIntentForCategory(category)]
}

func supportThemeFor(profile QuestionProfile, category string) string {
	switch profile.IntentType {
	case "推进计划":
		return "可借助的阶段性目标与已有资源，先明确「最小可验证一步」"
	case "关系沟通":
		return "可借助一次清晰、克制的表达，以及双方都能接受的沟通节奏"
	case "学习节奏":
		return "可借助短周期复盘与可完成的学习块，重建专注感"
	case "决策选择":
		return "可借助选项清单与约束条件，把比较变成可观察的信息"
	default:
		switch category {
		case "decision":
			return "可借助信息补齐与小步试探，降低一次性押注的压力"
		case "general":
			return "可借助问题拆分，把大议题还原成可观察的小问题"
		default:
			return "可借助记录与复盘，让下一步行动更可执行"
		}
	}
}

func cautionThemeFor(profile QuestionProfile, category string) string {
	switch profile.IntentType {
	case "推进计划":
		return "需要留意的执行分散与节奏过快，避免一次承担过多线程"
	case "关系沟通":
		return "需要留意的误解放大与边界模糊，避免一次谈过多议题"
	case "学习节奏":
		return "需要留意的精力透支与方法失配，避免只比较结果忽略过程"
	case "决策选择":
		return "需要留意的信息缺口与试探成本，避免在选项未写清前仓促定案"
	default:
		if profile.DecisionPressure == "高" {
			return "需要留意的犹豫循环与风险想象，先补齐事实再行动"
		}
		switch category {
		case "relationship":
			return "需要留意的情绪带入与过度解读，先观察再下结论"
		case "study":
			return "需要留意的节奏失衡与自我苛责，先恢复可完成感"
		default:
			return "需要留意的信息不完整与节奏失衡，避免把简化解读当作确定答案"
		}
	}
}

func pacingThemeFor(profile QuestionProfile, category string) string {
	switch {
	case profile.RiskTone == "保守" || profile.DecisionPressure == "高":
		return "适合先观察"
	case profile.IntentType == "推进计划" && profile.TimeHorizon == "短期":
		return "小步试探"
	case profile.IntentType == "决策选择":
		return "暂缓决策"
	case profile.IntentType == "学习节奏":
		return "稳步推进"
	case profile.IntentType == "关系沟通":
		return "小步试探"
	default:
		switch category {
		case "career":
			return "稳步推进"
		case "decision":
			return "暂缓决策"
		default:
			return "小步试探"
		}
	}
}

func containsAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if kw != "" && strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func countKeywordHits(text string, keywords []string) int {
	count := 0
	for _, kw := range keywords {
		if kw != "" && strings.Contains(text, kw) {
			count++
		}
	}
	return count
}

func validateProfileText(s string) bool {
	return utf8.RuneCountInString(strings.TrimSpace(s)) > 0
}
