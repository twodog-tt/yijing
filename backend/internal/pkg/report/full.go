package report

import (
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
)

const DisclaimerText = "本内容仅供娱乐和传统文化参考，不构成现实决策建议。"

type FullInput struct {
	Question        string
	CategoryName    string
	PrimaryName     string
	PrimarySummary  string
	ChangedName     string
	ChangedSummary  string
	MovingLines     []int
}

func BuildFull(in FullInput) model.FullReport {
	if strings.TrimSpace(in.CategoryName) == model.DailyFortuneCategoryName {
		return BuildDailyFull(in)
	}
	categoryName := strings.TrimSpace(in.CategoryName)
	if categoryName == "" {
		categoryName = "此事"
	}
	primaryName := in.PrimaryName
	if primaryName == "" {
		primaryName = "未知卦"
	}
	changedName := in.ChangedName
	if changedName == "" {
		changedName = "未知卦"
	}
	primarySummary := in.PrimarySummary
	if primarySummary == "" {
		primarySummary = "宜静观局势，顺势而为"
	}
	changedSummary := in.ChangedSummary
	if changedSummary == "" {
		changedSummary = "宜静观局势，顺势而为"
	}

	summary := fmt.Sprintf("围绕「%s」的卦象提示：以「%s」为底色，向「%s」演化，重在自我觉察与行动取舍。", categoryName, primaryName, changedName)

	overall := fmt.Sprintf(
		"你所问的「%s」，在传统文化语境下可借助「%s」来理解当下处境。%s 这一象，并不指向某个必然结局，"+
			"而是在提醒你：此刻更适合先厘清自己的位置，再决定推进、等待或调整。卦象更像一面镜子，"+
			"帮助你看清资源、节奏与心态是否匹配，而不是替你做决定。",
		strings.TrimSpace(in.Question), primaryName, primaryName,
	)

	currentState := fmt.Sprintf(
		"本卦「%s」显示：%s。结合你的问题，这往往意味着你正处在需要重新校准预期的阶段——"+
			"也许外在环境尚未完全明朗，但内在方向感正在形成。变卦「%s」则提示：%s。"+
			"若你感到犹豫，未必是能力不足，更可能是信息尚未收齐，或节奏需要再稳一些。",
		primaryName, primarySummary, changedName, changedSummary,
	)

	return model.FullReport{
		Summary:             summary,
		Overall:             overall,
		CurrentState:        currentState,
		Opportunity:         buildOpportunityText(primarySummary, len(in.MovingLines) > 0),
		Risk:                buildRiskText(changedName, changedSummary),
		ActionSteps:         buildActionSteps(categoryName),
		EmotionReminder:     buildEmotionReminder(categoryName),
		ReflectionQuestions: buildReflectionQuestions(categoryName, primaryName),
		Disclaimer:          DisclaimerText,
	}
}

func BuildDailyFull(in FullInput) model.FullReport {
	primaryName := in.PrimaryName
	if primaryName == "" {
		primaryName = "未知卦"
	}
	changedName := in.ChangedName
	if changedName == "" {
		changedName = "未知卦"
	}
	primarySummary := in.PrimarySummary
	if primarySummary == "" {
		primarySummary = "宜静观局势，顺势而为"
	}
	changedSummary := in.ChangedSummary
	if changedSummary == "" {
		changedSummary = "宜静观局势，顺势而为"
	}

	summary := fmt.Sprintf("今日卦象提醒：以「%s」为底色，向「%s」演化，重在整理节奏与行动分寸。", primaryName, changedName)
	overall := fmt.Sprintf(
		"今日这一卦更适合作为状态整理，而不是对具体结果的预测。%s 所呈现的象意，是在邀请你先看清今天的整体节奏："+
			"哪些地方适合主动，哪些地方适合保守，哪些情绪需要被看见而不是被放大。你可以把它当作一天的行动提醒，而不是命运判断。",
		primaryName,
	)
	currentState := fmt.Sprintf(
		"从本卦「%s」看，今天你可能处在「%s」的状态里；变卦「%s」则提示后续节奏可能偏向「%s」。"+
			"这更像是在描述你今天的能量分布与注意力方向，而不是断定某事必然发生。",
		primaryName, primarySummary, changedName, changedSummary,
	)

	return model.FullReport{
		Summary:             summary,
		Overall:             overall,
		CurrentState:        currentState,
		Opportunity:         buildDailyOpportunityText(primarySummary, len(in.MovingLines) > 0),
		Risk:                buildDailyRiskText(changedName, changedSummary),
		ActionSteps:         buildDailyActionSteps(),
		EmotionReminder:     buildDailyEmotionReminder(),
		ReflectionQuestions: buildDailyReflectionQuestions(primaryName),
		Disclaimer:          DisclaimerText,
	}
}

func buildDailyOpportunityText(base string, hasMoving bool) string {
	if hasMoving {
		return fmt.Sprintf(
			"今日若有动爻变化，往往意味着僵滞可能被松动。%s 提示你可以抓住「小步调整」的机会："+
				"先完成一件具体小事，再决定是否加大投入，比一次性押注更稳妥。",
			base,
		)
	}
	return fmt.Sprintf(
		"今日局势相对稳定，适合用来整理、复盘与铺垫。%s 提示你可以利用相对平静的节奏，"+
			"补齐信息、厘清优先级，为接下来的行动做好准备。",
		base,
	)
}

func buildDailyRiskText(changedName, changedSummary string) string {
	return fmt.Sprintf(
		"今日需要留意的风险包括：因急躁而过度行动、在信息不全时做绝对化判断，或因情绪波动而忽略身体与休息。"+
			"变卦「%s」也在提醒：%s。若感到压力上升，可先暂停扩大投入，给自己一点缓冲时间。",
		changedName, changedSummary,
	)
}

func buildDailyActionSteps() []string {
	return []string{
		"用 10 分钟写下今天最重要的 1–2 件事，区分「必须做」与「可以缓一缓」。",
		"安排一个低成本的小行动（如一次简短沟通、半小时整理），验证今天的节奏是否合适。",
		"在傍晚做一次简短复盘：今天哪些状态被验证了，哪些只是担心。",
	}
}

func buildDailyEmotionReminder() string {
	return "今天的情绪起伏很正常。卦象不是来加重负担，而是提醒你：允许不确定，也允许慢慢来。" +
		"若感到紧绷，可以先从呼吸、散步、喝水等基础自我照顾开始，再讨论具体行动。"
}

func buildDailyReflectionQuestions(primaryName string) []string {
	return []string{
		"今天我最想守住的状态是什么？",
		fmt.Sprintf("「%s」这一象中，哪一点最贴近我今天的真实感受？", primaryName),
		"如果今晚回顾今天，我会希望自己完成了哪一个小小进步？",
	}
}

func buildOpportunityText(base string, hasMoving bool) string {
	if hasMoving {
		return fmt.Sprintf(
			"动爻带来的变化并非全是压力，也可能意味着僵局将被打破。%s 所蕴含的积极面在于："+
				"当你愿意调整方法、放慢节奏或主动沟通时，往往能找到新的突破口。机会通常藏在"+
				"「把大问题拆成小步骤」的过程里，而不是一次性的完美答案。",
			base,
		)
	}
	return fmt.Sprintf(
		"局势相对稳定时，反而适合夯实基础。%s 提示你可以利用这段相对平静的时间，整理资源、"+
			"补齐信息、明确优先级。机会不在于追逐每一个变量，而在于把可控的部分做到位。",
		base,
	)
}

func buildRiskText(changedName, changedSummary string) string {
	return fmt.Sprintf(
		"需要留意的风险包括：在信息不完整时仓促下结论、因焦虑而过度行动，或忽视沟通中的细节。"+
			"变卦「%s」也在提醒：%s。若你感到压力上升，可先暂停扩大投入，"+
			"用一周时间记录事实与感受，区分「真实风险」与「想象风险」。",
		changedName, changedSummary,
	)
}

func buildActionSteps(categoryName string) []string {
	return []string{
		fmt.Sprintf("用一页纸写下与「%s」相关的已知事实、未知项与可选项，先分清什么是证据、什么是猜测。", categoryName),
		"为自己设定一个低成本试探动作（例如一次沟通、一个小实验、半天复盘），在可控范围内验证方向。",
		"安排固定的自我复盘时间，记录情绪起伏与决策依据，避免在情绪波动时做重大承诺。",
		"若涉及他人，优先做一次清晰、温和的表达，确认彼此期待是否一致，再决定下一步节奏。",
	}
}

func buildEmotionReminder(categoryName string) string {
	return fmt.Sprintf(
		"在「%s」相关议题上，焦虑与期待常常同时存在，这很正常。卦象不是来加重你的负担，"+
			"而是邀请你以更平静的方式看自己：允许不确定，也允许慢慢来。若情绪持续紧绷，"+
			"可以先从睡眠、散步、写日记等基础自我照顾开始，再讨论具体行动。",
		categoryName,
	)
}

func buildReflectionQuestions(categoryName, primaryName string) []string {
	return []string{
		fmt.Sprintf("关于「%s」，我真正想改善的是什么，而不是别人期待我做什么？", categoryName),
		fmt.Sprintf("「%s」这一象中，哪一点最贴近我当下的真实感受？", primaryName),
		"如果三个月后再回看今天，我会希望自己今天做了什么小步行动？",
	}
}
