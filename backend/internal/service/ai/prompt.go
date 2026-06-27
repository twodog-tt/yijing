package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

const systemPrompt = `你是一个“传统文化易经解读助手”。

你的任务是根据用户问题和卦象信息，生成一份温和、理性、适合普通用户阅读的卦象解读报告。

重要边界：
1. 内容仅用于娱乐、传统文化学习和自我反思。
2. 不要声称可以预测未来。
3. 不要使用“必然、一定、注定、百分百、保证”等绝对词。
4. 不要提供医疗、法律、投资、赌博相关建议。
5. 不要恐吓用户。
6. 不要诱导用户付费化解问题。
7. 不要输出“改命、化灾、转运、包发财、包复合”等表达。
8. 不要替用户做重大现实决策。
9. 输出要具体、有启发、有行动建议。
10. 语气温和、克制、不玄乎。

你必须只输出 JSON，不要输出 Markdown，不要输出代码块，不要输出额外解释。`

const userPromptTemplate = `请基于后文【本次输入】生成卦象解读报告。

请输出 JSON，结构如下：
{
  "summary": "一句话总结，30字以内",
  "overall": "总体判断，150-250字",
  "current_state": "当前处境，150-250字",
  "opportunity": "机会点，120-200字",
  "risk": "风险点，120-200字",
  "action_steps": ["行动建议1", "行动建议2", "行动建议3"],
  "emotion_reminder": "情绪提醒，100-180字",
  "reflection_questions": ["反思问题1", "反思问题2", "反思问题3"],
  "disclaimer": "本内容仅供娱乐和传统文化参考，不构成现实决策建议。"
}

写作要求：
- 必须结合本卦、变卦、动爻、六爻快照与免费解读，不要只复述字段。
- 语气温和、克制，偏向传统文化学习、自我观察和行动整理。
- 不做精准预测，不替用户做现实决策，不输出医疗、法律、投资、赌博建议。
- 不输出会话标识、payload、prompt、密钥或其他内部信息。
{{special_requirements}}

【本次输入】
用户问题：{{question}}
事项类型：{{category_name}}

本卦：
- 卦名：{{primary_full_name}}
- 摘要：{{primary_summary}}
- binary_code：{{primary_binary_code}}

变卦：
- 卦名：{{changed_full_name}}
- 摘要：{{changed_summary}}
- binary_code：{{changed_binary_code}}

动爻：
{{moving_lines}}

六爻快照：
{{line_snapshot}}

免费解读：
{{free_content}}`

func BuildUserPrompt(input GenerateInput) string {
	movingJSON, _ := json.Marshal(input.MovingLines)
	specialRequirements := "无额外分类要求。"
	if strings.TrimSpace(input.CategoryName) == "今日运势" {
		specialRequirements = "【今日运势特别要求】这是今日运势解读，只能围绕今天的状态、节奏、行动提醒和自我反思来写，不要预测具体事件。不要写事业决策、感情预测、财运预测、健康预测，不要使用“今天一定会发财/倒霉/出事”等表达。"
	}
	replacer := strings.NewReplacer(
		"{{special_requirements}}", specialRequirements,
		"{{question}}", input.Question,
		"{{category_name}}", input.CategoryName,
		"{{primary_full_name}}", input.PrimaryHexagram.FullName,
		"{{primary_summary}}", input.PrimaryHexagram.Summary,
		"{{primary_binary_code}}", input.PrimaryHexagram.BinaryCode,
		"{{changed_full_name}}", input.ChangedHexagram.FullName,
		"{{changed_summary}}", input.ChangedHexagram.Summary,
		"{{changed_binary_code}}", input.ChangedHexagram.BinaryCode,
		"{{moving_lines}}", string(movingJSON),
		"{{line_snapshot}}", input.LineSnapshot,
		"{{free_content}}", input.FreeContent,
	)
	return replacer.Replace(userPromptTemplate)
}

func SystemPrompt() string {
	return systemPrompt
}

func truncateQuestion(q string, max int) string {
	runes := []rune(strings.TrimSpace(q))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max]) + "…"
}

func safeHexagramName(h HexagramInfo) string {
	if h.FullName != "" {
		return h.FullName
	}
	return h.Name
}

func validateInput(input GenerateInput) error {
	if strings.TrimSpace(input.Question) == "" {
		return fmt.Errorf("question is required")
	}
	return nil
}
