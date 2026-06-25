package bazi_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/bazi"
)

// Golden fixtures validate bazi-simple-v1 simplified rule stability only, not professional accuracy.
func TestBaziSimpleV1GoldenFixtures(t *testing.T) {
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
			name: "known_hour", date: "1995-01-01", hour: "zi", unknown: false,
			wantYear: "乙亥", wantMonth: "丁丑", wantDay: "壬辰", wantHour: "庚子", wantHourPresent: true, wantDayMaster: "壬",
		},
		{
			name: "unknown_hour", date: "1995-01-01", hour: "", unknown: true,
			wantYear: "乙亥", wantMonth: "丁丑", wantDay: "壬辰", wantHourPresent: false, wantDayMaster: "壬",
		},
		{
			name: "year_boundary_1900", date: "1900-01-01", hour: "zi", unknown: false,
			wantYear: "庚子", wantMonth: "丁丑", wantDay: "甲戌", wantHour: "甲子", wantHourPresent: true, wantDayMaster: "甲",
		},
		{
			name: "year_boundary_1999", date: "1999-12-31", hour: "hai", unknown: false,
			wantYear: "己卯", wantMonth: "甲子", wantDay: "丁巳", wantHour: "辛亥", wantHourPresent: true, wantDayMaster: "丁",
		},
		{
			name: "month_boundary", date: "1995-01-31", hour: "wu", unknown: false,
			wantYear: "乙亥", wantMonth: "丁丑", wantDay: "壬戌", wantHour: "丙午", wantHourPresent: true, wantDayMaster: "壬",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calc, err := bazi.Calculate(tc.date, tc.hour, tc.unknown)
			if err != nil {
				t.Fatalf("Calculate: %v", err)
			}
			if calc.Pillars.Year != tc.wantYear || calc.Pillars.Month != tc.wantMonth || calc.Pillars.Day != tc.wantDay {
				t.Fatalf("pillars mismatch: got %v", calc.Pillars)
			}
			if calc.DayMaster != tc.wantDayMaster {
				t.Fatalf("day_master=%q want %q", calc.DayMaster, tc.wantDayMaster)
			}
			if tc.wantHourPresent && calc.Pillars.Hour != tc.wantHour {
				t.Fatalf("hour=%q want %q", calc.Pillars.Hour, tc.wantHour)
			}

			payload, err := calc.ResultPayload()
			if err != nil {
				t.Fatalf("ResultPayload: %v", err)
			}
			var result map[string]any
			if err := json.Unmarshal(payload, &result); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if _, ok := result["birth"]; ok {
				t.Fatalf("result_payload must not contain birth object")
			}
			pillars := result["pillars"].(map[string]any)
			if tc.wantHourPresent {
				if pillars["hour"] != tc.wantHour {
					t.Fatalf("result hour=%v want %q", pillars["hour"], tc.wantHour)
				}
			} else if _, ok := pillars["hour"]; ok {
				t.Fatalf("result pillars must not contain hour key")
			}
		})
	}
}

func TestParseBirthDateTrimsSpaces(t *testing.T) {
	calc, err := bazi.Calculate("  1995-01-01  ", "zi", false)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	payload, err := calc.InputPayload()
	if err != nil {
		t.Fatalf("InputPayload: %v", err)
	}
	var input map[string]any
	if err := json.Unmarshal(payload, &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if input["birth_date"] != "1995-01-01" {
		t.Fatalf("expected normalized birth_date, got %#v", input["birth_date"])
	}
}

func TestParseBirthDateRejectsFutureDate(t *testing.T) {
	loc := clock.Location()
	future := clock.Now().In(loc).AddDate(0, 0, 1).Format("2006-01-02")
	_, err := bazi.Calculate(future, "zi", false)
	if err == nil || !strings.Contains(err.Error(), "future") {
		t.Fatalf("expected future date rejection, got %v", err)
	}
}

func TestParseBirthDateAcceptsToday(t *testing.T) {
	today := clock.Now().In(clock.Location()).Format("2006-01-02")
	_, err := bazi.Calculate(today, "zi", false)
	if err != nil {
		t.Fatalf("today should be accepted: %v", err)
	}
}

func TestParseBirthDateYearBoundaries(t *testing.T) {
	if _, err := bazi.Calculate("1900-01-01", "zi", false); err != nil {
		t.Fatalf("1900-01-01 should be accepted: %v", err)
	}
	_, err := bazi.Calculate("1899-12-31", "zi", false)
	if err == nil {
		t.Fatalf("expected rejection before 1900")
	}
	_, err = bazi.Calculate("2101-01-01", "zi", false)
	if err == nil {
		t.Fatalf("expected rejection after 2100")
	}
	_, err = bazi.Calculate("2100-12-31", "zi", false)
	if err == nil || !strings.Contains(err.Error(), "future") {
		t.Fatalf("2100-12-31 should be in-range but rejected as future, got %v", err)
	}
}
