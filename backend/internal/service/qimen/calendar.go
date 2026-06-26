package qimen

import (
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/bazi/calendar"
)

const (
	jieqiBasisPOC     = "formula_approximation"
	timeBasisLocal    = "local_time"
	calendarNotePOC   = "当前为 POC 近似口径，节令时刻为公式近似，非天文台精确时刻"
	dunSourcePOC      = "poc_formula"
)

// CalendarBasis documents how v2 qimen timing was derived.
type CalendarBasis struct {
	SolarTerm   string `json:"solar_term"`
	JieqiBasis  string `json:"jieqi_basis"`
	TimeBasis   string `json:"time_basis"`
	Note        string `json:"note"`
}

// Dun describes yin/yang dun and ju for POC.
type Dun struct {
	Type   string `json:"type"`
	Ju     int    `json:"ju"`
	Source string `json:"source"`
}

// Xun holds simplified 旬首 and 空亡 branches for POC.
type Xun struct {
	XunShou       string   `json:"xun_shou"`
	EmptyBranches []string `json:"empty_branches"`
}

// Chief holds 值符 / 值使 placeholders for POC.
type Chief struct {
	ZhiFu string `json:"zhi_fu"`
	ZhiShi string `json:"zhi_shi"`
}

func normalizeMoment(t time.Time) time.Time {
	if t.IsZero() {
		t = clock.Now()
	}
	return t.In(clock.Location())
}

func currentJieName(t time.Time) string {
	t = normalizeMoment(t)

	var best time.Time
	name := jieDisplayName(calendar.JieXiaoHan)
	found := false
	for y := t.Year() - 1; y <= t.Year()+1; y++ {
		for jie := calendar.JieXiaoHan; jie <= calendar.JieDaXue; jie++ {
			at := calendar.JieTime(y, jie)
			if at.After(t) {
				continue
			}
			if !found || at.After(best) {
				best = at
				name = jieDisplayName(jie)
				found = true
			}
		}
	}
	return name
}

func jieDisplayName(jie calendar.Jie) string {
	switch jie {
	case calendar.JieXiaoHan:
		return "小寒"
	case calendar.JieLiChun:
		return "立春"
	case calendar.JieJingZhe:
		return "惊蛰"
	case calendar.JieQingMing:
		return "清明"
	case calendar.JieLiXia:
		return "立夏"
	case calendar.JieMangZhong:
		return "芒种"
	case calendar.JieXiaoShu:
		return "小暑"
	case calendar.JieLiQiu:
		return "立秋"
	case calendar.JieBaiLu:
		return "白露"
	case calendar.JieHanLu:
		return "寒露"
	case calendar.JieLiDong:
		return "立冬"
	case calendar.JieDaXue:
		return "大雪"
	default:
		return "小寒"
	}
}

// isYangDunPOC uses simplified 冬至后阳遁、夏至后阴遁 (approximate Gregorian boundaries).
func isYangDunPOC(t time.Time) bool {
	t = normalizeMoment(t)
	m := int(t.Month())
	d := t.Day()
	if m == 12 && d >= 22 {
		return true
	}
	if m >= 1 && m <= 5 {
		return true
	}
	if m == 6 && d < 21 {
		return true
	}
	return false
}

func dunForMoment(t time.Time, category string) Dun {
	yang := isYangDunPOC(t)
	dunType := "yin"
	if yang {
		dunType = "yang"
	}
	ju := juForMoment(t, category, yang)
	return Dun{
		Type:   dunType,
		Ju:     ju,
		Source: dunSourcePOC,
	}
}

func juForMoment(t time.Time, category string, yang bool) int {
	t = normalizeMoment(t)
	seed := hashSeed(
		t.Format(time.RFC3339),
		NormalizeCategory(category),
		boolString(yang),
	)
	ju := int(seed%9) + 1
	if ju < 1 || ju > 9 {
		ju = 1
	}
	return ju
}

func boolString(v bool) string {
	if v {
		return "yang"
	}
	return "yin"
}

func calendarBasisFor(t time.Time) CalendarBasis {
	t = normalizeMoment(t)
	return CalendarBasis{
		SolarTerm:  currentJieName(t),
		JieqiBasis: jieqiBasisPOC,
		TimeBasis:  timeBasisLocal,
		Note:       calendarNotePOC,
	}
}

var (
	xunShouList = []string{"甲子", "甲戌", "甲申", "甲午", "甲辰", "甲寅"}
	xunEmpty    = [][]string{
		{"戌", "亥"},
		{"申", "酉"},
		{"午", "未"},
		{"辰", "巳"},
		{"寅", "卯"},
		{"子", "丑"},
	}
)

func xunForMoment(t time.Time, category string) Xun {
	t = normalizeMoment(t)
	idx := int(hashSeed(t.Format("2006-01-02"), NormalizeCategory(category)) % uint32(len(xunShouList)))
	return Xun{
		XunShou:       xunShouList[idx],
		EmptyBranches: append([]string(nil), xunEmpty[idx]...),
	}
}

func chiefFor(dun Dun, palaces []Palace) Chief {
	zhiFu := "天禽"
	zhiShi := "开门"
	if len(palaces) > 0 {
		center := palaces[4]
		if center.Star != "" {
			zhiFu = center.Star
		}
	}
	for _, p := range palaces {
		if p.Index == dun.Ju {
			if p.Door != "" && p.Door != "—" {
				zhiShi = p.Door
			}
			if p.Star != "" && p.Star != "—" {
				zhiFu = p.Star
			}
			break
		}
	}
	return Chief{
		ZhiFu:  zhiFu,
		ZhiShi: zhiShi,
	}
}
