package bazi

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/bazi/calendar"
)

const (
	AlgorithmVersionBaziV2POC = "bazi-v2-poc"
	MethodNoteV2                = "本功能采用节气与立春换年的自研 Go 规则（bazi-v2-poc），仅供传统文化学习参考，不等同于专业八字排盘。"
	CompatibilityNoteV2       = "与 bazi-simple-v1 的年柱/月柱可能存在差异"
)

// CalendarBasis documents how v2 pillars were derived.
type CalendarBasis struct {
	YearBoundary   string `json:"year_boundary"`
	MonthBoundary  string `json:"month_boundary"`
	TrueSolarTime  bool   `json:"true_solar_time"`
	DayPillarBasis string `json:"day_pillar_basis"`
	Note           string `json:"note"`
}

// CalculationResultV2 is the isolated bazi-v2 POC result; not wired to Create API by default.
type CalculationResultV2 struct {
	BirthDate           string
	BirthHourBranch     string
	BirthHourUnknown    bool
	PillarsV2           Pillars
	CalendarBasis       CalendarBasis
	DayMaster           string
	FiveElements        FiveElements
	MethodNote          string
	CompatibilityNote   string
	Limits              []string
}

func (c CalculationResultV2) ResultPayload() (json.RawMessage, error) {
	pillars := map[string]string{
		"year":  c.PillarsV2.Year,
		"month": c.PillarsV2.Month,
		"day":   c.PillarsV2.Day,
	}
	if !c.BirthHourUnknown && c.PillarsV2.Hour != "" {
		pillars["hour"] = c.PillarsV2.Hour
	}

	payload := map[string]any{
		"algorithm_version":   AlgorithmVersionBaziV2POC,
		"method_note":         c.MethodNote,
		"calendar_basis":      c.CalendarBasis,
		"pillars_v2":          pillars,
		"compatibility_note":  c.CompatibilityNote,
		"day_master":          c.DayMaster,
		"five_elements":       c.FiveElements,
		"calculation_meta":    map[string]any{"limits": c.Limits},
	}
	return json.Marshal(payload)
}

// CalculateV2 computes year/month pillars with 立春换年 and 节气月柱.
// Day/hour pillars reuse v1 rules; true solar time is not applied (see ALG1.1).
func CalculateV2(birthDate string, birthHourBranch string, birthHourUnknown bool) (CalculationResultV2, error) {
	t, err := parseBirthDate(birthDate)
	if err != nil {
		return CalculationResultV2{}, err
	}

	moment := birthMoment(t, birthHourBranch, birthHourUnknown)

	baziYear := calendar.BaziYear(moment)
	yearPillar := yearPillar(baziYear)
	yearStemIdx := stemIndex(yearPillar)
	monthBranchIdx := calendar.MonthBranchIndex(moment)
	monthPillar := monthPillarFromBranch(yearStemIdx, monthBranchIdx)
	dayPillar := dayPillar(t)
	dayStemIdx := stemIndex(dayPillar)

	pillars := Pillars{
		Year:  yearPillar,
		Month: monthPillar,
		Day:   dayPillar,
		Hour:  "",
	}

	limits := []string{
		"年柱按立春换年（bazi-v2-poc）",
		"月柱按十二节令切换月令（bazi-v2-poc）",
		"日柱沿用 v1 固定基准日推算，未做真太阳时校正",
		"节令时刻按本地正午近似，非天文台精确时刻",
	}

	hourBranch := strings.TrimSpace(strings.ToLower(birthHourBranch))
	if !birthHourUnknown {
		if hourBranch == "" {
			return CalculationResultV2{}, fmt.Errorf("%w: birth_hour_branch is required", ErrInvalidParams)
		}
		if _, ok := validHourBranches[hourBranch]; !ok {
			return CalculationResultV2{}, fmt.Errorf("%w: invalid birth_hour_branch", ErrInvalidParams)
		}
		pillars.Hour = hourPillar(dayStemIdx, validHourBranches[hourBranch])
	} else {
		limits = append(limits, "出生时辰未知，未生成时柱")
	}

	elements := countFiveElements(pillars)
	dayMaster := string([]rune(dayPillar)[:1])

	return CalculationResultV2{
		BirthDate:         t.Format("2006-01-02"),
		BirthHourBranch:   hourBranch,
		BirthHourUnknown:  birthHourUnknown,
		PillarsV2:         pillars,
		CalendarBasis:     defaultCalendarBasis(),
		DayMaster:         dayMaster,
		FiveElements:      elements,
		MethodNote:        MethodNoteV2,
		CompatibilityNote: CompatibilityNoteV2,
		Limits:            limits,
	}, nil
}

func defaultCalendarBasis() CalendarBasis {
	return CalendarBasis{
		YearBoundary:   "lichun",
		MonthBoundary:  "solar_terms_jie",
		TrueSolarTime:  false,
		DayPillarBasis: "fixed_epoch_v1",
		Note:           "节气与立春换年仅用于传统文化学习参考；真太阳时延后至 ALG1.1",
	}
}

// monthPillarFromBranch applies 年上起月法（甲己丙作首…）to the solar-term month branch.
func monthPillarFromBranch(yearStemIdx, monthBranchIdx int) string {
	monthStemIdx := (yearStemIdx*2 + monthBranchIdx) % 10
	return heavenlyStems[monthStemIdx] + earthBranches[monthBranchIdx]
}

func birthMoment(birthDate time.Time, birthHourBranch string, birthHourUnknown bool) time.Time {
	loc := clock.Location()
	y, m, d := birthDate.Date()
	if birthHourUnknown {
		return time.Date(y, m, d, 12, 0, 0, 0, loc)
	}
	branch := strings.TrimSpace(strings.ToLower(birthHourBranch))
	idx, ok := validHourBranches[branch]
	if !ok {
		return time.Date(y, m, d, 12, 0, 0, 0, loc)
	}
	hour := branchRepresentativeHour(idx)
	return time.Date(y, m, d, hour, 0, 0, 0, loc)
}

func branchRepresentativeHour(branchIdx int) int {
	if branchIdx == 0 {
		return 0
	}
	return branchIdx * 2
}
