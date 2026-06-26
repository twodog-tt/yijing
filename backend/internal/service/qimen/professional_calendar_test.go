package qimen

import (
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func TestFormulaSolarTermProviderIncludesJieAndDunBoundaries(t *testing.T) {
	loc := clock.Location()
	provider := FormulaSolarTermProvider{}
	points := provider.TermPoints(2024, loc)
	if len(points) < 14 {
		t.Fatalf("expected at least 14 points, got %d", len(points))
	}
	names := map[string]bool{}
	for _, p := range points {
		names[p.Name] = true
		if p.Source != professionalSolarTermSource {
			t.Fatalf("unexpected source %q for %q", p.Source, p.Name)
		}
		if p.Precision != professionalSolarTermPrecision {
			t.Fatalf("unexpected precision %q for %q", p.Precision, p.Name)
		}
	}
	for _, want := range []string{"小寒", "惊蛰", "冬至", "夏至"} {
		if !names[want] {
			t.Fatalf("missing term %q", want)
		}
	}
}

func TestResolveProfessionalCalendarBasisUsesFormulaApproximation(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	basis := ResolveProfessionalCalendarBasis(when, FormulaSolarTermProvider{})
	if basis.SolarTerm == "" {
		t.Fatal("solar_term empty")
	}
	if basis.SolarTermTime == "" {
		t.Fatal("solar_term_time empty")
	}
	if basis.JieqiBasis != jieqiBasisPOC {
		t.Fatalf("jieqi_basis=%q", basis.JieqiBasis)
	}
	if basis.TimeBasis != professionalTimeBasis {
		t.Fatalf("time_basis=%q", basis.TimeBasis)
	}
}

func TestResolveProfessionalDunBoundaries2024(t *testing.T) {
	loc := clock.Location()
	provider := FormulaSolarTermProvider{}
	xiazhi := formulaXiaZhiTime(2024, loc)
	dongzhi := formulaDongZhiTime(2024, loc)

	tests := []struct {
		name     string
		when     time.Time
		wantType string
	}{
		{"before_xiazhi", xiazhi.Add(-2 * time.Hour), "yang"},
		{"after_xiazhi", xiazhi.Add(2 * time.Hour), "yin"},
		{"before_dongzhi", dongzhi.Add(-2 * time.Hour), "yin"},
		{"after_dongzhi", dongzhi.Add(2 * time.Hour), "yang"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dun := ResolveProfessionalDun(tc.when, provider)
			if dun.Type != tc.wantType {
				t.Fatalf("type=%q want %q (basis_term=%q basis_time=%q)", dun.Type, tc.wantType, dun.BasisTerm, dun.BasisTime)
			}
			if dun.Method != DunMethodSolarTermBoundary {
				t.Fatalf("method=%q", dun.Method)
			}
			if dun.BasisTerm == "" || dun.BasisTime == "" {
				t.Fatalf("missing basis fields: %+v", dun)
			}
		})
	}
}

func TestResolveProfessionalDunStableForSameInput(t *testing.T) {
	when := time.Date(2024, 8, 7, 15, 0, 0, 0, clock.Location())
	a := ResolveProfessionalDun(when, FormulaSolarTermProvider{})
	b := ResolveProfessionalDun(when, FormulaSolarTermProvider{})
	if a != b {
		t.Fatalf("dun not stable: %+v vs %+v", a, b)
	}
}

func TestResolveProfessionalDunDoesNotUseGregorianJune21(t *testing.T) {
	loc := clock.Location()
	when := time.Date(2024, 6, 21, 0, 30, 0, 0, loc)
	dun := ResolveProfessionalDun(when, FormulaSolarTermProvider{})
	xiazhi := formulaXiaZhiTime(2024, loc)
	if when.Before(xiazhi) && dun.Type != "yang" {
		t.Fatalf("before formula 夏至 should stay yang, got %q", dun.Type)
	}
	if !when.Before(xiazhi) && dun.Type != "yin" {
		t.Fatalf("after formula 夏至 should be yin, got %q", dun.Type)
	}
}
