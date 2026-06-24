package repository

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/model"
)

const createInsertSQL = `
		INSERT INTO analysis_records (
			session_id, module_type, algorithm_version, category_id, question,
			input_payload, result_payload,
			free_content, full_content, unlock_status, unlock_type, ai_provider, generation_status, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULL, NULL, 0, NULL, NULL, 0, 1)
	`

func TestCreateAnalysisSuccessUsesDefaultLockedState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	input := json.RawMessage(`{"birth_date":"1990-05-20","birth_hour_branch":"unknown"}`)
	mock.ExpectExec(regexp.QuoteMeta(createInsertSQL)).WithArgs(
		int64(10),
		model.ModuleTypeBazi,
		model.AlgorithmVersionBaziSimpleV1,
		nil,
		nil,
		input,
		nil,
	).WillReturnResult(sqlmock.NewResult(101, 1))

	repo := NewAnalysisRepository(db)
	id, err := repo.Create(context.Background(), CreateAnalysisParams{
		SessionID:        10,
		ModuleType:       model.ModuleTypeBazi,
		AlgorithmVersion: model.AlgorithmVersionBaziSimpleV1,
		InputPayload:     input,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 101 {
		t.Fatalf("expected id 101, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreateAnalysisRejectsInvalidModuleType(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	_, err := repo.Create(context.Background(), CreateAnalysisParams{
		SessionID:        1,
		ModuleType:       99,
		AlgorithmVersion: model.AlgorithmVersionBaziSimpleV1,
		InputPayload:     json.RawMessage(`{"birth_date":"1990-05-20"}`),
	})
	if !errors.Is(err, model.ErrInvalidModuleType) {
		t.Fatalf("expected invalid module type, got %v", err)
	}
}

func TestCreateAnalysisRejectsMismatchedAlgorithmVersion(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	_, err := repo.Create(context.Background(), CreateAnalysisParams{
		SessionID:        1,
		ModuleType:       model.ModuleTypeBazi,
		AlgorithmVersion: model.AlgorithmVersionQimenSimpleV1,
		InputPayload:     json.RawMessage(`{"birth_date":"1990-05-20"}`),
	})
	if !errors.Is(err, model.ErrInvalidAlgorithmVersion) {
		t.Fatalf("expected invalid algorithm version, got %v", err)
	}
}

func TestCreateAnalysisRejectsNonObjectInputPayload(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	for _, raw := range []json.RawMessage{[]byte(`null`), []byte(`[]`), []byte(`"x"`)} {
		_, err := repo.Create(context.Background(), CreateAnalysisParams{
			SessionID:        1,
			ModuleType:       model.ModuleTypeBazi,
			AlgorithmVersion: model.AlgorithmVersionBaziSimpleV1,
			InputPayload:     raw,
		})
		if !errors.Is(err, ErrInvalidAnalysisParams) {
			t.Fatalf("expected invalid params for %s, got %v", string(raw), err)
		}
	}
}

func TestFindOwnedByIDRequiresSessionMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	input := json.RawMessage(`{"birth_date":"1990-05-20"}`)
	mock.ExpectQuery(`(?s).*WHERE id = \? AND session_id = \? AND status = \?.*`).
		WithArgs(int64(5), int64(10), model.AnalysisStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "session_id", "module_type", "algorithm_version", "category_id", "question",
			"input_payload", "result_payload", "free_content", "full_content",
			"unlock_status", "unlock_type", "ai_provider", "generation_status", "status",
			"created_at", "updated_at",
		}).AddRow(
			5, 10, model.ModuleTypeBazi, model.AlgorithmVersionBaziSimpleV1, nil, nil,
			input, nil, nil, nil,
			model.AnalysisUnlockStatusLocked, nil, nil, model.AnalysisGenerationStatusPending, model.AnalysisStatusActive,
			now, now,
		))

	repo := NewAnalysisRepository(db)
	record, err := repo.FindOwnedByID(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("FindOwnedByID: %v", err)
	}
	if record == nil || record.ID != 5 || record.SessionID != 10 {
		t.Fatalf("unexpected record: %#v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestFindOwnedByIDSessionMismatchReturnsNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s).*WHERE id = \? AND session_id = \? AND status = \?.*`).
		WithArgs(int64(5), int64(99), model.AnalysisStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "session_id", "module_type", "algorithm_version", "category_id", "question",
			"input_payload", "result_payload", "free_content", "full_content",
			"unlock_status", "unlock_type", "ai_provider", "generation_status", "status",
			"created_at", "updated_at",
		}))

	repo := NewAnalysisRepository(db)
	record, err := repo.FindOwnedByID(context.Background(), 5, 99)
	if err != nil {
		t.Fatalf("FindOwnedByID: %v", err)
	}
	if record != nil {
		t.Fatalf("expected nil record for mismatched session, got %#v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestFindOwnedByIDDeletedStatusNotReturned(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s).*WHERE id = \? AND session_id = \? AND status = \?.*`).
		WithArgs(int64(5), int64(10), model.AnalysisStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "session_id", "module_type", "algorithm_version", "category_id", "question",
			"input_payload", "result_payload", "free_content", "full_content",
			"unlock_status", "unlock_type", "ai_provider", "generation_status", "status",
			"created_at", "updated_at",
		}))

	repo := NewAnalysisRepository(db)
	record, err := repo.FindOwnedByID(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("FindOwnedByID: %v", err)
	}
	if record != nil {
		t.Fatalf("expected nil for deleted/missing record, got %#v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestListBySessionDefaultPageSize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs(int64(10), model.AnalysisStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	now := time.Now()
	mock.ExpectQuery("SELECT id, module_type, algorithm_version").
		WithArgs(int64(10), model.AnalysisStatusActive, 20, int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "module_type", "algorithm_version", "category_id", "question",
			"unlock_status", "generation_status", "created_at",
		}).AddRow(
			1, model.ModuleTypeBazi, model.AlgorithmVersionBaziSimpleV1, nil, nil,
			model.AnalysisUnlockStatusLocked, model.AnalysisGenerationStatusPending, now,
		))

	repo := NewAnalysisRepository(db)
	items, total, page, pageSize, err := repo.ListBySession(context.Background(), 10, nil, 0, 0)
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if total != 1 || len(items) != 1 || page != 1 || pageSize != 20 {
		t.Fatalf("unexpected list result total=%d len=%d page=%d pageSize=%d", total, len(items), page, pageSize)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestComputeAnalysisPaginationLimits(t *testing.T) {
	_, pageSize, _, err := computeAnalysisPagination(1, 500)
	if err != nil || pageSize != model.MaxAnalysisPageSize {
		t.Fatalf("expected page_size capped to 100, got %d err=%v", pageSize, err)
	}

	_, _, offset, err := computeAnalysisPagination(2, 20)
	if err != nil || offset != 20 {
		t.Fatalf("expected offset 20, got %d err=%v", offset, err)
	}

	_, _, _, err = computeAnalysisPagination(model.MaxAnalysisPage+1, 20)
	if !errors.Is(err, ErrInvalidAnalysisParams) {
		t.Fatalf("expected page limit error, got %v", err)
	}
}

func TestListBySessionModuleFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	moduleType := model.ModuleTypeQimen
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs(int64(10), model.AnalysisStatusActive, moduleType).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT id, module_type, algorithm_version").
		WithArgs(int64(10), model.AnalysisStatusActive, moduleType, 20, int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "module_type", "algorithm_version", "category_id", "question",
			"unlock_status", "generation_status", "created_at",
		}))

	repo := NewAnalysisRepository(db)
	_, _, _, _, err = repo.ListBySession(context.Background(), 10, &moduleType, 1, 20)
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestListBySessionRejectsUnknownModule(t *testing.T) {
	repo := NewAnalysisRepository(nil)
	unknown := 99
	_, _, _, _, err := repo.ListBySession(context.Background(), 10, &unknown, 1, 20)
	if !errors.Is(err, model.ErrInvalidModuleType) {
		t.Fatalf("expected invalid module type, got %v", err)
	}
}

func TestAnalysisListItemHasNoSensitiveFields(t *testing.T) {
	itemType := reflect.TypeOf(model.AnalysisListItem{})
	for _, field := range []string{"InputPayload", "ResultPayload", "FreeContent", "FullContent"} {
		if _, ok := itemType.FieldByName(field); ok {
			t.Fatalf("AnalysisListItem must not contain field %s", field)
		}
	}
}

func TestValidateCreateAnalysisParamsTrimsAlgorithmVersion(t *testing.T) {
	version, err := validateCreateAnalysisParams(CreateAnalysisParams{
		SessionID:        1,
		ModuleType:       model.ModuleTypeQimen,
		AlgorithmVersion: "  " + model.AlgorithmVersionQimenSimpleV1 + "  ",
		InputPayload:     json.RawMessage(`{"question":"test"}`),
	})
	if err != nil {
		t.Fatalf("validateCreateAnalysisParams: %v", err)
	}
	if version != model.AlgorithmVersionQimenSimpleV1 {
		t.Fatalf("expected trimmed version, got %q", version)
	}
}

func TestCreateInsertSQLHardcodesDefaults(t *testing.T) {
	if !strings.Contains(createInsertSQL, "NULL, NULL, 0, NULL, NULL, 0, 1") {
		t.Fatalf("create SQL must hardcode locked/default state literals")
	}
}
