package interpretation

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/wangxintong/yijing/backend/internal/model"
	reportpkg "github.com/wangxintong/yijing/backend/internal/pkg/report"
)

func sampleInput() GenerateInput {
	return GenerateInput{
		Question:     "我现在适不适合继续推进这个 AI 易经小程序？",
		CategoryName: "事业",
		PrimaryHexagram: &model.Hexagram{
			Name:     "大有",
			FullName: "火天大有",
			Summary:  "丰盛在握，宜守正分享，忌骄奢。",
		},
		ChangedHexagram: &model.Hexagram{
			Name:     "大壮",
			FullName: "雷天大壮",
			Summary:  "气势充沛，宜把握分寸，不可恃强。",
		},
		MovingLines: []int{6},
		Lines: []model.Line{
			{Position: 6, Value: 9, IsYang: 1, IsMoving: 1},
		},
	}
}

func TestBuildFreeContentNotEmpty(t *testing.T) {
	content := BuildFreeContent(sampleInput())
	if strings.TrimSpace(content) == "" {
		t.Fatal("free content should not be empty")
	}
	length := utf8.RuneCountInString(content)
	if length < 100 {
		t.Fatalf("free content too short: %d runes", length)
	}
	if !strings.Contains(content, reportpkg.DisclaimerText) {
		t.Fatal("free content must include disclaimer")
	}
	if !strings.Contains(content, "事业") {
		t.Fatal("free content should mention category")
	}
}

func TestBuildFullReportStructure(t *testing.T) {
	report := BuildFullReport(sampleInput())

	if report.Summary == "" || report.Overall == "" || report.CurrentState == "" {
		t.Fatal("full report core fields must not be empty")
	}
	if report.Opportunity == "" || report.Risk == "" || report.EmotionReminder == "" {
		t.Fatal("full report section fields must not be empty")
	}
	if len(report.ActionSteps) < 3 {
		t.Fatalf("expected at least 3 action steps, got %d", len(report.ActionSteps))
	}
	if len(report.ReflectionQuestions) < 3 {
		t.Fatalf("expected at least 3 reflection questions, got %d", len(report.ReflectionQuestions))
	}
	if report.Disclaimer != reportpkg.DisclaimerText {
		t.Fatal("full report disclaimer mismatch")
	}

	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal full report: %v", err)
	}
	if len(raw) < 500 {
		t.Fatalf("full report json too short: %d bytes", len(raw))
	}
}

func TestBuildFreeContentLengthRange(t *testing.T) {
	content := BuildFreeContent(sampleInput())
	length := utf8.RuneCountInString(content)
	if length < 200 {
		t.Fatalf("expected around 200-400 runes, got %d", length)
	}
}

func TestBuildDailyFreeContent(t *testing.T) {
	in := sampleInput()
	in.CategoryName = model.DailyFortuneCategoryName
	in.Question = model.DailyFortuneQuestion
	content := BuildFreeContent(in)
	if !strings.Contains(content, "今天的状态提醒") {
		t.Fatal("daily fortune free content should mention today's state reminder")
	}
	if strings.Contains(content, "事业决策") {
		t.Fatal("daily fortune should not mention career decision framing")
	}
}

func TestBuildDailyFullReport(t *testing.T) {
	in := sampleInput()
	in.CategoryName = model.DailyFortuneCategoryName
	in.Question = model.DailyFortuneQuestion
	report := BuildFullReport(in)
	if !strings.Contains(report.Overall, "今日") {
		t.Fatal("daily full report should be today-oriented")
	}
}
