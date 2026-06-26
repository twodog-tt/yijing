package bazi_test

import (
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/service/bazi"
)

func TestFullReportForbiddenPhrasesList(t *testing.T) {
	calc, err := bazi.Calculate("1995-03-12", "mao", false)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	content, err := bazi.BuildFullContent(payload, bazi.BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	for _, phrase := range []string{
		"精准预测", "必成", "必败", "大吉", "大凶", "必发财", "必复合",
		"改运", "化灾", "转运", "投资建议", "医疗建议", "法律建议", "赌博建议", "军事行动建议",
	} {
		if strings.Contains(bazi.ReportBodyExcludingBoundaryForTest(content), phrase) {
			t.Fatalf("report body must not contain %q", phrase)
		}
	}
}
