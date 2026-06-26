package qimen

import (
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/bazi/calendar"
)

const (
	TermKindJie = "jie"
	TermKindQi  = "qi"

	professionalTermPrecisionJie       = "formula_approximation"
	professionalTermPrecisionQiMidpoint  = "formula_midpoint_approximation"
	professionalTwentyFourTermsNote      = "ALG2.4C 第一版：十二节沿用公式近似，十二气为相邻节令中点近似；非天文台秒级交节"
)

// ProfessionalSolarTerm is one of the twenty-four solar terms with metadata.
type ProfessionalSolarTerm struct {
	Name      string
	Index     int
	Kind      string
	Time      time.Time
	Source    string
	Precision string
	Note      string
}

// TwentyFourSolarTermCatalog defines stable order, index, and jie/qi kind for all 24 terms.
var TwentyFourSolarTermCatalog = []struct {
	Name string
	Kind string
}{
	{"小寒", TermKindJie},
	{"大寒", TermKindQi},
	{"立春", TermKindJie},
	{"雨水", TermKindQi},
	{"惊蛰", TermKindJie},
	{"春分", TermKindQi},
	{"清明", TermKindJie},
	{"谷雨", TermKindQi},
	{"立夏", TermKindJie},
	{"小满", TermKindQi},
	{"芒种", TermKindJie},
	{"夏至", TermKindQi},
	{"小暑", TermKindJie},
	{"大暑", TermKindQi},
	{"立秋", TermKindJie},
	{"处暑", TermKindQi},
	{"白露", TermKindJie},
	{"秋分", TermKindQi},
	{"寒露", TermKindJie},
	{"霜降", TermKindQi},
	{"立冬", TermKindJie},
	{"小雪", TermKindQi},
	{"大雪", TermKindJie},
	{"冬至", TermKindQi},
}

var jieNameToCalendar = map[string]calendar.Jie{
	"小寒": calendar.JieXiaoHan,
	"立春": calendar.JieLiChun,
	"惊蛰": calendar.JieJingZhe,
	"清明": calendar.JieQingMing,
	"立夏": calendar.JieLiXia,
	"芒种": calendar.JieMangZhong,
	"小暑": calendar.JieXiaoShu,
	"立秋": calendar.JieLiQiu,
	"白露": calendar.JieBaiLu,
	"寒露": calendar.JieHanLu,
	"立冬": calendar.JieLiDong,
	"大雪": calendar.JieDaXue,
}

// TwentyFourSolarTermProvider extends SolarTermProvider with full 24-term output.
type TwentyFourSolarTermProvider interface {
	SolarTermProvider
	TwentyFourTerms(year int, loc *time.Location) []ProfessionalSolarTerm
}

// TwentyFourTerms returns 24 solar-term points for the solar year anchored at 小寒 of year.
func (FormulaSolarTermProvider) TwentyFourTerms(year int, loc *time.Location) []ProfessionalSolarTerm {
	if loc == nil {
		loc = clock.Location()
	}
	terms := make([]ProfessionalSolarTerm, 0, 24)
	for i, spec := range TwentyFourSolarTermCatalog {
		at, precision, note := computeTwentyFourTermTime(spec.Name, spec.Kind, year, loc)
		terms = append(terms, ProfessionalSolarTerm{
			Name:      spec.Name,
			Index:     i,
			Kind:      spec.Kind,
			Time:      at,
			Source:    professionalSolarTermSource,
			Precision: precision,
			Note:      note,
		})
	}
	return terms
}

func computeTwentyFourTermTime(name, kind string, year int, loc *time.Location) (time.Time, string, string) {
	if kind == TermKindJie {
		if jie, ok := jieNameToCalendar[name]; ok {
			at := calendar.JieTime(year, jie).In(loc)
			return at, professionalTermPrecisionJie, "十二节公式近似，本地正午"
		}
	}
	if name == "夏至" {
		at := formulaXiaZhiTime(year, loc)
		return at, professionalTermPrecisionJie, "夏至中气公式近似，用于阴阳遁与起局"
	}
	if name == "冬至" {
		at := formulaDongZhiTime(year, loc)
		return at, professionalTermPrecisionJie, "冬至中气公式近似，用于阴阳遁与起局"
	}
	prevName, nextName := adjacentTwentyFourTermNames(name)
	prevAt := termAnchorTime(prevName, year, loc)
	nextAt := termAnchorTime(nextName, year, loc)
	if nextAt.Before(prevAt) {
		nextAt = termAnchorTime(nextName, year+1, loc)
	}
	mid := prevAt.Add(nextAt.Sub(prevAt) / 2)
	return mid, professionalTermPrecisionQiMidpoint, "十二气第一版：相邻节令中点近似，pending_verification"
}

func adjacentTwentyFourTermNames(name string) (prev, next string) {
	for i, spec := range TwentyFourSolarTermCatalog {
		if spec.Name != name {
			continue
		}
		prev = TwentyFourSolarTermCatalog[(i+23)%24].Name
		next = TwentyFourSolarTermCatalog[(i+1)%24].Name
		return prev, next
	}
	return "小寒", "大寒"
}

func termAnchorTime(name string, year int, loc *time.Location) time.Time {
	for _, spec := range TwentyFourSolarTermCatalog {
		if spec.Name != name {
			continue
		}
		at, _, _ := computeTwentyFourTermTime(name, spec.Kind, year, loc)
		return at
	}
	return calendar.JieTime(year, calendar.JieXiaoHan).In(loc)
}

func collectTwentyFourTermsAround(t time.Time, provider SolarTermProvider) []ProfessionalSolarTerm {
	t = normalizeProfessionalMoment(t)
	loc := t.Location()
	p := asTwentyFourProvider(provider)
	out := make([]ProfessionalSolarTerm, 0, 72)
	for y := t.Year() - 1; y <= t.Year()+1; y++ {
		out = append(out, p.TwentyFourTerms(y, loc)...)
	}
	return out
}

func asTwentyFourProvider(provider SolarTermProvider) TwentyFourSolarTermProvider {
	if provider == nil {
		return FormulaSolarTermProvider{}
	}
	if p, ok := provider.(TwentyFourSolarTermProvider); ok {
		return p
	}
	return FormulaSolarTermProvider{}
}

// ResolveCurrentProfessionalTerm returns the latest 24-term point not after t.
func ResolveCurrentProfessionalTerm(t time.Time, provider SolarTermProvider) ProfessionalSolarTerm {
	t = normalizeProfessionalMoment(t)
	var best time.Time
	current := ProfessionalSolarTerm{
		Name: "小寒", Index: 0, Kind: TermKindJie,
		Source: professionalSolarTermSource, Precision: professionalTermPrecisionJie,
		Note: professionalTwentyFourTermsNote,
	}
	found := false
	for _, term := range collectTwentyFourTermsAround(t, provider) {
		if term.Time.After(t) {
			continue
		}
		if !found || term.Time.After(best) {
			best = term.Time
			current = term
			found = true
		}
	}
	return current
}

// ResolvePreviousProfessionalTerm returns the 24-term point immediately before the current term at t.
func ResolvePreviousProfessionalTerm(t time.Time, provider SolarTermProvider) ProfessionalSolarTerm {
	current := ResolveCurrentProfessionalTerm(t, provider)
	var best time.Time
	prev := current
	found := false
	for _, term := range collectTwentyFourTermsAround(t, provider) {
		if !term.Time.Before(current.Time) {
			continue
		}
		if !found || term.Time.After(best) {
			best = term.Time
			prev = term
			found = true
		}
	}
	if !found {
		prev.Note = professionalTwentyFourTermsNote + "；previous term fallback=current"
	}
	return prev
}

func twentyFourTermCalendarNote(term ProfessionalSolarTerm) string {
	note := professionalTwentyFourTermsNote
	if term.Kind == TermKindQi && term.Precision == professionalTermPrecisionQiMidpoint {
		note += "；当前气令时刻为中点近似"
	}
	return note
}
