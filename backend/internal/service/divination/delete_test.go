package divination

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

func newDeleteServiceWithMock(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(db)
	divRepo := repository.NewDivinationRepository(db)
	svc := NewService(divRepo, nil, nil, sessionRepo, nil, nil)
	return svc, mock, func() { _ = db.Close() }
}

func expectSessionLookup(mock sqlmock.Sqlmock, sessionKey string, sessionID int64) {
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, session_key, status, created_at
		FROM user_sessions
		WHERE session_key = ? AND status = ?
		LIMIT 1
	`)).WithArgs(sessionKey, model.SessionStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"id", "session_key", "status", "created_at"}).
			AddRow(sessionID, sessionKey, model.SessionStatusActive, time.Now()))
}

func TestDeleteOwnRecordSuccess(t *testing.T) {
	svc, mock, cleanup := newDeleteServiceWithMock(t)
	defer cleanup()

	expectSessionLookup(mock, "sess-a", 10)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE divination_record
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND session_id = ? AND status = ?
	`)).WithArgs(model.DivinationStatusDeleted, int64(5), int64(10), model.DivinationStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.Delete(context.Background(), "sess-a", 5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDeleteOtherSessionFails(t *testing.T) {
	svc, mock, cleanup := newDeleteServiceWithMock(t)
	defer cleanup()

	expectSessionLookup(mock, "sess-b", 11)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE divination_record
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND session_id = ? AND status = ?
	`)).WithArgs(model.DivinationStatusDeleted, int64(5), int64(11), model.DivinationStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(context.Background(), "sess-b", 5)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDeleteMissingRecordFails(t *testing.T) {
	svc, mock, cleanup := newDeleteServiceWithMock(t)
	defer cleanup()

	expectSessionLookup(mock, "sess-a", 10)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE divination_record
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND session_id = ? AND status = ?
	`)).WithArgs(model.DivinationStatusDeleted, int64(999), int64(10), model.DivinationStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(context.Background(), "sess-a", 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestDeleteInvalidParams(t *testing.T) {
	svc, _, cleanup := newDeleteServiceWithMock(t)
	defer cleanup()

	err := svc.Delete(context.Background(), "sess-a", 0)
	if !errors.Is(err, ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}

	err = svc.Delete(context.Background(), "", 5)
	if !errors.Is(err, ErrSessionKeyEmpty) {
		t.Fatalf("expected session key empty, got %v", err)
	}
}
