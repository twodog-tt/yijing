// Package calendar implements traditional Chinese calendar boundaries for bazi-v2 POC.
// Rules are an engineering approximation for cultural learning reference only,
// not professional fortune-telling accuracy.
package calendar

import (
	"math"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

const termHour = 12 // 节令时刻按本地正午近似，birth_date 无具体时刻时使用同一口径。

// Jie marks the twelve 节 (month-start solar terms) used for 月令.
type Jie int

const (
	JieXiaoHan Jie = iota
	JieLiChun
	JieJingZhe
	JieQingMing
	JieLiXia
	JieMangZhong
	JieXiaoShu
	JieLiQiu
	JieBaiLu
	JieHanLu
	JieLiDong
	JieDaXue
)

type jieSpec struct {
	name      string
	month     int
	c20       float64
	c21       float64
	branchIdx int
}

// monthJieOrder follows 丑(小寒) → … → 子(大雪) in a solar year cycle.
var monthJieOrder = []jieSpec{
	{name: "xiaohan", month: 1, c20: 6.11, c21: 5.4055, branchIdx: 1},
	{name: "lichun", month: 2, c20: 4.6295, c21: 3.87, branchIdx: 2},
	{name: "jingzhe", month: 3, c20: 6.382, c21: 5.63, branchIdx: 3},
	{name: "qingming", month: 4, c20: 5.59, c21: 4.81, branchIdx: 4},
	{name: "lixia", month: 5, c20: 6.318, c21: 5.59, branchIdx: 5},
	{name: "mangzhong", month: 6, c20: 6.5, c21: 6.182, branchIdx: 6},
	{name: "xiaoshu", month: 7, c20: 7.928, c21: 7.928, branchIdx: 7},
	{name: "liqiu", month: 8, c20: 8.35, c21: 7.646, branchIdx: 8},
	{name: "bailu", month: 9, c20: 8.44, c21: 8.318, branchIdx: 9},
	{name: "hanlu", month: 10, c20: 9.098, c21: 8.318, branchIdx: 10},
	{name: "lidong", month: 11, c20: 8.218, c21: 7.438, branchIdx: 11},
	{name: "daxue", month: 12, c20: 7.9, c21: 7.18, branchIdx: 0},
}

// JieTime returns the approximate local time of a 节 in the given Gregorian year.
// Formula: [Y*D + C] - L (Y = year % 100, D = 0.2422, L = floor((Y-1)/4)).
func JieTime(year int, jie Jie) time.Time {
	spec := monthJieOrder[jie]
	y := year % 100
	c := spec.c20
	if year >= 2000 {
		c = spec.c21
	}
	day := int(math.Floor(float64(y)*0.2422+c)) - ((y - 1) / 4)
	if day < 1 {
		day = 1
	}
	loc := clock.Location()
	return time.Date(year, time.Month(spec.month), day, termHour, 0, 0, 0, loc)
}

// LiChunTime is the 立春 boundary for year-pillar change.
func LiChunTime(year int) time.Time {
	return JieTime(year, JieLiChun)
}

// MonthBranchIndex returns the 月支 index (0=子 … 11=亥) at instant t.
func MonthBranchIndex(t time.Time) int {
	loc := clock.Location()
	t = t.In(loc)

	var best time.Time
	branchIdx := monthJieOrder[0].branchIdx
	found := false
	for y := t.Year() - 1; y <= t.Year()+1; y++ {
		for i, spec := range monthJieOrder {
			at := JieTime(y, Jie(i))
			if at.After(t) {
				continue
			}
			if !found || at.After(best) {
				best = at
				branchIdx = spec.branchIdx
				found = true
			}
		}
	}
	return branchIdx
}

// BaziYear returns the Gregorian year used for 年柱 after 立春换年.
func BaziYear(t time.Time) int {
	loc := clock.Location()
	t = t.In(loc)
	y := t.Year()
	if t.Before(LiChunTime(y)) {
		return y - 1
	}
	return y
}
