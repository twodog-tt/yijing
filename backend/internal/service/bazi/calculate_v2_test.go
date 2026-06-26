package bazi_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/service/bazi"
)

func TestBaziV2GoldenFixtures(t *testing.T) {
	cases := []struct {
		name            string
		date            string
		hour            string
		unknown         bool
		wantYear        string
		wantMonth       string
		wantDay         string
		wantHour        string
		wantHourPresent bool
		wantDayMaster   string
	}{
		{
			name: "lichun_before_still_previous_year",
			date: "2024-02-03", hour: "wu", unknown: false,
			wantYear: "癸卯", wantMonth: "癸丑", wantDay: "丁酉", wantHour: "丙午", wantHourPresent: true, wantDayMaster: "丁",
		},
		{
			name: "lichun_after_new_year_pillar",
			date: "2024-02-05", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "丙寅", wantDay: "己亥", wantHour: "庚午", wantHourPresent: true, wantDayMaster: "己",
		},
		{
			name: "gregorian_jan_before_lichun",
			date: "1995-01-01", hour: "zi", unknown: false,
			wantYear: "甲戌", wantMonth: "甲子", wantDay: "壬辰", wantHour: "庚子", wantHourPresent: true, wantDayMaster: "壬",
		},
		{
			name: "jingzhe_before_yin_month",
			date: "2024-03-05", hour: "chen", unknown: false,
			wantYear: "甲辰", wantMonth: "丙寅", wantDay: "戊辰", wantHour: "丙辰", wantHourPresent: true, wantDayMaster: "戊",
		},
		{
			name: "jingzhe_after_mao_month",
			date: "2024-03-07", hour: "chen", unknown: false,
			wantYear: "甲辰", wantMonth: "丁卯", wantDay: "庚午", wantHour: "庚辰", wantHourPresent: true, wantDayMaster: "庚",
		},
		{
			name: "qingming_before_mao",
			date: "2024-04-04", hour: "si", unknown: false,
			wantYear: "甲辰", wantMonth: "丁卯", wantDay: "戊戌", wantHour: "丁巳", wantHourPresent: true, wantDayMaster: "戊",
		},
		{
			name: "qingming_after_chen",
			date: "2024-04-06", hour: "si", unknown: false,
			wantYear: "甲辰", wantMonth: "戊辰", wantDay: "庚子", wantHour: "辛巳", wantHourPresent: true, wantDayMaster: "庚",
		},
		{
			name: "xiaohan_chou_month",
			date: "2025-01-06", hour: "hai", unknown: false,
			wantYear: "甲辰", wantMonth: "乙丑", wantDay: "乙亥", wantHour: "丁亥", wantHourPresent: true, wantDayMaster: "乙",
		},
		{
			name: "daxue_zi_month",
			date: "2024-12-09", hour: "zi", unknown: false,
			wantYear: "甲辰", wantMonth: "甲子", wantDay: "丁未", wantHour: "庚子", wantHourPresent: true, wantDayMaster: "丁",
		},
		{
			name: "unknown_hour_no_hour_pillar",
			date: "2024-02-05", hour: "", unknown: true,
			wantYear: "甲辰", wantMonth: "丙寅", wantDay: "己亥", wantHourPresent: false, wantDayMaster: "己",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertBaziV2Golden(t, tc.date, tc.hour, tc.unknown, tc.wantYear, tc.wantMonth, tc.wantDay, tc.wantHour, tc.wantHourPresent, tc.wantDayMaster)
		})
	}
}

func TestBaziV2SolarTermBoundaryFixtures(t *testing.T) {
	cases := []struct {
		name            string
		date            string
		hour            string
		unknown         bool
		wantYear        string
		wantMonth       string
		wantDay         string
		wantHour        string
		wantHourPresent bool
		wantDayMaster   string
	}{
		{
			name: "lichun_day_before_1995",
			date: "1995-02-03", hour: "zi", unknown: false,
			wantYear: "甲戌", wantMonth: "乙丑", wantDay: "乙丑", wantHour: "丙子", wantHourPresent: true, wantDayMaster: "乙",
		},
		{
			name: "lichun_formula_day_2024",
			date: "2024-02-04", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "丙寅", wantDay: "戊戌", wantHour: "戊午", wantHourPresent: true, wantDayMaster: "戊",
		},
		{
			name: "xiaoshu_before_wu_month",
			date: "2024-07-07", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "庚午", wantDay: "壬申", wantHour: "丙午", wantHourPresent: true, wantDayMaster: "壬",
		},
		{
			name: "xiaoshu_after_wei_month",
			date: "2024-07-09", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "辛未", wantDay: "甲戌", wantHour: "庚午", wantHourPresent: true, wantDayMaster: "甲",
		},
		{
			name: "liqiu_before_wei_month",
			date: "2024-08-07", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "辛未", wantDay: "癸卯", wantHour: "戊午", wantHourPresent: true, wantDayMaster: "癸",
		},
		{
			name: "liqiu_after_shen_month",
			date: "2024-08-09", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "壬申", wantDay: "乙巳", wantHour: "壬午", wantHourPresent: true, wantDayMaster: "乙",
		},
		{
			name: "bailu_before_shen_month",
			date: "2024-09-08", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "壬申", wantDay: "乙亥", wantHour: "壬午", wantHourPresent: true, wantDayMaster: "乙",
		},
		{
			name: "bailu_after_you_month",
			date: "2024-09-10", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "癸酉", wantDay: "丁丑", wantHour: "丙午", wantHourPresent: true, wantDayMaster: "丁",
		},
		{
			name: "hanlu_before_you_month",
			date: "2024-10-08", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "癸酉", wantDay: "乙巳", wantHour: "壬午", wantHourPresent: true, wantDayMaster: "乙",
		},
		{
			name: "hanlu_after_xu_month",
			date: "2024-10-10", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "甲戌", wantDay: "丁未", wantHour: "丙午", wantHourPresent: true, wantDayMaster: "丁",
		},
		{
			name: "lidong_before_xu_month",
			date: "2024-11-07", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "甲戌", wantDay: "乙亥", wantHour: "壬午", wantHourPresent: true, wantDayMaster: "乙",
		},
		{
			name: "lidong_after_hai_month",
			date: "2024-11-09", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "乙亥", wantDay: "丁丑", wantHour: "丙午", wantHourPresent: true, wantDayMaster: "丁",
		},
		{
			name: "daxue_before_hai_month",
			date: "2024-12-06", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "乙亥", wantDay: "甲辰", wantHour: "庚午", wantHourPresent: true, wantDayMaster: "甲",
		},
		{
			name: "daxue_after_zi_month",
			date: "2024-12-08", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "甲子", wantDay: "丙午", wantHour: "甲午", wantHourPresent: true, wantDayMaster: "丙",
		},
		{
			name: "xiaohan_before_zi_month",
			date: "2025-01-04", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "甲子", wantDay: "癸酉", wantHour: "戊午", wantHourPresent: true, wantDayMaster: "癸",
		},
		{
			name: "xiaohan_after_chou_month",
			date: "2025-01-07", hour: "wu", unknown: false,
			wantYear: "甲辰", wantMonth: "乙丑", wantDay: "丙子", wantHour: "甲午", wantHourPresent: true, wantDayMaster: "丙",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertBaziV2Golden(t, tc.date, tc.hour, tc.unknown, tc.wantYear, tc.wantMonth, tc.wantDay, tc.wantHour, tc.wantHourPresent, tc.wantDayMaster)
		})
	}
}

func assertBaziV2Golden(
	t *testing.T,
	date, hour string,
	unknown bool,
	wantYear, wantMonth, wantDay, wantHour string,
	wantHourPresent bool,
	wantDayMaster string,
) {
	t.Helper()
	calc, err := bazi.CalculateV2(date, hour, unknown)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	if calc.PillarsV2.Year != wantYear || calc.PillarsV2.Month != wantMonth || calc.PillarsV2.Day != wantDay {
		t.Fatalf("pillars mismatch: got %+v", calc.PillarsV2)
	}
	if calc.DayMaster != wantDayMaster {
		t.Fatalf("day_master=%q want %q", calc.DayMaster, wantDayMaster)
	}
	if wantHourPresent && calc.PillarsV2.Hour != wantHour {
		t.Fatalf("hour=%q want %q", calc.PillarsV2.Hour, wantHour)
	}
	if !wantHourPresent && calc.PillarsV2.Hour != "" {
		t.Fatalf("hour should be empty, got %q", calc.PillarsV2.Hour)
	}
}

func TestBaziV2LimitsDocumentSolarTermApproximation(t *testing.T) {
	calc, err := bazi.CalculateV2("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	found := false
	for _, limit := range calc.Limits {
		if strings.Contains(limit, "节令时刻按本地正午近似") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected solar-term approximation limit, got %v", calc.Limits)
	}
	if calc.CalendarBasis.Note == "" || !strings.Contains(calc.CalendarBasis.Note, "真太阳时延后") {
		t.Fatalf("calendar basis note should mention deferred true solar time: %+v", calc.CalendarBasis)
	}
}

func TestBaziSimpleV1UnchangedByV2POC(t *testing.T) {
	calc, err := bazi.Calculate("1995-01-01", "zi", false)
	if err != nil {
		t.Fatalf("Calculate v1: %v", err)
	}
	if calc.Pillars.Year != "乙亥" || calc.Pillars.Month != "丁丑" {
		t.Fatalf("v1 pillars changed unexpectedly: %+v", calc.Pillars)
	}
}

func TestBaziV2ResultPayloadShape(t *testing.T) {
	calc, err := bazi.CalculateV2("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	raw := string(payload)
	for _, forbidden := range []string{"birth_date", "session_key", "input_payload", "prompt"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("result_payload must not contain %q", forbidden)
		}
	}
	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["algorithm_version"] != bazi.AlgorithmVersionBaziV2POC {
		t.Fatalf("algorithm_version=%v", result["algorithm_version"])
	}
	basis := result["calendar_basis"].(map[string]any)
	if basis["year_boundary"] != "lichun" || basis["month_boundary"] != "solar_terms_jie" {
		t.Fatalf("calendar_basis=%v", basis)
	}
	if basis["true_solar_time"] != false {
		t.Fatalf("true_solar_time should be false")
	}
	if _, ok := result["pillars_v2"]; !ok {
		t.Fatalf("missing pillars_v2")
	}
}

func TestBaziV2DayPillarUsesV1Epoch(t *testing.T) {
	v1, err := bazi.Calculate("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("Calculate v1: %v", err)
	}
	v2, err := bazi.CalculateV2("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	if v1.Pillars.Day != v2.PillarsV2.Day {
		t.Fatalf("day pillar should match v1 epoch: v1=%q v2=%q", v1.Pillars.Day, v2.PillarsV2.Day)
	}
}

func TestBuildPromptInputDoesNotExposeBirthDate(t *testing.T) {
	calc, err := bazi.CalculateV2("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	// v2 POC does not wire prompt yet; ensure payload never carries birth_date.
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	if strings.Contains(string(payload), "2024-02-05") {
		t.Fatalf("result_payload must not contain raw birth date")
	}
}
