package bazi

import (
	"encoding/json"
	"fmt"
	"strings"
)

const fullReportDisclaimer = "本报告基于 bazi-simple-v1 简化干支文化规则生成，仅用于传统文化学习与自我反思，不等同于专业八字排盘，不构成现实决策依据。"

type parsedResultPayload struct {
	MethodNote        string       `json:"method_note"`
	Pillars           Pillars      `json:"pillars"`
	DayMaster         string       `json:"day_master"`
	FiveElements      FiveElements `json:"five_elements"`
	ReflectionFocus   string       `json:"reflection_focus"`
	ActionSuggestions []string     `json:"action_suggestions"`
}

// BuildFullContent generates a template full report from stored analysis payloads.
func BuildFullContent(resultPayload json.RawMessage, freeContent string) (string, error) {
	var parsed parsedResultPayload
	if err := json.Unmarshal(resultPayload, &parsed); err != nil {
		return "", fmt.Errorf("invalid result_payload")
	}
	if strings.TrimSpace(parsed.DayMaster) == "" {
		return "", fmt.Errorf("invalid result_payload")
	}

	hourUnknown := strings.TrimSpace(parsed.Pillars.Hour) == ""
	pillarParts := []string{
		fmt.Sprintf("年柱：%s", nonEmpty(parsed.Pillars.Year, "—")),
		fmt.Sprintf("月柱：%s", nonEmpty(parsed.Pillars.Month, "—")),
		fmt.Sprintf("日柱：%s", nonEmpty(parsed.Pillars.Day, "—")),
	}
	if hourUnknown {
		pillarParts = append(pillarParts, "时柱：时辰未知，本次不生成时柱")
	} else {
		pillarParts = append(pillarParts, fmt.Sprintf("时柱：%s", parsed.Pillars.Hour))
	}

	e := parsed.FiveElements
	elementSection := fmt.Sprintf(
		"木 %d、火 %d、土 %d、金 %d、水 %d。\n该分布来自简化规则下的干支示意，可作为观察性格与行动风格的参考，不等同于专业旺衰判断。",
		e.Wood, e.Fire, e.Earth, e.Metal, e.Water,
	)

	methodNote := strings.TrimSpace(parsed.MethodNote)
	if methodNote == "" {
		methodNote = MethodNote
	}

	dayMasterSection := fmt.Sprintf(
		"日主「%s」可作为理解自身行动风格的切入点。结合简化五行分布，你可以观察自己在节奏、表达、执行与休息之间的偏好，并据此做温和的自我调整。",
		parsed.DayMaster,
	)

	reflectionQuestions := []string{
		"最近哪些情境让我更容易进入稳定状态？",
		"我在压力之下通常会如何反应？",
		"哪些行动方式对我更有帮助，哪些需要适度收敛？",
	}
	if strings.TrimSpace(parsed.ReflectionFocus) != "" {
		reflectionQuestions = append([]string{parsed.ReflectionFocus}, reflectionQuestions...)
	}

	suggestions := append([]string{}, parsed.ActionSuggestions...)
	if len(suggestions) == 0 {
		suggestions = []string{
			"本周选择一件小事，按固定节奏完成并记录感受。",
			"每天留出 10 分钟做无干扰的自我整理。",
		}
	}

	observationDirections := []string{
		"记录一周内的精力高峰与低谷时段。",
		"观察自己在沟通、学习、休息三类活动中的投入比例。",
		"留意哪些环境或节奏更容易让你保持专注。",
	}
	if hourUnknown {
		observationDirections = append([]string{
			"时辰未知，本次不生成时柱，相关内容仅基于已知信息进行简化分析。",
		}, observationDirections...)
	}

	freeSnippet := strings.TrimSpace(freeContent)
	if freeSnippet != "" {
		observationDirections = append(observationDirections, "可参考免费解读中的自我反思要点，继续延伸记录。")
	}

	sections := []string{
		"【完整报告说明】\n" + fullReportDisclaimer,
		"【1. 简化干支示意】\n" + strings.Join(pillarParts, "\n"),
		"【2. 五行倾向展开】\n" + elementSection,
		"【3. 日主与行动风格】\n" + dayMasterSection,
		"【4. 自我反思问题】\n" + strings.Join(reflectionQuestions, "\n"),
		"【5. 近期行动建议】\n" + strings.Join(suggestions, "\n"),
		"【6. 适合记录的观察方向】\n" + strings.Join(observationDirections, "\n"),
		"【7. 免责声明】\n" + fullReportDisclaimer + "\n" + methodNote,
	}
	return strings.Join(sections, "\n\n"), nil
}

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
