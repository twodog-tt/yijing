package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/model"
)

const unlockWithFullContentSQL = `
		UPDATE analysis_records
		SET unlock_status = ?, unlock_type = ?, full_content = ?, generation_status = ?, updated_at = NOW()
		WHERE id = ? AND session_id = ? AND status = ? AND unlock_status = ?
	`

func TestUnlockWithFullContentSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(unlockWithFullContentSQL)).
		WithArgs(
			model.AnalysisUnlockStatusUnlocked,
			model.UnlockTypeRewardedVideoMock,
			"full report",
			model.AnalysisGenerationStatusFullDone,
			int64(5),
			int64(10),
			model.AnalysisStatusActive,
			model.AnalysisUnlockStatusLocked,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewAnalysisRepository(db)
	err = repo.UnlockWithFullContent(
		context.Background(),
		5,
		10,
		model.UnlockTypeRewardedVideoMock,
		"full report",
	)
	if err != nil {
		t.Fatalf("UnlockWithFullContent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUnlockWithFullContentRequiresSessionMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(unlockWithFullContentSQL)).
		WithArgs(
			model.AnalysisUnlockStatusUnlocked,
			model.UnlockTypeRewardedVideoMock,
			"full report",
			model.AnalysisGenerationStatusFullDone,
			int64(5),
			int64(99),
			model.AnalysisStatusActive,
			model.AnalysisUnlockStatusLocked,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo := NewAnalysisRepository(db)
	err = repo.UnlockWithFullContent(
		context.Background(),
		5,
		99,
		model.UnlockTypeRewardedVideoMock,
		"full report",
	)
	if !errors.Is(err, ErrAnalysisNotFound) {
		t.Fatalf("expected not found for session mismatch, got %v", err)
	}
}

func TestUnlockWithFullContentSQLScopesBySessionAndLockedStatus(t *testing.T) {
	if !regexp.MustCompile(`WHERE id = \? AND session_id = \? AND status = \? AND unlock_status = \?`).
		MatchString(unlockWithFullContentSQL) {
		t.Fatalf("unlock SQL must scope by id, session_id, status, and unlock_status")
	}
	if stringsContains(unlockWithFullContentSQL, "input_payload") || stringsContains(unlockWithFullContentSQL, "result_payload") {
		t.Fatalf("unlock SQL must not update input_payload or result_payload")
	}
}

func stringsContains(haystack, needle string) bool {
	return regexp.MustCompile(regexp.QuoteMeta(needle)).MatchString(haystack)
}

func TestUnlockWithFullContentRejectsInvalidParams(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	err := repo.UnlockWithFullContent(context.Background(), 0, 10, model.UnlockTypeRewardedVideoMock, "full")
	if !errors.Is(err, ErrInvalidAnalysisParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}
