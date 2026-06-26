package qimen_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

var v2ForbiddenPhrases = []string{
	"精准预测", "必成", "必败", "大吉", "大凶", "必发财", "必复合",
	"改运", "化灾", "转运", "投资建议", "医疗建议", "法律建议", "赌博建议", "军事行动建议",
}

func parseQimenV2Time(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02 15:04", value, clock.Location())
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
}

func TestCalculateV2GoldenFixtures(t *testing.T) {
	cases := []struct {
		name     string
		when     string
		category string
	}{
		{name: "career_spring", when: "2024-03-12 10:30", category: "career"},
		{name: "relationship_summer", when: "2024-08-20 15:00", category: "relationship"},
		{name: "decision_winter", when: "2024-12-05 23:10", category: "decision"},
		{name: "general_lichun", when: "2025-02-03 11:30", category: "general"},
		{name: "study_xiazhi", when: "2025-06-21 09:00", category: "study"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertCalculateV2Shape(t, tc.when, tc.category)
		})
	}
}

func assertCalculateV2Shape(t *testing.T, when, category string) {
	t.Helper()
	now := parseQimenV2Time(t, when)
	calc, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: category,
		Now:      now,
	})
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}

	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["algorithm_version"] != qimen.AlgorithmVersionQimenV2POC {
		t.Fatalf("algorithm_version=%v", result["algorithm_version"])
	}

	palaces, ok := result["palaces"].([]any)
	if !ok || len(palaces) != 9 {
		t.Fatalf("palaces len=%d", len(palaces))
	}
	for i, item := range palaces {
		p, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("palace[%d] type", i)
		}
		for _, key := range []string{"name", "star", "door", "deity", "earth_plate_stem", "heaven_plate_stem"} {
			if strings.TrimSpace(asString(p[key])) == "" {
				t.Fatalf("palace[%d] missing %s", i, key)
			}
		}
		if p["door"] == "—" && i != 4 {
			t.Fatalf("palace[%d] door should not be placeholder except center", i)
		}
	}

	assertV2Limits(t, result["limits"])
	assertV2PayloadCompliance(t, payload)
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func assertV2Limits(t *testing.T, raw any) {
	t.Helper()
	limits, ok := raw.([]any)
	if !ok {
		t.Fatalf("limits type=%T", raw)
	}
	text := strings.Join(toStrings(limits), " ")
	for _, want := range []string{"不提供精准预测", "不构成现实决策依据"} {
		if !strings.Contains(text, want) {
			t.Fatalf("limits missing %q: %v", want, limits)
		}
	}
}

func toStrings(items []any) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, asString(item))
	}
	return out
}

func assertV2PayloadCompliance(t *testing.T, payload []byte) {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	delete(result, "limits")
	if meta, ok := result["calculation_meta"].(map[string]any); ok {
		delete(meta, "limits")
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	bodyText := string(body)
	for _, phrase := range v2ForbiddenPhrases {
		if strings.Contains(bodyText, phrase) {
			t.Fatalf("forbidden phrase %q in conclusion body", phrase)
		}
	}
}

func TestCalculateV2StableForSameInput(t *testing.T) {
	now := parseQimenV2Time(t, "2024-03-12 10:30")
	input := qimen.CalculateInputV2{Category: "career", Now: now}

	first, err := qimen.CalculateV2(input)
	if err != nil {
		t.Fatalf("first CalculateV2: %v", err)
	}
	second, err := qimen.CalculateV2(input)
	if err != nil {
		t.Fatalf("second CalculateV2: %v", err)
	}

	firstPayload, err := first.ResultPayload()
	if err != nil {
		t.Fatalf("first ResultPayload: %v", err)
	}
	secondPayload, err := second.ResultPayload()
	if err != nil {
		t.Fatalf("second ResultPayload: %v", err)
	}
	if string(firstPayload) != string(secondPayload) {
		t.Fatalf("payload not stable:\nfirst=%s\nsecond=%s", firstPayload, secondPayload)
	}
}

func TestCalculateV2DifferentMomentsChangeBasisOrDun(t *testing.T) {
	summer, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "general",
		Now:      parseQimenV2Time(t, "2024-08-20 15:00"),
	})
	if err != nil {
		t.Fatalf("summer CalculateV2: %v", err)
	}
	winter, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "general",
		Now:      parseQimenV2Time(t, "2024-12-25 23:10"),
	})
	if err != nil {
		t.Fatalf("winter CalculateV2: %v", err)
	}

	same := summer.CalendarBasis.SolarTerm == winter.CalendarBasis.SolarTerm &&
		summer.Dun.Type == winter.Dun.Type &&
		summer.Dun.Ju == winter.Dun.Ju
	if same {
		t.Fatalf("expected summer/winter to differ in calendar or dun: summer=%+v winter=%+v", summer.CalendarBasis, summer.Dun)
	}
	if summer.Dun.Type != "yin" {
		t.Fatalf("summer dun type=%q want yin", summer.Dun.Type)
	}
	if winter.Dun.Type != "yang" {
		t.Fatalf("winter dun type=%q want yang", winter.Dun.Type)
	}
}

func TestQimenSimpleV1UnchangedByV2POC(t *testing.T) {
	now := parseQimenV2Time(t, "2024-03-12 10:30")
	calc, err := qimen.Calculate("我最近适合推进这个计划吗？", "career", now)
	if err != nil {
		t.Fatalf("Calculate v1: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	raw := string(payload)
	if !strings.Contains(raw, `"algorithm_version":"qimen-simple-v1"`) && !strings.Contains(raw, `"algorithm_version": "qimen-simple-v1"`) {
		if !strings.Contains(raw, "qimen-simple-v1") {
			t.Fatalf("v1 algorithm_version missing: %s", raw)
		}
	}
	if strings.Contains(raw, "palaces") {
		t.Fatalf("v1 payload must not contain palaces")
	}
}

func TestCalculateV2ResultPayloadFields(t *testing.T) {
	calc, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "decision",
		Now:      parseQimenV2Time(t, "2024-12-05 23:10"),
	})
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	if calc.MethodNote != qimen.MethodNoteV2 {
		t.Fatalf("method_note=%q", calc.MethodNote)
	}
	if calc.Chief.ZhiFu == "" || calc.Chief.ZhiShi == "" {
		t.Fatalf("chief=%+v", calc.Chief)
	}
	if calc.Xun.XunShou == "" || len(calc.Xun.EmptyBranches) != 2 {
		t.Fatalf("xun=%+v", calc.Xun)
	}
	if calc.CalendarBasis.JieqiBasis != "formula_approximation" || calc.CalendarBasis.TimeBasis != "local_time" {
		t.Fatalf("calendar_basis=%+v", calc.CalendarBasis)
	}
	if calc.Dun.Source != "poc_formula" {
		t.Fatalf("dun source=%q", calc.Dun.Source)
	}
}

func TestCalculateV2ForbiddenWordsAcrossFixtures(t *testing.T) {
	cases := []struct {
		when     string
		category string
	}{
		{"2024-03-12 10:30", "career"},
		{"2024-08-20 15:00", "relationship"},
		{"2024-12-05 23:10", "decision"},
		{"2025-02-03 11:30", "general"},
		{"2025-06-21 09:00", "study"},
	}
	for _, tc := range cases {
		calc, err := qimen.CalculateV2(qimen.CalculateInputV2{
			Category: tc.category,
			Now:      parseQimenV2Time(t, tc.when),
		})
		if err != nil {
			t.Fatalf("CalculateV2: %v", err)
		}
		payload, err := calc.ResultPayload()
		if err != nil {
			t.Fatalf("ResultPayload: %v", err)
		}
		assertV2PayloadCompliance(t, payload)
	}
}
