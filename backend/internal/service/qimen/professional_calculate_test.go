package qimen_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

func TestCalculateProfessionalPreviewStructure(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	result, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{
		Category: "career",
		Now:      when,
	})
	if err != nil {
		t.Fatalf("CalculateProfessionalPreview: %v", err)
	}
	payload, err := result.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(payload, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if obj["algorithm_version"] != qimen.AlgorithmVersionQimenV2Professional {
		t.Fatalf("algorithm_version=%v", obj["algorithm_version"])
	}
	assertProfessionalPreviewPayload(t, obj, when)
}

func TestCalculateProfessionalPreviewFixtures(t *testing.T) {
	provider := qimen.FormulaSolarTermProvider{}
	for _, plan := range qimen.ProfessionalFixturePlans {
		t.Run(plan.Name, func(t *testing.T) {
			when, err := time.ParseInLocation("2006-01-02 15:04", plan.When, clock.Location())
			if err != nil {
				t.Fatalf("parse when: %v", err)
			}
			result, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{
				Category: plan.Category,
				Now:      when,
				Provider: provider,
			})
			if err != nil {
				t.Fatalf("CalculateProfessionalPreview: %v", err)
			}
			payload, err := result.ResultPayload()
			if err != nil {
				t.Fatalf("ResultPayload: %v", err)
			}
			var obj map[string]any
			if err := json.Unmarshal(payload, &obj); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			assertProfessionalPreviewPayload(t, obj, when)

			dun := obj["dun"].(map[string]any)
			expected := qimen.ResolveProfessionalDun(when, provider)
			if dun["type"] != expected.Type {
				t.Fatalf("dun.type=%v want %q (provider 夏至/冬至边界)", dun["type"], expected.Type)
			}
		})
	}
}

func TestCalculateProfessionalPreviewChiefAndPalaces(t *testing.T) {
	result, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{
		Category: "general",
		Now:      time.Date(2024, 12, 22, 0, 30, 0, 0, clock.Location()),
	})
	if err != nil {
		t.Fatalf("CalculateProfessionalPreview: %v", err)
	}
	if result.Chief.ZhiFu == "professional_pending" || result.Chief.ZhiShi == "professional_pending" {
		t.Fatalf("chief should not be pending: %+v", result.Chief)
	}
	if len(result.Palaces) != 9 {
		t.Fatalf("palaces len=%d want 9", len(result.Palaces))
	}
	if result.Dun.Ju < 1 || result.Dun.Ju > 9 {
		t.Fatalf("ju should be 1-9, got %d", result.Dun.Ju)
	}
	if result.Dun.Source != qimen.JuBasisTwentyFourTermsChaiBu() {
		t.Fatalf("dun source/basis=%q", result.Dun.Source)
	}
	if result.Dun.Method != qimen.DunMethodChaiBu {
		t.Fatalf("method=%q want chai_bu", result.Dun.Method)
	}
}

func TestCalculateProfessionalPreviewPalaceLayoutFixtures(t *testing.T) {
	provider := qimen.FormulaSolarTermProvider{}
	fixtures := []struct {
		when string
		cat  string
	}{
		{"2024-03-20 09:00", "career"},
		{"2024-06-20 23:30", "study"},
		{"2024-06-21 00:30", "study"},
		{"2024-12-21 23:10", "general"},
		{"2024-12-22 00:30", "general"},
		{"2025-02-03 11:30", "career"},
	}
	for _, fx := range fixtures {
		t.Run(fx.when, func(t *testing.T) {
			when, _ := time.ParseInLocation("2006-01-02 15:04", fx.when, clock.Location())
			result, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{
				Category: fx.cat, Now: when, Provider: provider,
			})
			if err != nil {
				t.Fatalf("preview: %v", err)
			}
			assertProfessionalPalaces(t, result)
		})
	}
}

func TestCalculateProfessionalPreviewPalacesCategoryIndependent(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	var firstPalace string
	for i, cat := range []string{"general", "career", "study"} {
		result, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{Category: cat, Now: when})
		if err != nil {
			t.Fatalf("category %q: %v", cat, err)
		}
		if i == 0 {
			firstPalace = result.Palaces[0].EarthPlateStem
			continue
		}
		if result.Palaces[0].EarthPlateStem != firstPalace {
			t.Fatalf("category %q changed layout", cat)
		}
	}
}

func assertProfessionalPalaces(t *testing.T, result *qimen.CalculationResultV2Professional) {
	t.Helper()
	if result.Chief.ZhiFuPalace == "" || result.Chief.ZhiShiPalace == "" {
		t.Fatalf("chief incomplete: %+v", result.Chief)
	}
	if len(result.Palaces) != 9 {
		t.Fatalf("palaces len=%d", len(result.Palaces))
	}
	for _, p := range result.Palaces {
		if p.EarthPlateStem == "" || p.HeavenPlateStem == "" || p.Star == "" || p.Deity == "" {
			t.Fatalf("palace %d missing fields: %+v", p.Index, p)
		}
		if p.Index == 5 {
			if p.Star != "天禽" || p.Door != "—" {
				t.Fatalf("center palace invalid: %+v", p)
			}
		} else if p.Door == "" {
			t.Fatalf("palace %d door empty", p.Index)
		}
	}
}

func TestCalculateProfessionalPreviewTwentyFourTermFixtures(t *testing.T) {
	provider := qimen.FormulaSolarTermProvider{}
	fixtures := []struct {
		when string
		cat  string
	}{
		{"2024-01-06 09:00", "general"},
		{"2024-02-04 10:30", "general"},
		{"2024-03-20 09:00", "career"},
		{"2024-06-20 23:30", "study"},
		{"2024-06-21 00:30", "study"},
		{"2024-08-07 15:00", "relationship"},
		{"2024-09-22 18:30", "decision"},
		{"2024-12-21 23:10", "general"},
		{"2024-12-22 00:30", "general"},
		{"2025-01-05 09:00", "general"},
		{"2025-02-03 11:30", "career"},
	}
	for _, fx := range fixtures {
		t.Run(fx.when, func(t *testing.T) {
			when, err := time.ParseInLocation("2006-01-02 15:04", fx.when, clock.Location())
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			result, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{
				Category: fx.cat, Now: when, Provider: provider,
			})
			if err != nil {
				t.Fatalf("preview: %v", err)
			}
			if result.CalendarBasis.SolarTerm == "" {
				t.Fatal("calendar solar_term empty")
			}
			if result.Dun.Ju < 1 || result.Dun.Ju > 9 {
				t.Fatalf("ju=%d", result.Dun.Ju)
			}
			if result.Dun.Source != qimen.JuBasisTwentyFourTermsChaiBu() {
				t.Fatalf("basis=%q", result.Dun.Source)
			}
		})
	}
}

func TestCalculateProfessionalPreviewJuIndependentOfCategory(t *testing.T) {
	when := time.Date(2024, 8, 7, 15, 0, 0, 0, clock.Location())
	var firstJu float64
	var firstYuan string
	for i, cat := range []string{"general", "career", "relationship", "decision"} {
		result, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{Category: cat, Now: when})
		if err != nil {
			t.Fatalf("category %q: %v", cat, err)
		}
		if i == 0 {
			firstJu = float64(result.Dun.Ju)
			firstYuan = result.Dun.Yuan
			continue
		}
		if float64(result.Dun.Ju) != firstJu || result.Dun.Yuan != firstYuan {
			t.Fatalf("category %q changed ju/yuan: ju=%d yuan=%q (want ju=%v yuan=%q)", cat, result.Dun.Ju, result.Dun.Yuan, firstJu, firstYuan)
		}
	}
}

func TestCalculateProfessionalPreviewDoesNotAffectPOC(t *testing.T) {
	when := time.Date(2024, 6, 21, 0, 30, 0, 0, clock.Location())
	poc, err := qimen.CalculateV2(qimen.CalculateInputV2{Category: "study", Now: when})
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	if poc.Dun.Type != "yin" {
		t.Fatalf("POC dun should remain yin on 2024-06-21 per gregorian rule")
	}
	pro, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{Category: "study", Now: when})
	if err != nil {
		t.Fatalf("CalculateProfessionalPreview: %v", err)
	}
	if pro.Dun.Method != qimen.DunMethodChaiBu {
		t.Fatalf("professional method=%q", pro.Dun.Method)
	}
}

func assertProfessionalPreviewPayload(t *testing.T, obj map[string]any, when time.Time) {
	t.Helper()
	basis := obj["calendar_basis"].(map[string]any)
	if basis["solar_term"] == "" || basis["solar_term_time"] == "" {
		t.Fatalf("calendar_basis incomplete: %v", basis)
	}
	if basis["jieqi_basis"] != "formula_approximation" {
		t.Fatalf("jieqi_basis=%v", basis["jieqi_basis"])
	}

	dun := obj["dun"].(map[string]any)
	if dun["type"] != "yang" && dun["type"] != "yin" {
		t.Fatalf("dun.type=%v", dun["type"])
	}
	if dun["method"] != qimen.DunMethodChaiBu {
		t.Fatalf("dun.method=%v", dun["method"])
	}
	ju, ok := dun["ju"].(float64)
	if !ok || ju < 1 || ju > 9 {
		t.Fatalf("dun.ju=%v want 1-9", dun["ju"])
	}
	yuan, _ := dun["yuan"].(string)
	if yuan != qimen.DunYuanUpper && yuan != qimen.DunYuanMiddle && yuan != qimen.DunYuanLower {
		t.Fatalf("dun.yuan=%v", dun["yuan"])
	}
	note, _ := dun["note"].(string)
	if note == "" || !strings.Contains(note, "第一版") {
		t.Fatalf("dun.note should mention first version approximation: %q", note)
	}
	if src, _ := dun["source"].(string); src != qimen.JuBasisTwentyFourTermsChaiBu() {
		t.Fatalf("dun.source/basis=%v", dun["source"])
	}

	gz := obj["ganzhi"].(map[string]any)
	for _, key := range []string{"year", "month", "day", "hour"} {
		if gz[key] == "" {
			t.Fatalf("ganzhi.%s empty", key)
		}
	}

	xun := obj["xun"].(map[string]any)
	if xun["xun_shou"] == "" {
		t.Fatal("xun.xun_shou empty")
	}
	if _, ok := xun["empty_branches"]; !ok {
		t.Fatal("xun.empty_branches missing")
	}

	chief := obj["chief"].(map[string]any)
	if chief["zhi_fu"] == "professional_pending" || chief["zhi_shi"] == "professional_pending" {
		t.Fatalf("chief should not be pending: %v", chief)
	}
	if chief["zhi_fu_palace"] == "" || chief["zhi_shi_palace"] == "" {
		t.Fatalf("chief missing palace fields: %v", chief)
	}

	palaces, ok := obj["palaces"].([]any)
	if !ok || len(palaces) != 9 {
		t.Fatalf("palaces want 9, got %v", obj["palaces"])
	}

	limits, ok := obj["limits"].([]any)
	if !ok || len(limits) == 0 {
		t.Fatal("limits missing")
	}
	limitText := strings.Join(toStringSlice(limits), " ")
	for _, phrase := range []string{"不提供精准预测", "不构成现实决策依据"} {
		if !strings.Contains(limitText, phrase) {
			t.Fatalf("limits missing %q: %s", phrase, limitText)
		}
	}

	_ = when
}

func toStringSlice(items []any) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.(string))
	}
	return out
}
