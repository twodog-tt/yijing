package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/model"
)

func TestUpdateFreeContentRequiresSessionMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE analysis_records
		SET free_content = ?, generation_status = ?, updated_at = NOW()
		WHERE id = ? AND session_id = ? AND status = ?
	`)).WithArgs(
		"free text",
		model.AnalysisGenerationStatusFreeDone,
		int64(5),
		int64(10),
		model.AnalysisStatusActive,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewAnalysisRepository(db)
	if err := repo.UpdateFreeContent(context.Background(), 5, 10, "free text"); err != nil {
		t.Fatalf("UpdateFreeContent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpdateFreeContentSessionMismatchNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE analysis_records
		SET free_content = ?, generation_status = ?, updated_at = NOW()
		WHERE id = ? AND session_id = ? AND status = ?
	`)).WithArgs(
		"free text",
		model.AnalysisGenerationStatusFreeDone,
		int64(5),
		int64(99),
		model.AnalysisStatusActive,
	).WillReturnResult(sqlmock.NewResult(0, 0))

	repo := NewAnalysisRepository(db)
	err = repo.UpdateFreeContent(context.Background(), 5, 99, "free text")
	if err == nil || !errors.Is(err, ErrInvalidAnalysisParams) {
		t.Fatalf("expected invalid params for session mismatch, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpdateFreeContentRejectsEmptyContent(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	err := repo.UpdateFreeContent(context.Background(), 1, 1, "   ")
	if err == nil || !errors.Is(err, ErrInvalidAnalysisParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}
