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

	seed := hashSeed(question, category, bucket)
	overview := buildSituationOverview(category, bucket, seed)
	risks := buildRiskObservations(category, bucket, seed)
	pacing := buildActionPacing(category, bucket)
	reflections := buildReflectionQuestions(category, seed)
	actions := buildActionSuggestions(category, bucket, seed)

	return CalculationResult{
		Question: question,
		Category: category,
		TimeContext: TimeContext{
			CreatedAt:  now.Format(time.RFC3339),
			TimeBucket: bucket,
		},
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

func buildSituationOverview(category, bucket string, seed uint32) string {
	base := map[string]string{
		"career":       "当前局势更像是在整理方向与节奏，适合先观察再推进，而不是一次性做满决策。",
		"relationship": "当前局势更像是在梳理沟通与边界，适合先理解彼此节奏，再调整互动方式。",
		"study":        "当前局势更像是在调整学习节奏与方法，适合先复盘状态，再安排下一步。",
		"decision":     "当前局势更像是在权衡选项与代价，适合先列出事实，再分步推进。",
		"general":      "当前局势更像是在做自我观察与节奏整理，适合先稳住心态，再安排行动。",
	}
	text := base[category]
	if bucket == "night" {
		text += " 夜间时段更适合慢下来整理，而非仓促定论。"
	}
	if seed%2 == 0 {
		text += " 可把关注点放在「当下能做什么」，而非一次看清全部。"
	}
	return text
}

func buildRiskObservations(category, bucket string, seed uint32) []string {
	common := []string{
		"过度依赖单一结论，可能忽略现实细节与变化。",
		"在信息不完整时仓促行动，容易放大情绪波动。",
	}
	switch category {
	case "career":
		common = append(common, "只关注短期结果，可能忽视长期节奏与边界。")
	case "relationship":
		common = append(common, "把一次互动放大成整体判断，可能增加误解。")
	case "study":
		common = append(common, "只比较结果而忽略方法，可能让学习节奏失衡。")
	case "decision":
		common = append(common, "在选项未写清前急于选择，可能遗漏关键约束。")
	}
	if bucket == "morning" {
		common = append(common, "上午时段容易高估可执行时间，建议预留缓冲。")
	}
	if seed%3 == 0 {
		common = append(common, "把简化规则当作确定答案，可能偏离自我反思初衷。")
	}
	if len(common) > 3 {
		return common[:3]
	}
	return common
}

func buildActionPacing(category, bucket string) string {
	pacing := map[string]string{
		"career":       "建议分三步：先整理现状与目标，再列出本周可执行的小动作，最后定期复盘是否偏离节奏。",
		"relationship": "建议分三步：先记录近期互动感受，再明确一次可沟通的边界或期待，最后观察调整后的变化。",
		"study":        "建议分三步：先复盘当前方法与精力，再设定一段可完成的学习块，最后记录收获与卡点。",
		"decision":     "建议分三步：先写下选项与约束，再列出每个选项的代价与收益，最后选择一个小步验证。",
		"general":      "建议分三步：先观察当下状态，再安排一件可完成的小事，最后记录感受与下一步。",
	}
	text := pacing[category]
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

func buildReflectionQuestions(category string, seed uint32) []string {
	questions := map[string][]string{
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
	items := append([]string{}, questions[category]...)
	if seed%2 == 1 {
		items = append(items, "我是否把简化解读当作确定结果，而忽视了现实验证？")
	}
	return items
}

func buildActionSuggestions(category, bucket string, seed uint32) []string {
	suggestions := map[string][]string{
		"career": {
			"用一页纸写下现状、目标与本周一件可执行的小事。",
			"与可信任的同事或朋友做一次非承诺性的思路整理。",
		},
		"relationship": {
			"先记录一次最近的互动感受，再决定是否沟通。",
			"明确一条可执行的边界或期待，避免一次谈过多议题。",
		},
		"study": {
			"设定 25–40 分钟的专注学习块，并记录收获。",
			"把卡点拆成「已知 / 未知 / 下一步查什么」。",
		},
		"decision": {
			"列出选项、约束与可接受的最坏情况。",
			"先做一个低成本的试探动作，再决定是否加码。",
		},
		"general": {
			"安排一件今天能完成的小事，建立可控感。",
			"把问题改写成「我想观察什么」而非「结果会怎样」。",
		},
	}
	items := append([]string{}, suggestions[category]...)
	if bucket == "night" {
		items = append(items, "夜间优先休息与整理，避免在疲劳状态下做重大决定。")
	}
	if seed%3 == 0 {
		items = append(items, "定期回看记录，检查行动是否仍符合当前节奏。")
	}
	if len(items) > 3 {
		return items[:3]
	}
	return items
}
