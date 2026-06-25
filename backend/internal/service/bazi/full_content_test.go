package bazi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildFullContentIncludesRequiredSections(t *testing.T) {
	payload, err := CalculationResult{
		Pillars:           Pillars{Year: "乙亥", Month: "己丑", Day: "甲子", Hour: "甲子"},
		DayMaster:         "甲",
		FiveElements:      FiveElements{Wood: 3, Fire: 1, Earth: 2, Metal: 1, Water: 1},
		ReflectionFocus:   "留意节奏与休息。",
		ActionSuggestions: []string{"固定一项日常练习。"},
	}.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}

	content, err := BuildFullContent(payload, "免费解读摘要")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}

	required := []string{
		"【1. 简化干支示意】",
		"【2. 五行倾向展开】",
		"【3. 日主与行动风格】",
		"【4. 自我反思问题】",
		"【5. 近期行动建议】",
		"【6. 适合记录的观察方向】",
		"【7. 免责声明】",
		fullReportDisclaimer,
	}
	for _, snippet := range required {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected full content to contain %q", snippet)
		}
	}

	forbidden := []string{"精准算命", "保证发财", "改运化解", "医疗建议"}
	for _, word := range forbidden {
		if strings.Contains(content, word) {
			t.Fatalf("full content must not contain %q", word)
		}
	}
}

func TestBuildFullContentHourUnknownNote(t *testing.T) {
	payload := json.RawMessage(`{
		"day_master":"甲",
		"pillars":{"year":"乙亥","month":"己丑","day":"甲子"},
		"five_elements":{"wood":2,"fire":1,"earth":1,"metal":1,"water":1},
		"reflection_focus":"从年月日三柱出发观察。",
		"action_suggestions":["记录一周节奏。"]
	}`)

	content, err := BuildFullContent(payload, "")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	if !strings.Contains(content, "时辰未知，本次不生成时柱") {
		t.Fatalf("expected hour-unknown note in full content")
	}
}

func TestBuildFullContentRejectsInvalidPayload(t *testing.T) {
	_, err := BuildFullContent(json.RawMessage(`{"pillars":{}}`), "")
	if err == nil {
		t.Fatalf("expected error for invalid payload")
	}
}
