package bazi

import "strings"

// fullReportForbiddenPhrases must not appear as report body advice (may appear in boundary disclaimers listing prohibitions).
var fullReportForbiddenPhrases = []string{
	"精准预测", "精准算命", "一生命运", "婚姻必然", "财运必然", "婚姻财运预测",
	"必成", "必败", "大吉", "大凶", "必发财", "必复合", "保证发财", "保证复合",
	"改运", "化灾", "转运", "改运化解", "看广告改运",
	"疾病寿命", "预测未来", "保证结果", "百分百", "注定", "必然", "一定会", "一定得",
	"投资建议", "医疗建议", "法律建议", "赌博建议", "军事行动建议",
}

func containsForbiddenReportPhrase(content string) bool {
	body := reportBodyExcludingBoundary(content)
	for _, phrase := range fullReportForbiddenPhrases {
		if phrase != "" && strings.Contains(body, phrase) {
			return true
		}
	}
	return false
}

func reportBodyExcludingBoundary(content string) string {
	markers := []string{
		sectionBoundary,
		"【七、边界声明】",
		"七、边界声明",
		"【7. 免责声明】",
		"7. 免责声明",
	}
	for _, marker := range markers {
		if idx := strings.Index(content, marker); idx >= 0 {
			return content[:idx]
		}
	}
	return content
}
