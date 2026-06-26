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

var v2GoldenFixtures = []struct {
	name         string
	when         string
	category     string
	wantTerm     string
	wantDunType  string
	wantJu       int
	wantXunShou  string
	wantEmpty    []string
	wantZhiFu    string
	wantZhiShi   string
}{
	{
		name: "lichun_before_xiaohan", when: "2024-02-04 10:30", category: "general",
		wantTerm: "小寒", wantDunType: "yang", wantJu: 5,
		wantXunShou: "甲寅", wantEmpty: []string{"子", "丑"},
		wantZhiFu: "天蓬", wantZhiShi: "开门",
	},
	{
		name: "jingzhe_career", when: "2024-03-20 09:00", category: "career",
		wantTerm: "惊蛰", wantDunType: "yang", wantJu: 2,
		wantXunShou: "甲子", wantEmpty: []string{"戌", "亥"},
		wantZhiFu: "天冲", wantZhiShi: "生门",
	},
	{
		name: "xiazhi_day_study", when: "2024-06-21 09:00", category: "study",
		wantTerm: "芒种", wantDunType: "yin", wantJu: 2,
		wantXunShou: "甲午", wantEmpty: []string{"辰", "巳"},
		wantZhiFu: "天辅", wantZhiShi: "生门",
	},
	{
		name: "day_after_xiazhi_study", when: "2024-06-22 09:00", category: "study",
		wantTerm: "芒种", wantDunType: "yin", wantJu: 7,
		wantXunShou: "甲申", wantEmpty: []string{"午", "未"},
		wantZhiFu: "天辅", wantZhiShi: "惊门",
	},
	{
		name: "xiaoshu_relationship", when: "2024-08-07 15:00", category: "relationship",
		wantTerm: "小暑", wantDunType: "yin", wantJu: 8,
		wantXunShou: "甲辰", wantEmpty: []string{"寅", "卯"},
		wantZhiFu: "天柱", wantZhiShi: "开门",
	},
	{
		name: "bailu_decision", when: "2024-09-22 18:30", category: "decision",
		wantTerm: "白露", wantDunType: "yin", wantJu: 6,
		wantXunShou: "甲申", wantEmpty: []string{"午", "未"},
		wantZhiFu: "天辅", wantZhiShi: "死门",
	},
	{
		name: "before_dongzhi_general", when: "2024-12-21 23:10", category: "general",
		wantTerm: "大雪", wantDunType: "yin", wantJu: 8,
		wantXunShou: "甲午", wantEmpty: []string{"辰", "巳"},
		wantZhiFu: "天心", wantZhiShi: "开门",
	},
	{
		name: "dongzhi_day_general", when: "2024-12-22 09:00", category: "general",
		wantTerm: "大雪", wantDunType: "yang", wantJu: 2,
		wantXunShou: "甲辰", wantEmpty: []string{"寅", "卯"},
		wantZhiFu: "天禽", wantZhiShi: "生门",
	},
	{
		name: "xiaohan_career", when: "2025-02-03 11:30", category: "career",
		wantTerm: "小寒", wantDunType: "yang", wantJu: 6,
		wantXunShou: "甲寅", wantEmpty: []string{"子", "丑"},
		wantZhiFu: "天芮", wantZhiShi: "死门",
	},
	{
		name: "xiazhi_next_year_study", when: "2025-06-21 09:00", category: "study",
		wantTerm: "芒种", wantDunType: "yin", wantJu: 7,
		wantXunShou: "甲申", wantEmpty: []string{"午", "未"},
		wantZhiFu: "天禽", wantZhiShi: "惊门",
	},
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
	for _, tc := range v2GoldenFixtures {
		t.Run(tc.name, func(t *testing.T) {
			assertCalculateV2Golden(t, tc.when, tc.category, tc.wantTerm, tc.wantDunType, tc.wantJu, tc.wantXunShou, tc.wantEmpty, tc.wantZhiFu, tc.wantZhiShi)
		})
	}
}

func assertCalculateV2Golden(
	t *testing.T,
	when, category string,
	wantTerm, wantDunType string,
	wantJu int,
	wantXunShou string,
	wantEmpty []string,
	wantZhiFu, wantZhiShi string,
) {
	t.Helper()
	now := parseQimenV2Time(t, when)
	calc, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: category,
		Now:      now,
	})
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}

	if calc.CalendarBasis.SolarTerm != wantTerm {
		t.Fatalf("solar_term=%q want %q", calc.CalendarBasis.SolarTerm, wantTerm)
	}
	if calc.Dun.Type != wantDunType {
		t.Fatalf("dun.type=%q want %q", calc.Dun.Type, wantDunType)
	}
	if calc.Dun.Ju != wantJu {
		t.Fatalf("dun.ju=%d want %d", calc.Dun.Ju, wantJu)
	}
	if calc.Dun.Source != "poc_formula" {
		t.Fatalf("dun.source=%q", calc.Dun.Source)
	}
	if calc.Xun.XunShou != wantXunShou {
		t.Fatalf("xun_shou=%q want %q", calc.Xun.XunShou, wantXunShou)
	}
	if len(calc.Xun.EmptyBranches) != len(wantEmpty) {
		t.Fatalf("empty_branches=%v want %v", calc.Xun.EmptyBranches, wantEmpty)
	}
	for i := range wantEmpty {
		if calc.Xun.EmptyBranches[i] != wantEmpty[i] {
			t.Fatalf("empty_branches=%v want %v", calc.Xun.EmptyBranches, wantEmpty)
		}
	}
	if calc.Chief.ZhiFu != wantZhiFu {
		t.Fatalf("zhi_fu=%q want %q", calc.Chief.ZhiFu, wantZhiFu)
	}
	if calc.Chief.ZhiShi != wantZhiShi {
		t.Fatalf("zhi_shi=%q want %q", calc.Chief.ZhiShi, wantZhiShi)
	}

	assertCalculateV2PayloadShape(t, calc)
}

func assertCalculateV2PayloadShape(t *testing.T, calc qimen.CalculationResultV2) {
	t.Helper()
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{
		"algorithm_version", "calendar_basis", "dun", "xun", "chief", "palaces", "method_note", "limits",
	} {
		if _, ok := result[key]; !ok {
			t.Fatalf("missing payload key %q", key)
		}
	}
	if result["algorithm_version"] != qimen.AlgorithmVersionQimenV2POC {
		t.Fatalf("algorithm_version=%v", result["algorithm_version"])
	}

	basis, ok := result["calendar_basis"].(map[string]any)
	if !ok {
		t.Fatalf("calendar_basis type=%T", result["calendar_basis"])
	}
	if basis["jieqi_basis"] != "formula_approximation" {
		t.Fatalf("jieqi_basis=%v", basis["jieqi_basis"])
	}
	if basis["time_basis"] != "local_time" {
		t.Fatalf("time_basis=%v", basis["time_basis"])
	}
	if strings.TrimSpace(asString(basis["note"])) == "" {
		t.Fatalf("calendar_basis.note empty")
	}

	methodNote := asString(result["method_note"])
	if !strings.Contains(methodNote, "POC") {
		t.Fatalf("method_note should mention POC: %q", methodNote)
	}

	assertV2PalacesComplete(t, result["palaces"])
	assertV2Limits(t, result["limits"])
	assertV2PayloadCompliance(t, payload)
}

func assertV2PalacesComplete(t *testing.T, raw any) {
	t.Helper()
	palaces, ok := raw.([]any)
	if !ok || len(palaces) != 9 {
		t.Fatalf("palaces len=%d", len(palaces))
	}
	for i, item := range palaces {
		p, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("palace[%d] type", i)
		}
		for _, key := range []string{
			"index", "name", "earth_plate_stem", "heaven_plate_stem", "star", "door", "deity", "summary",
		} {
			if key == "index" {
				if _, ok := p[key].(float64); !ok {
					t.Fatalf("palace[%d] missing or invalid index", i)
				}
				continue
			}
			if strings.TrimSpace(asString(p[key])) == "" {
				t.Fatalf("palace[%d] missing %s", i, key)
			}
		}
		if int(p["index"].(float64)) != i+1 {
			t.Fatalf("palace[%d] index=%v want %d", i, p["index"], i+1)
		}
		if p["door"] == "—" && i != 4 {
			t.Fatalf("palace[%d] door should not be placeholder except center", i)
		}
		if i == 4 && p["door"] != "—" {
			t.Fatalf("center palace door=%v want —", p["door"])
		}
		if p["name"] != "中五宫" && i == 4 {
			t.Fatalf("center palace name=%v", p["name"])
		}
		if asString(p["summary"]) != "用于结构化观察，不作强预测" {
			t.Fatalf("palace[%d] summary=%q", i, p["summary"])
		}
	}
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
	now := parseQimenV2Time(t, "2024-03-20 09:00")
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

func TestCalculateV2DongzhiBoundaryYinToYang(t *testing.T) {
	before, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "general",
		Now:      parseQimenV2Time(t, "2024-12-21 23:10"),
	})
	if err != nil {
		t.Fatalf("before dongzhi: %v", err)
	}
	after, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "general",
		Now:      parseQimenV2Time(t, "2024-12-22 09:00"),
	})
	if err != nil {
		t.Fatalf("dongzhi day: %v", err)
	}
	if before.Dun.Type != "yin" {
		t.Fatalf("before dongzhi dun=%q want yin", before.Dun.Type)
	}
	if after.Dun.Type != "yang" {
		t.Fatalf("dongzhi day dun=%q want yang", after.Dun.Type)
	}
	if before.Dun.Type == after.Dun.Type {
		t.Fatalf("expected dun type flip across dongzhi POC boundary")
	}
}

func TestCalculateV2XiazhiBoundaryYangToYin(t *testing.T) {
	before, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "study",
		Now:      parseQimenV2Time(t, "2024-06-20 09:00"),
	})
	if err != nil {
		t.Fatalf("before xiazhi: %v", err)
	}
	after, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "study",
		Now:      parseQimenV2Time(t, "2024-06-21 09:00"),
	})
	if err != nil {
		t.Fatalf("xiazhi day: %v", err)
	}
	if before.Dun.Type != "yang" {
		t.Fatalf("before xiazhi dun=%q want yang", before.Dun.Type)
	}
	if after.Dun.Type != "yin" {
		t.Fatalf("xiazhi day dun=%q want yin", after.Dun.Type)
	}
	if before.Dun.Type == after.Dun.Type {
		t.Fatalf("expected dun type flip across xiazhi POC boundary")
	}
}

func TestCalculateV2CategoryAffectsJuOrPalaces(t *testing.T) {
	when := parseQimenV2Time(t, "2024-08-07 15:00")
	general, err := qimen.CalculateV2(qimen.CalculateInputV2{Category: "general", Now: when})
	if err != nil {
		t.Fatalf("general: %v", err)
	}
	relationship, err := qimen.CalculateV2(qimen.CalculateInputV2{Category: "relationship", Now: when})
	if err != nil {
		t.Fatalf("relationship: %v", err)
	}
	if general.Dun.Ju == relationship.Dun.Ju && general.Xun.XunShou == relationship.Xun.XunShou {
		t.Fatalf("expected category to affect ju or xun at same moment")
	}
	if general.Palaces[0].Star == relationship.Palaces[0].Star &&
		general.Palaces[0].HeavenPlateStem == relationship.Palaces[0].HeavenPlateStem {
		t.Fatalf("expected category to affect palace rotation or stems")
	}
}

func TestQimenSimpleV1UnchangedByV2POC(t *testing.T) {
	now := parseQimenV2Time(t, "2024-03-20 09:00")
	calc, err := qimen.Calculate("我最近适合推进这个计划吗？", "career", now)
	if err != nil {
		t.Fatalf("Calculate v1: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	raw := string(payload)
	if !strings.Contains(raw, "qimen-simple-v1") {
		t.Fatalf("v1 algorithm_version missing: %s", raw)
	}
	if strings.Contains(raw, "palaces") {
		t.Fatalf("v1 payload must not contain palaces")
	}
}

func TestCalculateV2ResultPayloadFields(t *testing.T) {
	calc, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "decision",
		Now:      parseQimenV2Time(t, "2024-09-22 18:30"),
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
	for _, tc := range v2GoldenFixtures {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}
