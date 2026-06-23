package interpretation

import (
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/report"
)

const disclaimerText = "本内容仅供娱乐和传统文化参考，不构成现实决策建议。"

type GenerateInput struct {
	DivinationID    int64
	Question        string
	CategoryName    string
	PrimaryHexagram *model.Hexagram
	ChangedHexagram *model.Hexagram
	MovingLines     []int
	Lines           []model.Line
	LineSnapshot    string
	FreeContent     string
}

func BuildFreeContent(in GenerateInput) string {
	if isDailyFortune(in.CategoryName) {
		return buildDailyFreeContent(in)
	}
	primarySummary := safeSummary(in.PrimaryHexagram)
	changedSummary := safeSummary(in.ChangedHexagram)
	primaryFullName := safeFullName(in.PrimaryHexagram)
	changedFullName := safeFullName(in.ChangedHexagram)
	categoryName := strings.TrimSpace(in.CategoryName)
	if categoryName == "" {
		categoryName = "此事"
	}

	movingHint := buildMovingLinesHint(in.MovingLines, in.Lines)

	var b strings.Builder
	fmt.Fprintf(&b, "你问的是「%s」相关的问题：「%s」。\n\n", categoryName, strings.TrimSpace(in.Question))
	fmt.Fprintf(&b, "本卦为「%s」，它的核心提示是：%s。\n", primaryFullName, primarySummary)
	fmt.Fprintf(&b, "变卦为「%s」，说明这件事后续可能会呈现出「%s」的趋势。\n", changedFullName, changedSummary)

	if movingHint != "" {
		b.WriteString(movingHint)
		b.WriteString("\n")
	}

	b.WriteString("\n从卦象角度看，这件事更适合被理解为一种处境提醒，而不是确定的结果。")
	b.WriteString("你可以重点观察：当前资源是否足够、节奏是否过快、沟通是否清晰、")
	b.WriteString("以及自己是否仍在主动选择而非被动等待。\n\n")
	b.WriteString("建议你先问自己三个问题：\n")
	b.WriteString("1. 我现在最确定的资源是什么？\n")
	b.WriteString("2. 我最担心的风险是否真实存在？\n")
	b.WriteString("3. 下一步有没有一个低成本试探动作？\n\n")
	b.WriteString(disclaimerText)

	return b.String()
}

func BuildFullReport(in GenerateInput) model.FullReport {
	primaryName := safeFullName(in.PrimaryHexagram)
	changedName := safeFullName(in.ChangedHexagram)
	input := report.FullInput{
		Question:       in.Question,
		CategoryName:   in.CategoryName,
		PrimaryName:    primaryName,
		PrimarySummary: safeSummary(in.PrimaryHexagram),
		ChangedName:    changedName,
		ChangedSummary: safeSummary(in.ChangedHexagram),
		MovingLines:    in.MovingLines,
	}
	return report.BuildFull(input)
}

func isDailyFortune(categoryName string) bool {
	return strings.TrimSpace(categoryName) == model.DailyFortuneCategoryName
}

func buildDailyFreeContent(in GenerateInput) string {
	primarySummary := safeSummary(in.PrimaryHexagram)
	changedSummary := safeSummary(in.ChangedHexagram)
	primaryFullName := safeFullName(in.PrimaryHexagram)
	changedFullName := safeFullName(in.ChangedHexagram)
	movingHint := buildMovingLinesHint(in.MovingLines, in.Lines)

	var b strings.Builder
	b.WriteString("这一卦更适合作为今天的状态提醒，而不是对结果的预测。\n")
	b.WriteString("你可以把它理解成：今天适合先看清节奏，再决定行动力度。\n\n")
	fmt.Fprintf(&b, "本卦「%s」提示今日底色：%s。\n", primaryFullName, primarySummary)
	fmt.Fprintf(&b, "变卦「%s」提示今日后续节奏可能偏向：%s。\n", changedFullName, changedSummary)
	if movingHint != "" {
		b.WriteString(movingHint)
		b.WriteString("\n")
	}
	b.WriteString("\n今日适合：先整理状态，再决定是主动推进还是保守观察；")
	b.WriteString("沟通上宜清晰表达，行动上宜小步试探，情绪上给自己留一点缓冲空间。\n\n")
	b.WriteString("建议你今天问自己：\n")
	b.WriteString("1. 今天我最需要守住的是什么节奏？\n")
	b.WriteString("2. 有没有一件低成本就能完成的小行动？\n")
	b.WriteString("3. 哪些担心其实只是情绪，而非事实？\n\n")
	b.WriteString(disclaimerText)
	return b.String()
}

func buildMovingLinesHint(movingLines []int, lines []model.Line) string {
	if len(movingLines) == 0 {
		return "本次卦象没有动爻，说明当下局势相对稳定，可优先从本卦整体含义入手理解。"
	}

	posText := make([]string, 0, len(movingLines))
	for _, pos := range movingLines {
		posText = append(posText, fmt.Sprintf("第%d爻", pos))
	}
	hint := fmt.Sprintf("动爻出现在%s，提示这些位置的变化值得关注。", strings.Join(posText, "、"))

	for _, line := range lines {
		for _, pos := range movingLines {
			if line.Position == pos {
				if line.Value == 6 {
					hint += " 老阴动爻，象征旧有模式正在松动，适合反思哪些习惯需要更新。"
				} else if line.Value == 9 {
					hint += " 老阳动爻，象征势能正在转化，宜把握分寸，避免用力过猛。"
				}
			}
		}
	}
	return hint
}

func safeSummary(h *model.Hexagram) string {
	if h == nil || strings.TrimSpace(h.Summary) == "" {
		return "宜静观局势，顺势而为"
	}
	return strings.TrimSpace(h.Summary)
}

func safeFullName(h *model.Hexagram) string {
	if h == nil {
		return "未知卦"
	}
	if strings.TrimSpace(h.FullName) != "" {
		return h.FullName
	}
	return h.Name
}
