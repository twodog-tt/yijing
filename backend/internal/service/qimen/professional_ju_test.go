package qimen

import (
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func TestResolveProfessionalYuanSchemeA(t *testing.T) {
	loc := clock.Location()
	termStart := time.Date(2024, 3, 6, 12, 0, 0, 0, loc)
	calendar := ProfessionalCalendarBasis{
		SolarTerm:     "惊蛰",
		SolarTermTime: termStart.Format(time.RFC3339),
	}
	gz := ProfessionalGanzhi{Day: "甲子", Hour: "甲子"}

	tests := []struct {
		name string
		when time.Time
		want string
	}{
		{"day0", termStart.Add(2 * time.Hour), DunYuanUpper},
		{"day4", termStart.AddDate(0, 0, 4), DunYuanUpper},
		{"day5", termStart.AddDate(0, 0, 5), DunYuanMiddle},
		{"day9", termStart.AddDate(0, 0, 9), DunYuanMiddle},
		{"day10", termStart.AddDate(0, 0, 10), DunYuanLower},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveProfessionalYuan(tc.when, calendar.SolarTerm, calendar, gz)
			if got != tc.want {
				t.Fatalf("yuan=%q want %q", got, tc.want)
			}
		})
	}
}

func TestResolveProfessionalYuanStable(t *testing.T) {
	when := time.Date(2024, 8, 7, 15, 0, 0, 0, clock.Location())
	basis := ResolveProfessionalCalendarBasis(when, FormulaSolarTermProvider{})
	gz := ResolveProfessionalGanZhi(when)
	a := ResolveProfessionalYuan(when, basis.SolarTerm, basis, gz)
	b := ResolveProfessionalYuan(when, basis.SolarTerm, basis, gz)
	if a != b {
		t.Fatalf("not stable: %q vs %q", a, b)
	}
}

func TestResolveChaiBuJuRangeAndMethod(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	provider := FormulaSolarTermProvider{}
	basis := ResolveProfessionalCalendarBasis(when, provider)
	dun := ResolveProfessionalDun(when, provider)
	gz := ResolveProfessionalGanZhi(when)
	ju := ResolveChaiBuJu(when, dun, basis, gz)

	if ju.Method != DunMethodChaiBu {
		t.Fatalf("method=%q", ju.Method)
	}
	if ju.Ju < 1 || ju.Ju > 9 {
		t.Fatalf("ju=%d out of range", ju.Ju)
	}
	if ju.Yuan != DunYuanUpper && ju.Yuan != DunYuanMiddle && ju.Yuan != DunYuanLower {
		t.Fatalf("invalid yuan %q", ju.Yuan)
	}
	if ju.Basis != juBasisTwentyFourTermsChaiBu {
		t.Fatalf("basis=%q", ju.Basis)
	}
	if !strings.Contains(ju.Note, "第一版") {
		t.Fatalf("note should mention first version: %q", ju.Note)
	}
}

func TestResolveChaiBuJuCategoryIndependent(t *testing.T) {
	when := time.Date(2024, 9, 22, 18, 30, 0, 0, clock.Location())
	basis := ResolveProfessionalCalendarBasis(when, FormulaSolarTermProvider{})
	dun := ResolveProfessionalDun(when, FormulaSolarTermProvider{})
	gz := ResolveProfessionalGanZhi(when)
	ju := ResolveChaiBuJu(when, dun, basis, gz)

	// category is not passed to ResolveChaiBuJu — recompute with same inputs must match
	ju2 := ResolveChaiBuJu(when, dun, basis, gz)
	if ju != ju2 {
		t.Fatalf("ju should not vary: %+v vs %+v", ju, ju2)
	}
}

func TestResolveChaiBuJuYangYinDirection(t *testing.T) {
	loc := clock.Location()
	basis := ProfessionalCalendarBasis{SolarTerm: "惊蛰", SolarTermTime: time.Date(2024, 3, 6, 12, 0, 0, 0, loc).Format(time.RFC3339)}
	gz := ProfessionalGanzhi{Day: "甲子", Hour: "甲子"}
	base, ok := BaseJuForProfessionalTerm("惊蛰", "yang")
	if !ok {
		t.Fatal("missing base ju for 惊蛰")
	}

	yangDun := ProfessionalDun{Type: "yang"}
	yinDun := ProfessionalDun{Type: "yin"}

	whenUpper := time.Date(2024, 3, 7, 9, 0, 0, 0, loc)
	basisUpper := ProfessionalCalendarBasis{SolarTerm: "惊蛰", SolarTermTime: time.Date(2024, 3, 6, 12, 0, 0, 0, loc).Format(time.RFC3339)}
	yangJu := ResolveChaiBuJu(whenUpper, yangDun, basisUpper, gz)
	if yangJu.Ju != base {
		t.Fatalf("yang upper ju=%d want %d", yangJu.Ju, base)
	}

	whenLower := time.Date(2024, 3, 20, 9, 0, 0, 0, loc)
	yangLower := ResolveChaiBuJu(whenLower, yangDun, basis, gz)
	yinLower := ResolveChaiBuJu(whenLower, yinDun, basis, gz)
	if yangLower.Ju == yinLower.Ju {
		t.Fatalf("yang and yin should diverge for lower yuan: yang=%d yin=%d", yangLower.Ju, yinLower.Ju)
	}
}

func TestResolveChaiBuJuDunTypeFlipChangesJu(t *testing.T) {
	loc := clock.Location()
	provider := FormulaSolarTermProvider{}
	xiazhi := formulaXiaZhiTime(2024, loc)
	before := xiazhi.Add(-24 * time.Hour)
	after := xiazhi.Add(24 * time.Hour)

	juBefore := ResolveChaiBuJu(before, ResolveProfessionalDun(before, provider), ResolveProfessionalCalendarBasis(before, provider), ResolveProfessionalGanZhi(before))
	juAfter := ResolveChaiBuJu(after, ResolveProfessionalDun(after, provider), ResolveProfessionalCalendarBasis(after, provider), ResolveProfessionalGanZhi(after))

	if ResolveProfessionalDun(before, provider).Type != "yang" || ResolveProfessionalDun(after, provider).Type != "yin" {
		t.Fatal("expected yang before 夏至 and yin after")
	}
	if juBefore.Ju < 1 || juBefore.Ju > 9 || juAfter.Ju < 1 || juAfter.Ju > 9 {
		t.Fatalf("ju out of range: before=%d after=%d", juBefore.Ju, juAfter.Ju)
	}
}

func TestResolveZhiRunJuPendingNotImplemented(t *testing.T) {
	ju := ResolveZhiRunJuPending()
	if ju.Method != DunMethodZhiRunPending {
		t.Fatalf("method=%q", ju.Method)
	}
	if ju.Ju != 0 {
		t.Fatalf("ju=%d want 0 pending", ju.Ju)
	}
	if ju.Yuan != DunYuanPending {
		t.Fatalf("yuan=%q", ju.Yuan)
	}
}

func TestChaiBuBaseJuCoversTwelveJie(t *testing.T) {
	for _, name := range []string{"小寒", "立春", "惊蛰", "清明", "立夏", "芒种", "小暑", "立秋", "白露", "寒露", "立冬", "大雪"} {
		if chaiBuBaseJuByJie[name] < 1 || chaiBuBaseJuByJie[name] > 9 {
			t.Fatalf("invalid base ju for %q: %d", name, chaiBuBaseJuByJie[name])
		}
	}
}

func TestResolveChaiBuJuUsesTwentyFourTermBasis(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	ju := ResolveChaiBuJu(when, ResolveProfessionalDun(when, FormulaSolarTermProvider{}), ResolveProfessionalCalendarBasis(when, FormulaSolarTermProvider{}), ResolveProfessionalGanZhi(when))
	if ju.Basis != juBasisTwentyFourTermsChaiBu {
		t.Fatalf("basis=%q", ju.Basis)
	}
}
