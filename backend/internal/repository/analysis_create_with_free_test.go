package repository

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/model"
)

const createWithFreeContentSQL = `
		INSERT INTO analysis_records (
			session_id, module_type, algorithm_version, category_id, question,
			input_payload, result_payload,
			free_content, full_content, unlock_status, unlock_type, ai_provider, generation_status, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL, 0, NULL, NULL, ?, 1)
	`

func TestCreateWithFreeContentAtomicInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	input := json.RawMessage(`{"birth_date":"1995-01-01","birth_hour_branch":"zi"}`)
	result := json.RawMessage(`{"algorithm_version":"bazi-simple-v1"}`)
	mock.ExpectExec(regexp.QuoteMeta(createWithFreeContentSQL)).WithArgs(
		int64(10),
		model.ModuleTypeBazi,
		model.AlgorithmVersionBaziSimpleV1,
		nil,
		nil,
		input,
		result,
		"free text",
		model.AnalysisGenerationStatusFreeDone,
	).WillReturnResult(sqlmock.NewResult(101, 1))

	repo := NewAnalysisRepository(db)
	id, err := repo.CreateWithFreeContent(context.Background(), CreateAnalysisWithFreeContentParams{
		CreateAnalysisParams: CreateAnalysisParams{
			SessionID:        10,
			ModuleType:       model.ModuleTypeBazi,
			AlgorithmVersion: model.AlgorithmVersionBaziSimpleV1,
			InputPayload:     input,
			ResultPayload:    result,
		},
		FreeContent: "free text",
	})
	if err != nil {
		t.Fatalf("CreateWithFreeContent: %v", err)
	}
	if id != 101 {
		t.Fatalf("expected id 101, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreateWithFreeContentRejectsEmptyFreeContent(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	_, err := repo.CreateWithFreeContent(context.Background(), CreateAnalysisWithFreeContentParams{
		CreateAnalysisParams: CreateAnalysisParams{
			SessionID:        1,
			ModuleType:       model.ModuleTypeBazi,
			AlgorithmVersion: model.AlgorithmVersionBaziSimpleV1,
			InputPayload:     json.RawMessage(`{"birth_date":"1995-01-01"}`),
		},
		FreeContent: "   ",
	})
	if err == nil || !errors.Is(err, ErrInvalidAnalysisParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestCreateWithFreeContentSQLHardcodesLockedDefaults(t *testing.T) {
	if !strings.Contains(createWithFreeContentSQL, "NULL, 0, NULL, NULL") {
		t.Fatalf("create with free content SQL must hardcode locked defaults")
	}
}
