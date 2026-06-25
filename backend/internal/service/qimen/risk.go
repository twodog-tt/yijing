package qimen

import (
	"context"
	"strings"
	"unicode/utf8"
)

var staticBlockedKeywords = []string{
	"赌博", "彩票", "下注", "博彩",
	"投资", "股票", "基金", "理财", "炒股",
	"医疗", "疾病", "癌症", "手术", "用药", "诊断",
	"法律", "诉讼", "犯罪", "违法", "判刑",
	"军事", "战争", "武器",
	"自杀", "自残", "伤害他人", "杀人",
	"必死", "必败", "必成", "必发财", "必复合",
	"改运", "化灾", "精准预测", "精准算命",
	"大凶", "保证", "保证结果", "一定会", "注定",
	"转运", "趋吉避凶",
}

type questionRiskChecker interface {
	CheckQuestion(ctx context.Context, question string) (bool, error)
}

func ValidateQuestionLength(question string) bool {
	length := utf8.RuneCountInString(strings.TrimSpace(question))
	return length >= MinQuestionRunes && length <= MaxQuestionRunes
}

func containsBlockedKeyword(question string) bool {
	normalized := strings.ToLower(strings.TrimSpace(question))
	for _, kw := range staticBlockedKeywords {
		if kw == "" {
			continue
		}
		if strings.Contains(normalized, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func checkQuestionRisk(ctx context.Context, checker questionRiskChecker, question string) (bool, error) {
	if containsBlockedKeyword(question) {
		return true, nil
	}
	if checker == nil {
		return false, nil
	}
	return checker.CheckQuestion(ctx, question)
}

func trimLower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
