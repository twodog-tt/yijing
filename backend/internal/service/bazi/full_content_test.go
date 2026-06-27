package bazi

import (
	"strings"
	"testing"
)

func TestBuildFullContentUsesReportSections(t *testing.T) {
	calc, err := Calculate("1995-03-12", "mao", false)
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
		sectionBrief, sectionPillars, sectionTendency, sectionPacing,
		sectionReflection, sectionActions, sectionBoundary,
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected full content to contain %q", snippet)
		}
	}
	if containsForbiddenReportPhrase(content) {
		t.Fatalf("full content body must not contain forbidden phrases")
	}
	if strings.Contains(content, sectionV2Overall) {
		t.Fatalf("bazi-simple-v1 content must not use v2 sections")
	}
}

func TestBuildFullContentHourUnknownNote(t *testing.T) {
	calc, err := Calculate("1988-05-18", "", true)
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
	if !strings.Contains(content, "时辰未知") || !strings.Contains(content, "不生成时柱") {
		t.Fatalf("expected hour-unknown note in full content")
	}
	if strings.Contains(reportBodyExcludingBoundary(content), "时柱：甲") {
		t.Fatalf("must not analyze hour pillar when unknown")
	}
}

func TestBuildFullContentRejectsInvalidPayload(t *testing.T) {
	_, err := BuildFullContent(jsonRaw(`{"pillars":{}}`), "")
	if err == nil {
		t.Fatalf("expected error for invalid payload")
	}
}

func TestBuildFullContentV2UsesEightSectionsAndStructuredFields(t *testing.T) {
	content, v2 := buildV2FullContentForTest(t, "1995-03-12", "mao", false)

	for _, snippet := range []string{
		sectionV2Overall, sectionV2Basis, sectionV2Pillars, sectionV2Elements,
		sectionV2Strength, sectionV2Caution, sectionV2Pacing, sectionV2Boundary,
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected v2 full content to contain %q", snippet)
		}
	}
	for _, snippet := range []string{
		"bazi-v2-poc",
		"year_boundary=lichun",
		"month_boundary=solar_terms_jie",
		"true_solar_time=false",
		"年柱：" + v2.PillarsV2.Year,
		"月柱：" + v2.PillarsV2.Month,
		"日柱：" + v2.PillarsV2.Day,
		"时柱：" + v2.PillarsV2.Hour,
		"五行结构：木",
		"火",
		"土",
		"金",
		"水",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected v2 content to contain %q\n%s", snippet, content)
		}
	}
	if strings.Contains(content, v2.BirthDate) {
		t.Fatalf("v2 full content must not contain full birth date")
	}
	for _, forbidden := range []string{"session_key", "input_payload", "result_payload", "prompt"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v2 full content must not contain %q", forbidden)
		}
	}
	if containsForbiddenReportPhrase(content) {
		t.Fatalf("v2 full content body must not contain forbidden phrases")
	}
}

func TestBuildFullContentV2HourUnknownDoesNotInventHourPillar(t *testing.T) {
	content, _ := buildV2FullContentForTest(t, "1988-05-18", "", true)
	if !strings.Contains(content, "时辰未知") || !strings.Contains(content, "本次未生成时柱") {
		t.Fatalf("expected v2 hour unknown note")
	}
	if strings.Contains(content, "时柱：甲") || strings.Contains(content, "时柱：乙") ||
		strings.Contains(content, "时柱：丙") || strings.Contains(content, "时柱：丁") ||
		strings.Contains(content, "时柱：戊") || strings.Contains(content, "时柱：己") ||
		strings.Contains(content, "时柱：庚") || strings.Contains(content, "时柱：辛") ||
		strings.Contains(content, "时柱：壬") || strings.Contains(content, "时柱：癸") {
		t.Fatalf("v2 unknown-hour content must not invent hour pillar")
	}
	if strings.Contains(content, "1988-05-18") {
		t.Fatalf("v2 unknown-hour content must not contain full birth date")
	}
}

func TestBaziFullReportSampleDifferentiation(t *testing.T) {
	cases := []struct {
		name string
		date string
		hour string
		unk  bool
	}{
		{"1995-03-12 mao", "1995-03-12", "mao", false},
		{"1995-08-20 wu", "1995-08-20", "wu", false},
		{"2000-12-05 zi", "2000-12-05", "zi", false},
		{"1988-05-18 unknown", "1988-05-18", "", true},
		{"2003-10-01 you", "2003-10-01", "you", false},
	}

	contents := make(map[string]string, len(cases))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calc, err := Calculate(tc.date, tc.hour, tc.unk)
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

			if strings.Contains(content, tc.date) {
				t.Fatalf("full content must not contain birth date")
			}
			if !strings.Contains(content, calc.BaziProfile.ElementBalanceType) {
				t.Fatalf("expected element balance %q in content", calc.BaziProfile.ElementBalanceType)
			}
			if !strings.Contains(content, calc.BaziProfile.ActionStyle) {
				t.Fatalf("expected action style %q in content", calc.BaziProfile.ActionStyle)
			}
			if !strings.Contains(content, calc.BaziProfile.ReflectionTheme) {
				t.Fatalf("expected reflection theme %q in content", calc.BaziProfile.ReflectionTheme)
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

	if contents["1995-03-12 mao"] == contents["1995-08-20 wu"] {
		t.Fatalf("expected different full content for different inputs")
	}
	if contents["2000-12-05 zi"] == contents["2003-10-01 you"] {
		t.Fatalf("expected different full content for different inputs")
	}
}

func jsonRaw(s string) []byte { return []byte(s) }

func buildV2FullContentForTest(t *testing.T, date, hour string, unknown bool) (string, CalculationResultV2) {
	t.Helper()
	v2, err := CalculateV2(date, hour, unknown)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	calc := CalculationResultFromV2(v2)
	payload, err := BuildV2APIResultPayload(v2, calc)
	if err != nil {
		t.Fatalf("BuildV2APIResultPayload: %v", err)
	}
	content, err := BuildFullContent(payload, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	return content, v2
}
