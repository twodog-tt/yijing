package qimen

import (
	"math"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/bazi/calendar"
)

const (
	professionalSolarTermSource   = "formula_solar_term_provider"
	professionalSolarTermPrecision = "day_hour_formula_approximation"
	professionalCalendarNote        = "ALG2.4C 二十四节气第一版：节令公式近似、气令中点近似（本地时），非天文台秒级交节；后续可替换权威交节表"
	professionalTimeBasis           = "local_time"
)

// ProfessionalSolarTermPoint is one solar-term boundary returned by a provider.
type ProfessionalSolarTermPoint struct {
	Name      string
	Time      time.Time
	Source    string
	Precision string
	Note      string
}

// SolarTermProvider supplies solar-term boundary points for a Gregorian year.
// Implementations may be swapped for authoritative ephemeris tables in later phases.
type SolarTermProvider interface {
	TermPoints(year int, loc *time.Location) []ProfessionalSolarTermPoint
}

// FormulaSolarTermProvider uses engineering formulas (same family as bazi/calendar 节令).
// It is NOT astronomical observatory precision.
type FormulaSolarTermProvider struct{}

func (FormulaSolarTermProvider) TermPoints(year int, loc *time.Location) []ProfessionalSolarTermPoint {
	if loc == nil {
		loc = clock.Location()
	}
	points := make([]ProfessionalSolarTermPoint, 0, 14)
	for jie := calendar.JieXiaoHan; jie <= calendar.JieDaXue; jie++ {
		at := calendar.JieTime(year, jie)
		points = append(points, ProfessionalSolarTermPoint{
			Name:      jieDisplayName(jie),
			Time:      at.In(loc),
			Source:    professionalSolarTermSource,
			Precision: professionalSolarTermPrecision,
			Note:      "十二节令公式近似，时刻按本地正午",
		})
	}
	for _, spec := range []struct {
		name string
		at   time.Time
		note string
	}{
		{"冬至", formulaDongZhiTime(year, loc), "中气冬至，用于阴阳遁边界"},
		{"夏至", formulaXiaZhiTime(year, loc), "中气夏至，用于阴阳遁边界"},
	} {
		points = append(points, ProfessionalSolarTermPoint{
			Name:      spec.name,
			Time:      spec.at,
			Source:    professionalSolarTermSource,
			Precision: professionalSolarTermPrecision,
			Note:      spec.note,
		})
	}
	return points
}

// formulaDongZhiTime approximates 冬至 local time using [Y*D+C]-L.
func formulaDongZhiTime(year int, loc *time.Location) time.Time {
	y := year % 100
	c := 22.60
	if year >= 2000 {
		c = 21.94
	}
	day := solarTermDay(y, c)
	return time.Date(year, time.December, day, calendarTermHour, 0, 0, 0, loc)
}

// formulaXiaZhiTime approximates 夏至 local time using [Y*D+C]-L.
func formulaXiaZhiTime(year int, loc *time.Location) time.Time {
	y := year % 100
	c := 21.37
	day := solarTermDay(y, c)
	return time.Date(year, time.June, day, calendarTermHour, 0, 0, 0, loc)
}

const calendarTermHour = 12

func solarTermDay(y int, c float64) int {
	day := int(math.Floor(float64(y)*0.2422+c)) - ((y - 1) / 4)
	if day < 1 {
		day = 1
	}
	return day
}

func defaultSolarTermProvider() SolarTermProvider {
	return FormulaSolarTermProvider{}
}

func normalizeProfessionalMoment(t time.Time) time.Time {
	if t.IsZero() {
		t = clock.Now()
	}
	return t.In(clock.Location())
}

func resolveCurrentJiePoint(t time.Time, provider SolarTermProvider) ProfessionalSolarTermPoint {
	term := ResolveCurrentProfessionalTerm(t, provider)
	return professionalSolarTermPointFromTerm(term)
}

func professionalSolarTermPointFromTerm(term ProfessionalSolarTerm) ProfessionalSolarTermPoint {
	return ProfessionalSolarTermPoint{
		Name:      term.Name,
		Time:      term.Time,
		Source:    term.Source,
		Precision: term.Precision,
		Note:      term.Note,
	}
}

func isDunBoundaryTerm(name string) bool {
	return name == "冬至" || name == "夏至"
}

// ResolveProfessionalCalendarBasis builds calendar metadata for professional preview (24-term).
func ResolveProfessionalCalendarBasis(t time.Time, provider SolarTermProvider) ProfessionalCalendarBasis {
	if provider == nil {
		provider = defaultSolarTermProvider()
	}
	t = normalizeProfessionalMoment(t)
	term := ResolveCurrentProfessionalTerm(t, provider)
	return ProfessionalCalendarBasis{
		SolarTerm:     term.Name,
		SolarTermTime: term.Time.Format(time.RFC3339),
		JieqiBasis:    jieqiBasisPOC,
		TimeBasis:     professionalTimeBasis,
		Note:          twentyFourTermCalendarNote(term),
	}
}

func findLatestDunBoundary(t time.Time, provider SolarTermProvider) (term string, at time.Time) {
	t = normalizeProfessionalMoment(t)
	loc := t.Location()
	var best time.Time
	termName := "冬至"
	found := false
	for y := t.Year() - 1; y <= t.Year()+1; y++ {
		for _, name := range []string{"冬至", "夏至"} {
			var boundary time.Time
			if name == "冬至" {
				boundary = formulaDongZhiTime(y, loc)
			} else {
				boundary = formulaXiaZhiTime(y, loc)
			}
			if boundary.After(t) {
				continue
			}
			if !found || boundary.After(best) {
				best = boundary
				termName = name
				found = true
			}
		}
	}
	if !found {
		return "冬至", time.Time{}
	}
	return termName, best
}

// ResolveProfessionalDun binds yin/yang dun to the latest 冬至/夏至 boundary from the provider.
func ResolveProfessionalDun(t time.Time, provider SolarTermProvider) ProfessionalDun {
	if provider == nil {
		provider = defaultSolarTermProvider()
	}
	t = normalizeProfessionalMoment(t)
	basisTerm, basisTime := findLatestDunBoundary(t, provider)
	dunType := "yin"
	if basisTerm == "冬至" {
		dunType = "yang"
	}
	note := "阴阳遁按最近一次冬至/夏至交节划分；交节时刻仍为公式近似"
	if basisTime.IsZero() {
		note = "未能定位冬至/夏至交节，回退为阴遁；仍为公式近似"
	}
	return ProfessionalDun{
		Type:       dunType,
		Ju:         0,
		Method:     DunMethodSolarTermBoundary,
		Yuan:       DunYuanPending,
		BasisTerm:  basisTerm,
		BasisTime:  formatBasisTime(basisTime),
		Note:       note,
	}
}

func formatBasisTime(at time.Time) string {
	if at.IsZero() {
		return "professional_pending"
	}
	return at.Format(time.RFC3339)
}
