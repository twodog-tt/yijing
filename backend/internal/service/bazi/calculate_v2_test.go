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
			calc, err := bazi.CalculateV2(tc.date, tc.hour, tc.unknown)
			if err != nil {
				t.Fatalf("CalculateV2: %v", err)
			}
			if calc.PillarsV2.Year != tc.wantYear || calc.PillarsV2.Month != tc.wantMonth || calc.PillarsV2.Day != tc.wantDay {
				t.Fatalf("pillars mismatch: got %+v", calc.PillarsV2)
			}
			if calc.DayMaster != tc.wantDayMaster {
				t.Fatalf("day_master=%q want %q", calc.DayMaster, tc.wantDayMaster)
			}
			if tc.wantHourPresent && calc.PillarsV2.Hour != tc.wantHour {
				t.Fatalf("hour=%q want %q", calc.PillarsV2.Hour, tc.wantHour)
			}
			if !tc.wantHourPresent && calc.PillarsV2.Hour != "" {
				t.Fatalf("hour should be empty, got %q", calc.PillarsV2.Hour)
			}
		})
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
