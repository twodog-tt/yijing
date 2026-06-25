package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/model"
)

const deleteOwnedByIDSQL = `
		DELETE FROM analysis_records
		WHERE id = ? AND session_id = ? AND status = ?
	`

func TestDeleteOwnedByIDSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(deleteOwnedByIDSQL)).
		WithArgs(int64(5), int64(10), model.AnalysisStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewAnalysisRepository(db)
	if err := repo.DeleteOwnedByID(context.Background(), 5, 10); err != nil {
		t.Fatalf("DeleteOwnedByID: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDeleteOwnedByIDRequiresSessionMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(deleteOwnedByIDSQL)).
		WithArgs(int64(5), int64(99), model.AnalysisStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo := NewAnalysisRepository(db)
	err = repo.DeleteOwnedByID(context.Background(), 5, 99)
	if !errors.Is(err, ErrAnalysisNotFound) {
		t.Fatalf("expected not found for session mismatch, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDeleteOwnedByIDUsesHardDeleteSQL(t *testing.T) {
	if !regexp.MustCompile(`DELETE FROM analysis_records`).MatchString(deleteOwnedByIDSQL) {
		t.Fatalf("expected hard delete SQL")
	}
	if !regexp.MustCompile(`WHERE id = \? AND session_id = \? AND status = \?`).MatchString(deleteOwnedByIDSQL) {
		t.Fatalf("delete SQL must scope by id, session_id, and status")
	}
}

func TestDeleteOwnedByIDRejectsInvalidParams(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	err := repo.DeleteOwnedByID(context.Background(), 0, 10)
	if !errors.Is(err, ErrInvalidAnalysisParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}
