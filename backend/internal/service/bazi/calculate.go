package bazi

import (
	"fmt"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

var (
	heavenlyStems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	earthBranches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
)

// gregorianMonthBranch maps公历月份到月支（简化固定分界，非节气）。
var gregorianMonthBranch = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0}

func Calculate(birthDate string, birthHourBranch string, birthHourUnknown bool) (CalculationResult, error) {
	t, err := parseBirthDate(birthDate)
	if err != nil {
		return CalculationResult{}, err
	}

	year, month, _ := t.Date()
	limits := []string{
		"年柱按公历年份简化推算，未按立春换年",
		"月柱按公历月份固定映射月支，非节气月柱",
		"日柱按公历日期推算，未做真太阳时校正",
	}

	yearPillar := yearPillar(year)
	yearStemIdx := stemIndex(yearPillar)
	monthPillar := simplifiedMonthPillar(yearStemIdx, int(month))
	dayPillar := dayPillar(t)
	dayStemIdx := stemIndex(dayPillar)

	pillars := Pillars{
		Year:  yearPillar,
		Month: monthPillar,
		Day:   dayPillar,
		Hour:  "",
	}

	hourBranch := strings.TrimSpace(strings.ToLower(birthHourBranch))
	if !birthHourUnknown {
		if hourBranch == "" {
			return CalculationResult{}, fmt.Errorf("%w: birth_hour_branch is required", ErrInvalidParams)
		}
		if _, ok := validHourBranches[hourBranch]; !ok {
			return CalculationResult{}, fmt.Errorf("%w: invalid birth_hour_branch", ErrInvalidParams)
		}
		pillars.Hour = hourPillar(dayStemIdx, validHourBranches[hourBranch])
	} else {
		limits = append(limits, "出生时辰未知，未生成时柱")
	}

	elements := countFiveElements(pillars)
	dayMaster := string([]rune(dayPillar)[:1])
	reflection, actions := buildReflection(dayMaster, elements)

	return CalculationResult{
		BirthDate:         t.Format("2006-01-02"),
		BirthHourBranch:   hourBranch,
		BirthHourUnknown:  birthHourUnknown,
		Pillars:           pillars,
		DayMaster:         dayMaster,
		FiveElements:      elements,
		ReflectionFocus:   reflection,
		ActionSuggestions: actions,
		MethodNote:        MethodNote,
		Limits:            limits,
	}, nil
}

func parseBirthDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: birth_date must be YYYY-MM-DD", ErrInvalidParams)
	}
	if t.Year() < 1900 || t.Year() > 2100 {
		return time.Time{}, fmt.Errorf("%w: birth_date out of supported range", ErrInvalidParams)
	}

	loc := clock.Location()
	today := clock.Now().In(loc)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	birthDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	if birthDate.After(todayDate) {
		return time.Time{}, fmt.Errorf("%w: birth_date cannot be in the future", ErrInvalidParams)
	}
	return birthDate, nil
}

func sexagenary(index int) string {
	if index < 0 {
		index += 60
	}
	index %= 60
	return heavenlyStems[index%10] + earthBranches[index%12]
}

func stemIndex(pillar string) int {
	if pillar == "" {
		return 0
	}
	runes := []rune(pillar)
	for i, stem := range heavenlyStems {
		if runes[0] == []rune(stem)[0] {
			return i
		}
	}
	return 0
}

func branchIndex(pillar string) int {
	if pillar == "" {
		return 0
	}
	runes := []rune(pillar)
	for i, branch := range earthBranches {
		if runes[1] == []rune(branch)[0] {
			return i
		}
	}
	return 0
}

func yearPillar(year int) string {
	idx := (year - 4) % 60
	return sexagenary(idx)
}

func dayPillar(t time.Time) string {
	base := time.Date(1900, 1, 1, 12, 0, 0, 0, time.UTC)
	current := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
	days := int(current.Sub(base).Hours() / 24)
	idx := (days + 10) % 60
	return sexagenary(idx)
}

func simplifiedMonthPillar(yearStemIdx, month int) string {
	if month < 1 || month > 12 {
		return "not_available"
	}
	monthBranchIdx := gregorianMonthBranch[month-1]
	monthStemIdx := (yearStemIdx*2 + monthBranchIdx) % 10
	return heavenlyStems[monthStemIdx] + earthBranches[monthBranchIdx]
}

func hourPillar(dayStemIdx, hourBranchIdx int) string {
	hourStemIdx := (dayStemIdx*2 + hourBranchIdx) % 10
	return heavenlyStems[hourStemIdx] + earthBranches[hourBranchIdx]
}

func countFiveElements(p Pillars) FiveElements {
	counts := FiveElements{}
	for _, pillar := range []string{p.Year, p.Month, p.Day, p.Hour} {
		if pillar == "" || pillar == "not_available" {
			continue
		}
		addStemElements(&counts, pillar)
		addBranchElements(&counts, pillar)
	}
	return counts
}

func addStemElements(counts *FiveElements, pillar string) {
	switch stemIndex(pillar) {
	case 0, 1:
		counts.Wood++
	case 2, 3:
		counts.Fire++
	case 4, 5:
		counts.Earth++
	case 6, 7:
		counts.Metal++
	case 8, 9:
		counts.Water++
	}
}

func addBranchElements(counts *FiveElements, pillar string) {
	switch branchIndex(pillar) {
	case 2, 3:
		counts.Wood++
	case 5, 6:
		counts.Fire++
	case 1, 4, 7, 10:
		counts.Earth++
	case 8, 9:
		counts.Metal++
	case 0, 11:
		counts.Water++
	}
}

func buildReflection(dayMaster string, elements FiveElements) (string, []string) {
	elementName := stemElementName(dayMaster)
	reflection := fmt.Sprintf("基于简化干支文化规则的学习参考：日主为「%s」，可先从%s相关的自我观察入手，作为行动整理的一个角度。", dayMaster, elementName)

	actions := []string{
		"记录近期一件让你有感受的小事，尝试从性格倾向角度做自我观察。",
		"选择一项可执行的小行动，先完成再复盘，而非追求结论。",
	}
	if dominant, weak := dominantElements(elements); dominant != "" {
		actions = append(actions, fmt.Sprintf("五行分布中%s元素相对较多、%s相对较少，可作为学习参考，不构成现实决策依据。", dominant, weak))
	}
	actions = append(actions, "以上内容基于简化干支文化规则，仅供自我观察与行动整理参考。")
	return reflection, actions
}

func stemElementName(dayMaster string) string {
	switch dayMaster {
	case "甲", "乙":
		return "木"
	case "丙", "丁":
		return "火"
	case "戊", "己":
		return "土"
	case "庚", "辛":
		return "金"
	case "壬", "癸":
		return "水"
	default:
		return "五行"
	}
}

func dominantElements(elements FiveElements) (string, string) {
	type pair struct {
		name  string
		count int
	}
	items := []pair{
		{"木", elements.Wood},
		{"火", elements.Fire},
		{"土", elements.Earth},
		{"金", elements.Metal},
		{"水", elements.Water},
	}
	maxItem, minItem := items[0], items[0]
	for _, item := range items[1:] {
		if item.count > maxItem.count {
			maxItem = item
		}
		if item.count < minItem.count {
			minItem = item
		}
	}
	if maxItem.count == 0 {
		return "", ""
	}
	return maxItem.name, minItem.name
}
