package qimen

import (
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func TestResolveProfessionalGanZhiFieldsPresent(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	gz := ResolveProfessionalGanZhi(when)
	for _, field := range []string{gz.Year, gz.Month, gz.Day, gz.Hour} {
		if len([]rune(field)) != 2 {
			t.Fatalf("invalid pillar %q", field)
		}
	}
	if gz.Basis == "" || gz.Note == "" {
		t.Fatalf("missing basis/note: %+v", gz)
	}
}

func TestResolveProfessionalGanZhiStable(t *testing.T) {
	when := time.Date(2024, 9, 22, 18, 30, 0, 0, clock.Location())
	a := ResolveProfessionalGanZhi(when)
	b := ResolveProfessionalGanZhi(when)
	if a != b {
		t.Fatalf("not stable: %+v vs %+v", a, b)
	}
}

func TestResolveProfessionalGanZhiHourChangesWithClock(t *testing.T) {
	day := time.Date(2024, 6, 20, 0, 0, 0, 0, clock.Location())
	morning := day.Add(9 * time.Hour)
	evening := day.Add(21 * time.Hour)
	gMorning := ResolveProfessionalGanZhi(morning)
	gEvening := ResolveProfessionalGanZhi(evening)
	if gMorning.Hour == gEvening.Hour {
		t.Fatalf("expected different hour pillars for 09:00 vs 21:00, got %q", gMorning.Hour)
	}
	if gMorning.Day != gEvening.Day {
		t.Fatalf("same day should share day pillar: %q vs %q", gMorning.Day, gEvening.Day)
	}
}

func TestResolveXunFromGanZhiUsesHourPillar(t *testing.T) {
	xun := ResolveXunFromGanZhi("甲子", "戊寅")
	if xun.XunShou != "甲戌" {
		t.Fatalf("xun_shou=%q want 甲戌", xun.XunShou)
	}
	if len(xun.EmptyBranches) != 2 {
		t.Fatalf("empty_branches=%v", xun.EmptyBranches)
	}
}

func TestResolveXunFromGanZhiKnownCycle(t *testing.T) {
	xun := ResolveXunFromGanZhi("ignored", "甲戌")
	if xun.XunShou != "甲戌" {
		t.Fatalf("xun_shou=%q want 甲戌", xun.XunShou)
	}
	if xun.EmptyBranches[0] != "申" || xun.EmptyBranches[1] != "酉" {
		t.Fatalf("empty=%v", xun.EmptyBranches)
	}
}

func TestResolveXunFromGanZhiDoesNotUseCategoryHash(t *testing.T) {
	when := time.Date(2024, 2, 4, 10, 30, 0, 0, clock.Location())
	gz := ResolveProfessionalGanZhi(when)
	xunA := ResolveXunFromGanZhi(gz.Day, gz.Hour)
	xunB := ResolveXunFromGanZhi(gz.Day, gz.Hour)
	if xunA.XunShou != xunB.XunShou {
		t.Fatalf("xun should be deterministic from ganzhi")
	}
	if xunA.XunShou == "" || xunA.XunShou == "professional_pending" {
		t.Fatalf("expected derived xun_shou, got %q", xunA.XunShou)
	}
}
