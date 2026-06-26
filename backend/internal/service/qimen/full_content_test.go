package qimen

import (
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func fixedNow() time.Time {
	loc := clock.Location()
	return time.Date(2024, 6, 15, 14, 0, 0, 0, loc)
}

func TestBuildFullContentUsesReportSections(t *testing.T) {
	calc, err := Calculate("我最近适合推进这个项目吗？", "career", fixedNow())
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	content, err := BuildFullContent(payload, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	for _, snippet := range []string{
		sectionSummary, sectionFocus, sectionSupport, sectionRisks,
		sectionPacing, sectionReflection, sectionBoundary,
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected full content to contain %q", snippet)
		}
	}
	if containsForbiddenReportPhrase(content) {
		t.Fatalf("full content body must not contain forbidden phrases")
	}
}

func TestQimenFullReportSampleDifferentiation(t *testing.T) {
	cases := []struct {
		name     string
		question string
		category string
		marker   string
	}{
		{"career", "我最近适合推进这个项目吗？", "career", "推进顺序"},
		{"relationship", "我和合作伙伴沟通总是不顺，应该怎么调整？", "relationship", "沟通节奏"},
		{"study", "我最近学习状态不好，怎么安排节奏？", "study", "复盘"},
		{"decision", "我是否应该现在换一个方向发展？", "decision", "信息补齐"},
		{"general", "我最近做决定容易犹豫，应该先看哪些风险？", "general", "问题整理"},
	}

	contents := make(map[string]string, len(cases))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calc, err := Calculate(tc.question, tc.category, fixedNow())
			if err != nil {
				t.Fatalf("Calculate: %v", err)
			}
			payload, err := calc.ResultPayload()
			if err != nil {
				t.Fatalf("ResultPayload: %v", err)
			}
			content, err := BuildFullContent(payload, BuildFreeContent(calc))
			if err != nil {
				t.Fatalf("BuildFullContent: %v", err)
			}
			contents[tc.name] = content

			if strings.Contains(content, tc.question) {
				t.Fatalf("full content must not contain raw question")
			}
			if !strings.Contains(content, calc.QuestionProfile.IntentType) {
				t.Fatalf("expected intent_type %q in content", calc.QuestionProfile.IntentType)
			}
			if !strings.Contains(content, tc.marker) {
				t.Fatalf("expected category marker %q in content for %s", tc.marker, tc.category)
			}
			for _, forbidden := range []string{"session_key", "input_payload", "result_payload", "prompt"} {
				if strings.Contains(content, forbidden) {
					t.Fatalf("must not contain %q", forbidden)
				}
			}
			if containsForbiddenReportPhrase(content) {
				t.Fatalf("must not contain forbidden report phrases in body")
			}
		})
	}

	if contents["career"] == contents["relationship"] {
		t.Fatalf("expected different full content across categories")
	}
}

func TestBuildFullContentForbiddenWordsInBody(t *testing.T) {
	calc, err := Calculate("我最近适合推进这个项目吗？", "career", fixedNow())
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	content, err := BuildFullContent(payload, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	body := reportBodyExcludingBoundary(content)
	for _, phrase := range []string{
		"精准预测", "必成", "必败", "大吉", "大凶", "必发财", "必复合",
		"改运", "化灾", "转运", "投资建议", "医疗建议", "法律建议", "赌博建议", "军事行动建议",
	} {
		if strings.Contains(body, phrase) {
			t.Fatalf("template content body must not contain %q", phrase)
		}
	}
}
