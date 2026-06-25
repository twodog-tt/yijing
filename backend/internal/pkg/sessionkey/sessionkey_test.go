package sessionkey_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/pkg/sessionkey"
)

func TestResolveForCreatePrefersMatchingHeaderAndBody(t *testing.T) {
	key, err := sessionkey.ResolveForCreate("sess-a", "sess-a")
	if err != nil || key != "sess-a" {
		t.Fatalf("unexpected result key=%q err=%v", key, err)
	}
}

func TestResolveForCreateRejectsConflict(t *testing.T) {
	_, err := sessionkey.ResolveForCreate("sess-a", "sess-b")
	if !errors.Is(err, sessionkey.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestValidateLengthRejectsTooLongKey(t *testing.T) {
	err := sessionkey.ValidateLength(strings.Repeat("a", sessionkey.MaxLength+1))
	if !errors.Is(err, sessionkey.ErrTooLong) {
		t.Fatalf("expected too long, got %v", err)
	}
}
