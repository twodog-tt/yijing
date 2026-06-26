package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/model"
)

const softDeleteOwnedDivinationSQL = `
		UPDATE divination_record
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND session_id = ? AND status = ?
	`

func TestSoftDeleteOwnedByIDSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(softDeleteOwnedDivinationSQL)).
		WithArgs(model.DivinationStatusDeleted, int64(5), int64(10), model.DivinationStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewDivinationRepository(db)
	if err := repo.SoftDeleteOwnedByID(context.Background(), 5, 10); err != nil {
		t.Fatalf("SoftDeleteOwnedByID: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSoftDeleteOwnedByIDRequiresSessionMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(softDeleteOwnedDivinationSQL)).
		WithArgs(model.DivinationStatusDeleted, int64(5), int64(99), model.DivinationStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo := NewDivinationRepository(db)
	err = repo.SoftDeleteOwnedByID(context.Background(), 5, 99)
	if !errors.Is(err, ErrDivinationNotFound) {
		t.Fatalf("expected not found for session mismatch, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSoftDeleteOwnedByIDUsesSoftDeleteSQL(t *testing.T) {
	if regexp.MustCompile(`DELETE FROM divination_record`).MatchString(softDeleteOwnedDivinationSQL) {
		t.Fatalf("must not physically delete divination records")
	}
	if !regexp.MustCompile(`UPDATE divination_record`).MatchString(softDeleteOwnedDivinationSQL) {
		t.Fatalf("expected soft delete UPDATE SQL")
	}
	if !regexp.MustCompile(`WHERE id = \? AND session_id = \? AND status = \?`).MatchString(softDeleteOwnedDivinationSQL) {
		t.Fatalf("delete SQL must scope by id, session_id, and status")
	}
}

func TestSoftDeleteOwnedByIDRejectsInvalidParams(t *testing.T) {
	repo := NewDivinationRepository(nil)
	err := repo.SoftDeleteOwnedByID(context.Background(), 0, 10)
	if !errors.Is(err, ErrInvalidDivinationParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}
