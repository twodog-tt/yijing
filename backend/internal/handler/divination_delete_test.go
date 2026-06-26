package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wangxintong/yijing/backend/internal/handler"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/sessionkey"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/divination"
)

func newDeleteDivinationHandler(t *testing.T) (*handler.DivinationHandler, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(db)
	divRepo := repository.NewDivinationRepository(db)
	svc := divination.NewService(divRepo, nil, nil, sessionRepo, nil, nil)
	return handler.NewDivinationHandler(svc), mock, func() { _ = db.Close() }
}

func expectDivinationSessionLookup(mock sqlmock.Sqlmock, sessionKey string, sessionID int64) {
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, session_key, status, created_at
		FROM user_sessions
		WHERE session_key = ? AND status = ?
		LIMIT 1
	`)).WithArgs(sessionKey, model.SessionStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"id", "session_key", "status", "created_at"}).
			AddRow(sessionID, sessionKey, model.SessionStatusActive, time.Now()))
}

func expectDivinationSoftDelete(mock sqlmock.Sqlmock, id, sessionID int64, affected int64) {
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE divination_record
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND session_id = ? AND status = ?
	`)).WithArgs(model.DivinationStatusDeleted, id, sessionID, model.DivinationStatusActive).
		WillReturnResult(sqlmock.NewResult(0, affected))
}

func TestDeleteDivinationSuccess(t *testing.T) {
	h, mock, cleanup := newDeleteDivinationHandler(t)
	defer cleanup()

	expectDivinationSessionLookup(mock, "sess-a", 10)
	expectDivinationSoftDelete(mock, 5, 10, 1)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/divinations/5", nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "session_key") || strings.Contains(rec.Body.String(), "payload") {
		t.Fatalf("delete response must not expose sensitive fields: %s", rec.Body.String())
	}
}

func TestDeleteDivinationOtherSessionNotFound(t *testing.T) {
	h, mock, cleanup := newDeleteDivinationHandler(t)
	defer cleanup()

	expectDivinationSessionLookup(mock, "sess-b", 11)
	expectDivinationSoftDelete(mock, 5, 11, 0)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/divinations/5", nil)
	req.Header.Set(sessionkey.HeaderName, "sess-b")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for other session delete, got %d", rec.Code)
	}
}

func TestDeleteDivinationRequiresHeaderSessionKey(t *testing.T) {
	h, _, cleanup := newDeleteDivinationHandler(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/divinations/5", nil)
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteDivinationRejectsQuerySessionKey(t *testing.T) {
	h, _, cleanup := newDeleteDivinationHandler(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/divinations/5?session_key=sess-a", nil)
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteDivinationInvalidID(t *testing.T) {
	h, _, cleanup := newDeleteDivinationHandler(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/divinations/abc", nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteDivinationListExcludesDeletedRecord(t *testing.T) {
	h, mock, cleanup := newDeleteDivinationHandler(t)
	defer cleanup()

	expectDivinationSessionLookup(mock, "sess-a", 10)
	expectDivinationSoftDelete(mock, 5, 10, 1)

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/divinations/5", nil)
	delReq.Header.Set(sessionkey.HeaderName, "sess-a")
	delRec := httptest.NewRecorder()
	h.Delete(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("delete failed: %s", delRec.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, session_key, status, created_at
		FROM user_sessions
		WHERE session_key = ? AND status = ?
		LIMIT 1
	`)).WithArgs("sess-a", model.SessionStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"id", "session_key", "status", "created_at"}).
			AddRow(10, "sess-a", model.SessionStatusActive, time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM divination_record
		WHERE session_id = ? AND status = ?
	`)).WithArgs(int64(10), model.DivinationStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, session_id, category_id, question, method,
			primary_hexagram_id, changed_hexagram_id,
			moving_lines, line_snapshot, seed,
			unlock_status, status, created_at
		FROM divination_record
		WHERE session_id = ? AND status = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`)).WithArgs(int64(10), model.DivinationStatusActive, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "session_id", "category_id", "question", "method",
			"primary_hexagram_id", "changed_hexagram_id",
			"moving_lines", "line_snapshot", "seed",
			"unlock_status", "status", "created_at",
		}))

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/divinations?session_key=sess-a", nil)
	listRec := httptest.NewRecorder()
	h.List(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list failed: %s", listRec.Body.String())
	}

	var listResp struct {
		Data struct {
			Items []map[string]any `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Data.Items) != 0 {
		t.Fatalf("expected empty list after delete")
	}
}

func TestDeleteDivinationMissingRecord(t *testing.T) {
	h, mock, cleanup := newDeleteDivinationHandler(t)
	defer cleanup()

	expectDivinationSessionLookup(mock, "sess-a", 10)
	expectDivinationSoftDelete(mock, 999, 10, 0)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/divinations/"+strconv.FormatInt(999, 10), nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteAnalysisStillWorksAfterDivinationDeleteFeature(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/analysis/"+strconv.FormatInt(id, 10), nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected analysis delete to remain working, got %d", rec.Code)
	}
}
