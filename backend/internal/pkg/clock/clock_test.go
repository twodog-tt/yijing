package clock_test

import (
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func TestNowUsesShanghaiOffset(t *testing.T) {
	now := clock.Now()
	_, offset := now.Zone()
	if offset != 8*3600 {
		t.Fatalf("expected UTC+8 offset, got %d", offset)
	}
}

func TestFormatRFC3339IncludesShanghaiOffset(t *testing.T) {
	formatted := clock.FormatRFC3339(time.Date(2026, 6, 23, 14, 3, 32, 0, time.UTC))
	if !strings.Contains(formatted, "+08:00") {
		t.Fatalf("expected +08:00 in %q", formatted)
	}
}
