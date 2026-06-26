package qimen

import (
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/service/bazi/calendar"
)

var (
	proHeavenlyStems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	proEarthBranches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
)

const (
	ganzhiBasisProfessional = "bazi_calendar_v2_rules"
	ganzhiNoteProfessional    = "年柱立春换年、月柱十二节令切换；日柱固定基准日；时柱按本地时钟两小时段；未做真太阳时与晚子时换日校正"
)

// ResolveProfessionalGanZhi derives four pillars for qimen professional preview.
func ResolveProfessionalGanZhi(t time.Time) ProfessionalGanzhi {
	t = normalizeProfessionalMoment(t)
	baziYear := calendar.BaziYear(t)
	yearPillar := proYearPillar(baziYear)
	yearStemIdx := proStemIndex(yearPillar)
	monthBranchIdx := calendar.MonthBranchIndex(t)
	monthPillar := proMonthPillarFromBranch(yearStemIdx, monthBranchIdx)
	dayPillar := proDayPillar(t)
	hourPillar := proHourPillarFromClock(t, proStemIndex(dayPillar))
	return ProfessionalGanzhi{
		Year:  yearPillar,
		Month: monthPillar,
		Day:   dayPillar,
		Hour:  hourPillar,
		Basis: ganzhiBasisProfessional,
		Note:  ganzhiNoteProfessional,
	}
}

// ResolveXunFromGanZhi derives 旬首/空亡 from day/hour pillars (时家奇门优先用时柱).
func ResolveXunFromGanZhi(dayGanZhi, hourGanZhi string) Xun {
	source := strings.TrimSpace(hourGanZhi)
	if source == "" {
		source = strings.TrimSpace(dayGanZhi)
	}
	idx := proSexagenaryIndex(source)
	if idx < 0 {
		return Xun{XunShou: "professional_pending", EmptyBranches: []string{}}
	}
	xunIdx := idx / 10
	if xunIdx < 0 || xunIdx >= len(xunShouList) {
		return Xun{XunShou: "professional_pending", EmptyBranches: []string{}}
	}
	return Xun{
		XunShou:       xunShouList[xunIdx],
		EmptyBranches: append([]string(nil), xunEmpty[xunIdx]...),
	}
}

func proYearPillar(year int) string {
	return proSexagenary((year - 4) % 60)
}

func proDayPillar(t time.Time) string {
	base := time.Date(1900, 1, 1, 12, 0, 0, 0, time.UTC)
	current := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
	days := int(current.Sub(base).Hours() / 24)
	return proSexagenary((days + 10) % 60)
}

func proMonthPillarFromBranch(yearStemIdx, monthBranchIdx int) string {
	monthStemIdx := (yearStemIdx*2 + monthBranchIdx) % 10
	return proHeavenlyStems[monthStemIdx] + proEarthBranches[monthBranchIdx]
}

func proHourPillarFromClock(t time.Time, dayStemIdx int) string {
	hourBranchIdx := proHourBranchIndex(t.Hour())
	hourStemIdx := (dayStemIdx*2 + hourBranchIdx) % 10
	return proHeavenlyStems[hourStemIdx] + proEarthBranches[hourBranchIdx]
}

func proHourBranchIndex(hour int) int {
	switch {
	case hour == 23 || hour == 0:
		return 0
	default:
		return (hour + 1) / 2
	}
}

func proStemIndex(pillar string) int {
	if pillar == "" {
		return 0
	}
	runes := []rune(pillar)
	for i, stem := range proHeavenlyStems {
		if runes[0] == []rune(stem)[0] {
			return i
		}
	}
	return 0
}

func proSexagenary(idx int) string {
	idx = idx % 60
	if idx < 0 {
		idx += 60
	}
	return proHeavenlyStems[idx%10] + proEarthBranches[idx%12]
}

func proSexagenaryIndex(gz string) int {
	gz = strings.TrimSpace(gz)
	if len([]rune(gz)) != 2 {
		return -1
	}
	runes := []rune(gz)
	stemIdx := -1
	branchIdx := -1
	for i, stem := range proHeavenlyStems {
		if runes[0] == []rune(stem)[0] {
			stemIdx = i
			break
		}
	}
	for i, branch := range proEarthBranches {
		if runes[1] == []rune(branch)[0] {
			branchIdx = i
			break
		}
	}
	if stemIdx < 0 || branchIdx < 0 {
		return -1
	}
	for i := 0; i < 60; i++ {
		if i%10 == stemIdx && i%12 == branchIdx {
			return i
		}
	}
	return -1
}
