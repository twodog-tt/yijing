package qimen

import (
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func Calculate(question, category string, now time.Time) (CalculationResult, error) {
	question = strings.TrimSpace(question)
	if !ValidateQuestionLength(question) {
		return CalculationResult{}, fmt.Errorf("invalid question length")
	}

	category = NormalizeCategory(category)
	if !ValidateCategory(category) {
		return CalculationResult{}, fmt.Errorf("invalid category")
	}

	if now.IsZero() {
		now = clock.Now()
	}
	now = now.In(clock.Location())
	bucket := timeBucketFor(now)

	profile := ExtractQuestionProfile(question, category)
	lens := BuildQimenLens(profile, category)
	seed := hashSeed(question, category, bucket, profile.IntentType, profile.TimeHorizon)

	overview := buildSituationOverview(category, profile, lens, bucket, seed)
	risks := buildRiskObservations(category, profile, lens, bucket, seed)
	pacing := buildActionPacing(category, profile, lens, bucket)
	reflections := buildReflectionQuestions(category, profile, lens, seed)
	actions := buildActionSuggestions(category, profile, lens, bucket, seed)

	return CalculationResult{
		Question: question,
		Category: category,
		TimeContext: TimeContext{
			CreatedAt:  now.Format(time.RFC3339),
			TimeBucket: bucket,
		},
		QuestionProfile:     profile,
		QimenLens:           lens,
		DifferentiationSeed: BuildDifferentiationSeed(category, bucket),
		SafeQuestionSummary: BuildSafeQuestionSummary(profile),
		SituationOverview:   overview,
		RiskObservations:    risks,
		ActionPacing:        pacing,
		ReflectionQuestions: reflections,
		ActionSuggestions:   actions,
		MethodNote:          MethodNote,
		Limits:              calculationLimits,
	}, nil
}

func hashSeed(parts ...string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.Join(parts, "|")))
	return h.Sum32()
}

func buildSituationOverview(category string, profile QuestionProfile, lens QimenLens, bucket string, seed uint32) string {
	opening := situationOpening(category, profile)
	focusLine := fmt.Sprintf("从简化奇门文化视角看，当前更宜把「%s」作为观察主线，关注主题偏向「%s」。", profile.IntentType, lens.FocusTheme)
	horizonLine := ""
	switch profile.TimeHorizon {
	case "短期":
		horizonLine = "你提到的时间感偏近，更适合先处理眼前可验证的一步，而非一次定终身。"
	case "中期":
		horizonLine = "你提到的时间感偏阶段，适合把问题拆成几个可观察的小周期。"
	case "长期":
		horizonLine = "你提到的时间感偏长期，适合先明确方向感，再安排近期小动作。"
	}
	pressureLine := ""
	switch profile.DecisionPressure {
	case "高":
		pressureLine = "当前决策压力偏高，先把事实与情绪分开记录，会比急着下结论更稳。"
	case "低":
		pressureLine = "当前决策压力不算高，可以把重点放在观察与记录，而非立刻定案。"
	}
	parts := []string{opening, focusLine}
	if horizonLine != "" {
		parts = append(parts, horizonLine)
	}
	if pressureLine != "" {
		parts = append(parts, pressureLine)
	}
	if bucket == "night" {
		parts = append(parts, "夜间时段更适合慢下来整理，而非仓促定论。")
	}
	if seed%2 == 0 {
		parts = append(parts, "可把关注点放在「当下能做什么」，而非一次看清全部。")
	}
	return strings.Join(parts, " ")
}

func situationOpening(category string, profile QuestionProfile) string {
	switch category {
	case "career":
		return fmt.Sprintf("就事业/计划而言，当前局势更像是在整理推进顺序与资源协调，适合先观察再推进，而不是一次性做满决策。")
	case "relationship":
		return fmt.Sprintf("就人际/关系而言，当前局势更像是在梳理「%s」相关的沟通节奏与边界，适合先理解彼此节奏，再调整互动方式。", profile.IntentType)
	case "study":
		return fmt.Sprintf("就学习/成长而言，当前局势更像是在调整「%s」相关的方法与节奏，适合先复盘状态，再安排下一步。", profile.IntentType)
	case "decision":
		return fmt.Sprintf("就决策/选择而言，当前局势更像是在权衡选项与代价，适合先补齐信息，再分步推进。")
	default:
		return fmt.Sprintf("就综合问题而言，当前局势更像是在做自我观察与节奏整理，适合先稳住心态，再安排与「%s」相关的小步行动。", profile.IntentType)
	}
}

func buildRiskObservations(category string, profile QuestionProfile, lens QimenLens, bucket string, seed uint32) []string {
	items := []string{
		lens.CautionTheme + "。",
		commonRiskForProfile(profile),
	}
	switch category {
	case "career":
		items = append(items, "只关注短期结果，可能忽视长期节奏与执行边界。")
	case "relationship":
		items = append(items, "把一次互动放大成整体判断，可能增加误解。")
	case "study":
		items = append(items, "只比较结果而忽略方法，可能让学习节奏失衡。")
	case "decision":
		items = append(items, "在选项未写清前急于选择，可能遗漏关键约束。")
	}
	if profile.DecisionPressure == "高" {
		items = append(items, "在犹豫与担心之间反复切换，可能让行动迟迟无法启动。")
	}
	if bucket == "morning" {
		items = append(items, "上午时段容易高估可执行时间，建议预留缓冲。")
	}
	if seed%3 == 0 {
		items = append(items, "把简化规则当作确定答案，可能偏离自我反思初衷。")
	}
	return dedupeLimit(items, 3)
}

func commonRiskForProfile(profile QuestionProfile) string {
	switch profile.RiskTone {
	case "保守":
		return "过度放大风险信号，可能让本可小步尝试的行动也被无限推迟。"
	case "积极":
		return "在信息不完整时仓促行动，容易放大情绪波动。"
	default:
		return "过度依赖单一结论，可能忽略现实细节与变化。"
	}
}

func buildActionPacing(category string, profile QuestionProfile, lens QimenLens, bucket string) string {
	pacingIntro := pacingIntroForTheme(lens.PacingTheme)
	steps := pacingStepsForCategory(category, profile)
	text := pacingIntro + " " + steps
	switch bucket {
	case "morning":
		text += " 上午适合启动与规划，下午再复核细节。"
	case "day":
		text += " 白天适合推进可执行事项，避免同时处理过多线程。"
	case "evening":
		text += " 傍晚适合复盘与收束，不宜临时增加重大决定。"
	case "night":
		text += " 夜间更适合整理与休息，重大行动可留到次日再定。"
	}
	return text
}

func pacingIntroForTheme(theme string) string {
	switch theme {
	case "适合先观察":
		return "节奏建议：先观察再行动。"
	case "小步试探":
		return "节奏建议：小步试探，用低成本动作验证假设。"
	case "稳步推进":
		return "节奏建议：稳步推进，保持可执行的连续感。"
	case "暂缓决策":
		return "节奏建议：暂缓重大决策，先补齐信息与选项。"
	default:
		return "节奏建议：先整理，再小步行动。"
	}
}

func pacingStepsForCategory(category string, profile QuestionProfile) string {
	switch category {
	case "career":
		return "建议分三步：先整理现状与目标，再列出本周一件与「" + profile.IntentType + "」相关的小动作，最后定期复盘是否偏离节奏。"
	case "relationship":
		return "建议分三步：先记录近期互动感受，再明确一次与「" + profile.IntentType + "」相关的可沟通边界，最后观察调整后的变化。"
	case "study":
		return "建议分三步：先复盘当前方法与精力，再设定一段可完成的学习块，最后记录收获与卡点。"
	case "decision":
		return "建议分三步：先写下选项与约束，再列出每个选项的代价与收益，最后选择一个小步验证。"
	default:
		return "建议分三步：先观察当下状态，再安排一件与「" + profile.IntentType + "」相关的小事，最后记录感受与下一步。"
	}
}

func buildReflectionQuestions(category string, profile QuestionProfile, lens QimenLens, seed uint32) []string {
	base := map[string][]string{
		"career": {
			"我真正想推进的核心目标是什么？",
			"哪些阻力来自外部，哪些来自自己的节奏？",
		},
		"relationship": {
			"我期待的关系互动方式是什么？",
			"哪些沟通方式让我更稳定、更被理解？",
		},
		"study": {
			"当前学习方法是否匹配我的精力状态？",
			"哪一个小进步值得先被记录？",
		},
		"decision": {
			"我做出选择时最在意的是什么？",
			"如果信息仍不完整，我能接受的最小下一步是什么？",
		},
		"general": {
			"此刻我最需要整理的是情绪、信息还是行动？",
			"如果把问题拆小，第一步可以是什么？",
		},
	}
	items := append([]string{}, base[category]...)
	items = append(items, reflectionForIntent(profile.IntentType)...)
	if seed%2 == 1 {
		items = append(items, "我是否把简化解读当作确定结果，而忽视了现实验证？")
	}
	return dedupeLimit(items, 4)
}

func reflectionForIntent(intent string) []string {
	switch intent {
	case "推进计划":
		return []string{"如果只做最小可验证的一步，我会先做什么？"}
	case "关系沟通":
		return []string{"这次沟通里，我最想被理解的一点是什么？"}
	case "学习节奏":
		return []string{"最近哪一段学习状态最值得复盘？"}
	case "决策选择":
		return []string{"我还缺哪一条信息，才能做更稳妥的比较？"}
	default:
		return []string{"如果把「" + intent + "」拆成更小的问题，我会先观察什么？"}
	}
}

func buildActionSuggestions(category string, profile QuestionProfile, lens QimenLens, bucket string, seed uint32) []string {
	items := append([]string{}, actionForIntent(profile, category)...)
	items = append(items, lens.SupportTheme+"。")
	if bucket == "night" {
		items = append(items, "夜间优先休息与整理，避免在疲劳状态下做重大决定。")
	}
	if seed%3 == 0 {
		items = append(items, "定期回看记录，检查行动是否仍符合当前节奏。")
	}
	return dedupeLimit(items, 3)
}

func actionForIntent(profile QuestionProfile, category string) []string {
	switch profile.IntentType {
	case "推进计划":
		return []string{
			"用一页纸写下现状、目标与本周一件可执行的小事。",
			"与可信任的同事或朋友做一次非承诺性的思路整理。",
		}
	case "关系沟通":
		return []string{
			"先记录一次最近的互动感受，再决定是否沟通。",
			"明确一条与「" + profile.IntentType + "」相关的可执行边界或期待，避免一次谈过多议题。",
		}
	case "学习节奏":
		return []string{
			"设定 25–40 分钟的专注学习块，并记录收获。",
			"把卡点拆成「已知 / 未知 / 下一步查什么」。",
		}
	case "决策选择":
		return []string{
			"列出选项、约束与可接受的最坏情况。",
			"先做一个低成本的试探动作，再决定是否加码。",
		}
	default:
		switch category {
		case "decision":
			return []string{
				"把选项写成清单，并标注每条还缺什么信息。",
				"先安排一个低成本的试探动作，观察反馈后再比较。",
			}
		case "general":
			return []string{
				"安排一件今天能完成的小事，建立可控感。",
				"把问题改写成「我想观察什么」而非「结果会怎样」。",
			}
		default:
			return []string{
				"安排一件今天能完成的小事，建立可控感。",
				"把问题改写成「我想观察什么」而非「结果会怎样」。",
			}
		}
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
