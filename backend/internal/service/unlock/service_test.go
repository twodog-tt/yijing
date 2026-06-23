package unlock

import (
	"testing"

	"github.com/wangxintong/yijing/backend/internal/model"
)

func TestValidateSessionAccess(t *testing.T) {
	tests := []struct {
		name      string
		divSession int64
		reqSession *model.Session
		wantErr   error
	}{
		{
			name:       "match",
			divSession: 10,
			reqSession: &model.Session{ID: 10, SessionKey: "a"},
			wantErr:    nil,
		},
		{
			name:       "mismatch",
			divSession: 10,
			reqSession: &model.Session{ID: 99, SessionKey: "b"},
			wantErr:    ErrForbidden,
		},
		{
			name:       "nil session",
			divSession: 10,
			reqSession: nil,
			wantErr:    ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSessionAccess(tt.divSession, tt.reqSession)
			if err != tt.wantErr {
				t.Fatalf("validateSessionAccess() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnlockStatusGate(t *testing.T) {
	if err := validateUnlocked(model.UnlockStatusLocked); err != ErrNotUnlocked {
		t.Fatalf("locked should return ErrNotUnlocked, got %v", err)
	}
	if err := validateUnlocked(model.UnlockStatusUnlocked); err != nil {
		t.Fatalf("unlocked should pass, got %v", err)
	}
}
