package calendar_test

import (
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/bazi/calendar"
)

func TestLiChunTimeApproximation(t *testing.T) {
	loc := clock.Location()
	cases := []struct {
		year int
		want time.Time
	}{
		{1995, time.Date(1995, 2, 4, 12, 0, 0, 0, loc)},
		{2024, time.Date(2024, 2, 4, 12, 0, 0, 0, loc)},
	}
	for _, tc := range cases {
		got := calendar.LiChunTime(tc.year)
		if !got.Equal(tc.want) {
			t.Fatalf("LiChunTime(%d)=%v want %v", tc.year, got, tc.want)
		}
	}
}

// TestJieTimeApproximation2024 locks formula outputs for the twelve 节 used by bazi-v2-poc.
// These dates come from the century-calendar approximation in jieqi.go, not observatory ephemeris.
func TestJieTimeApproximation2024(t *testing.T) {
	loc := clock.Location()
	cases := []struct {
		jie  calendar.Jie
		want time.Time
	}{
		{calendar.JieXiaoHan, time.Date(2024, 1, 6, 12, 0, 0, 0, loc)},
		{calendar.JieLiChun, time.Date(2024, 2, 4, 12, 0, 0, 0, loc)},
		{calendar.JieJingZhe, time.Date(2024, 3, 6, 12, 0, 0, 0, loc)},
		{calendar.JieQingMing, time.Date(2024, 4, 5, 12, 0, 0, 0, loc)},
		{calendar.JieLiXia, time.Date(2024, 5, 6, 12, 0, 0, 0, loc)},
		{calendar.JieMangZhong, time.Date(2024, 6, 6, 12, 0, 0, 0, loc)},
		{calendar.JieXiaoShu, time.Date(2024, 7, 8, 12, 0, 0, 0, loc)},
		{calendar.JieLiQiu, time.Date(2024, 8, 8, 12, 0, 0, 0, loc)},
		{calendar.JieBaiLu, time.Date(2024, 9, 9, 12, 0, 0, 0, loc)},
		{calendar.JieHanLu, time.Date(2024, 10, 9, 12, 0, 0, 0, loc)},
		{calendar.JieLiDong, time.Date(2024, 11, 8, 12, 0, 0, 0, loc)},
		{calendar.JieDaXue, time.Date(2024, 12, 7, 12, 0, 0, 0, loc)},
	}
	for _, tc := range cases {
		got := calendar.JieTime(2024, tc.jie)
		if !got.Equal(tc.want) {
			t.Fatalf("JieTime(2024,%d)=%v want %v", tc.jie, got, tc.want)
		}
	}
}

func TestJieTimeFormulaIsLocalNoonApproximation(t *testing.T) {
	at := calendar.JieTime(2024, calendar.JieLiChun)
	if at.Hour() != 12 || at.Minute() != 0 {
		t.Fatalf("JieTime should use local noon approximation, got %v", at)
	}
	// Published almanac 2024 立春 is also Feb 4, but exact clock time differs; we only pin the formula day.
	if at.Month() != time.February || at.Day() != 4 {
		t.Fatalf("unexpected lichun formula day: %v", at)
	}
}

func TestBaziYearLiChunBoundary(t *testing.T) {
	loc := clock.Location()
	cases := []struct {
		name string
		at   time.Time
		want int
	}{
		{
			name: "day_before_lichun",
			at:   time.Date(2024, 2, 3, 12, 0, 0, 0, loc),
			want: 2023,
		},
		{
			name: "lichun_formula_day_at_noon",
			at:   time.Date(2024, 2, 4, 12, 0, 0, 0, loc),
			want: 2024,
		},
		{
			name: "lichun_formula_day_before_noon",
			at:   time.Date(2024, 2, 4, 11, 0, 0, 0, loc),
			want: 2023,
		},
		{
			name: "day_after_lichun",
			at:   time.Date(2024, 2, 5, 12, 0, 0, 0, loc),
			want: 2024,
		},
		{
			name: "1995_day_before_lichun",
			at:   time.Date(1995, 2, 3, 12, 0, 0, 0, loc),
			want: 1994,
		},
		{
			name: "1995_lichun_formula_day",
			at:   time.Date(1995, 2, 4, 12, 0, 0, 0, loc),
			want: 1995,
		},
		{
			name: "gregorian_new_year_before_lichun",
			at:   time.Date(1995, 1, 1, 12, 0, 0, 0, loc),
			want: 1994,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := calendar.BaziYear(tc.at); got != tc.want {
				t.Fatalf("BaziYear=%d want %d", got, tc.want)
			}
		})
	}
}

func TestMonthBranchSolarTerms(t *testing.T) {
	loc := clock.Location()
	cases := []struct {
		name string
		at   time.Time
		want int
	}{
		{
			name: "before_jingzhe_still_yin",
			at:   time.Date(2024, 3, 5, 12, 0, 0, 0, loc),
			want: 2, // 寅
		},
		{
			name: "after_jingzhe_mao",
			at:   time.Date(2024, 3, 7, 12, 0, 0, 0, loc),
			want: 3, // 卯
		},
		{
			name: "before_qingming_mao",
			at:   time.Date(2024, 4, 4, 12, 0, 0, 0, loc),
			want: 3,
		},
		{
			name: "after_qingming_chen",
			at:   time.Date(2024, 4, 6, 12, 0, 0, 0, loc),
			want: 4,
		},
		{
			name: "before_xiaoshu_still_wu",
			at:   time.Date(2024, 7, 7, 12, 0, 0, 0, loc),
			want: 6, // 午
		},
		{
			name: "after_xiaoshu_wei",
			at:   time.Date(2024, 7, 9, 12, 0, 0, 0, loc),
			want: 7, // 未
		},
		{
			name: "before_liqiu_still_wei",
			at:   time.Date(2024, 8, 7, 12, 0, 0, 0, loc),
			want: 7,
		},
		{
			name: "after_liqiu_shen",
			at:   time.Date(2024, 8, 9, 12, 0, 0, 0, loc),
			want: 8, // 申
		},
		{
			name: "before_bailu_still_shen",
			at:   time.Date(2024, 9, 8, 12, 0, 0, 0, loc),
			want: 8,
		},
		{
			name: "after_bailu_you",
			at:   time.Date(2024, 9, 10, 12, 0, 0, 0, loc),
			want: 9, // 酉
		},
		{
			name: "before_hanlu_still_you",
			at:   time.Date(2024, 10, 8, 12, 0, 0, 0, loc),
			want: 9,
		},
		{
			name: "after_hanlu_xu",
			at:   time.Date(2024, 10, 10, 12, 0, 0, 0, loc),
			want: 10, // 戌
		},
		{
			name: "before_lidong_still_xu",
			at:   time.Date(2024, 11, 7, 12, 0, 0, 0, loc),
			want: 10,
		},
		{
			name: "after_lidong_hai",
			at:   time.Date(2024, 11, 9, 12, 0, 0, 0, loc),
			want: 11, // 亥
		},
		{
			name: "before_daxue_still_hai",
			at:   time.Date(2024, 12, 6, 12, 0, 0, 0, loc),
			want: 11,
		},
		{
			name: "after_daxue_zi",
			at:   time.Date(2024, 12, 9, 12, 0, 0, 0, loc),
			want: 0, // 子
		},
		{
			name: "before_xiaohan_still_zi",
			at:   time.Date(2025, 1, 4, 12, 0, 0, 0, loc),
			want: 0,
		},
		{
			name: "after_xiaohan_chou",
			at:   time.Date(2025, 1, 6, 12, 0, 0, 0, loc),
			want: 1, // 丑
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := calendar.MonthBranchIndex(tc.at); got != tc.want {
				t.Fatalf("MonthBranchIndex=%d want %d", got, tc.want)
			}
		})
	}
}
