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
			name: "day_after_lichun",
			at:   time.Date(2024, 2, 5, 12, 0, 0, 0, loc),
			want: 2024,
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
			name: "after_daxue_zi",
			at:   time.Date(2024, 12, 9, 12, 0, 0, 0, loc),
			want: 0,
		},
		{
			name: "after_xiaohan_chou",
			at:   time.Date(2025, 1, 6, 12, 0, 0, 0, loc),
			want: 1,
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
