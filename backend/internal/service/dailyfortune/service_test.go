package dailyfortune_test

import (
	"errors"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/service/dailyfortune"
)

func TestGetOrCreateTodayInvalidDate(t *testing.T) {
	svc := dailyfortune.NewService(nil, nil, nil, nil)
	_, err := svc.GetOrCreateToday(t.Context(), dailyfortune.TodayInput{
		SessionKey: "test-session",
		LocalDate:  "06-23-2026",
	})
	if !errors.Is(err, dailyfortune.ErrInvalidDate) {
		t.Fatalf("expected invalid date error, got %v", err)
	}
}

func TestGetOrCreateTodayEmptySession(t *testing.T) {
	svc := dailyfortune.NewService(nil, nil, nil, nil)
	_, err := svc.GetOrCreateToday(t.Context(), dailyfortune.TodayInput{
		SessionKey: "",
		LocalDate:  "2026-06-23",
	})
	if !errors.Is(err, dailyfortune.ErrSessionKeyEmpty) {
		t.Fatalf("expected session key error, got %v", err)
	}
}
