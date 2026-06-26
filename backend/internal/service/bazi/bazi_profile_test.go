package bazi

import (
	"strings"
	"testing"
)

func TestDifferentBirthDatesProduceDifferentProfiles(t *testing.T) {
	cases := []struct {
		date    string
		hour    string
		unknown bool
	}{
		{"1995-03-12", "mao", false},
		{"1995-08-20", "wu", false},
		{"2000-12-05", "zi", false},
		{"1988-05-18", "", true},
		{"2003-10-01", "you", false},
	}

	balanceTypes := make(map[string]struct{})
	freeContents := make(map[string]struct{})
	for _, c := range cases {
		calc, err := Calculate(c.date, c.hour, c.unknown)
		if err != nil {
			t.Fatalf("Calculate(%s): %v", c.date, err)
		}
		if calc.BaziProfile.ElementBalanceType == "" {
			t.Fatalf("expected element_balance_type for %s", c.date)
		}
		if calc.InterpretationLens.StrengthHint == "" {
			t.Fatalf("expected strength_hint for %s", c.date)
		}
		balanceTypes[calc.BaziProfile.ElementBalanceType] = struct{}{}
		freeContents[BuildFreeContent(calc)] = struct{}{}
		t.Logf("%s: day_master=%s balance=%s style=%s theme=%s season=%s",
			c.date, calc.DayMaster, calc.BaziProfile.ElementBalanceType,
			calc.BaziProfile.ActionStyle, calc.BaziProfile.ReflectionTheme, calc.BaziProfile.SeasonTendency)
	}
	if len(balanceTypes) < 2 {
		t.Fatalf("expected varied balance types across test cases, got %d", len(balanceTypes))
	}
	if len(freeContents) < len(cases) {
		t.Fatalf("expected distinct free_content per birth info")
	}
}

func TestUnknownHourOmitsHourPillar(t *testing.T) {
	calc, err := Calculate("1988-05-18", "", true)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if calc.Pillars.Hour != "" {
		t.Fatalf("hour pillar must be empty when unknown")
	}
	raw, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	body := string(raw)
	if strings.Contains(body, `"hour"`) {
		t.Fatalf("result_payload must not contain hour key when unknown")
	}
	free := BuildFreeContent(calc)
	if !strings.Contains(free, "时辰未知") {
		t.Fatalf("free_content must mention unknown hour")
	}
}

func TestBuildFreeContentDiffersByElements(t *testing.T) {
	c1, _ := Calculate("1995-03-12", "mao", false)
	c2, _ := Calculate("2000-12-05", "zi", false)
	f1 := BuildFreeContent(c1)
	f2 := BuildFreeContent(c2)
	if f1 == f2 {
		t.Fatalf("free_content must differ for different birth info")
	}
	if !strings.Contains(f1, c1.BaziProfile.ElementBalanceType) {
		t.Fatalf("free_content should include balance type")
	}
}

func TestBuildBaziUserPromptIncludesProfileAndLens(t *testing.T) {
	calc, err := Calculate("1995-03-12", "mao", false)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	raw, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	input, err := buildFullReportPromptInput(raw, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	prompt := buildBaziUserPrompt(input)
	for _, required := range []string{
		"bazi_profile",
		"interpretation_lens",
		"element_balance_type=",
		"reflection_theme=",
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("prompt missing %q", required)
		}
	}
}

func TestBuildBaziUserPromptExcludesSensitiveFields(t *testing.T) {
	calc, err := Calculate("1995-03-12", "mao", false)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	raw, _ := calc.ResultPayload()
	prompt := buildBaziUserPrompt(mustPromptInput(t, raw, BuildFreeContent(calc)))
	for _, forbidden := range []string{
		"session_key",
		"input_payload",
		"result_payload",
		"1995-03-12",
		"birth_date",
		"birth_hour_branch",
		"mao",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt must not contain %s", forbidden)
		}
	}
}

func TestResultPayloadIncludesProfileFields(t *testing.T) {
	calc, err := Calculate("2003-10-01", "you", false)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	raw, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	body := string(raw)
	for _, key := range []string{"bazi_profile", "interpretation_lens", "differentiation_seed"} {
		if !strings.Contains(body, key) {
			t.Fatalf("result_payload missing %s", key)
		}
	}
	if strings.Contains(body, "2003-10-01") {
		t.Fatalf("result_payload must not contain birth date")
	}
}

func mustPromptInput(t *testing.T, raw []byte, free string) *fullReportPromptInput {
	t.Helper()
	input, err := buildFullReportPromptInput(raw, free)
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	return input
}
