package qimen

import "strings"

var fullReportForbiddenPhrases = []string{
	"精准预测", "精准算命", "必成", "必败", "大吉", "大凶", "必发财", "必复合",
	"改运", "化灾", "转运", "改运化解", "看广告改运",
	"预测未来", "保证结果", "百分百", "注定", "必然", "一定会", "一定得",
	"婚姻财运预测", "保证发财", "一生命运",
	"投资建议", "医疗建议", "法律建议", "赌博建议", "军事行动建议", "军事行动",
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
		v2SectionBoundary,
		"【七、边界声明】",
		"七、边界声明",
		"九、边界声明",
		"【九、边界声明】",
		"【8. 免责声明】",
		"8. 免责声明",
	}
	for _, marker := range markers {
		if idx := strings.Index(content, marker); idx >= 0 {
			return content[:idx]
		}
	}
	return content
}
