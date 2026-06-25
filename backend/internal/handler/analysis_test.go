package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/handler"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/sessionkey"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/analysis"
	"github.com/wangxintong/yijing/backend/internal/service/bazi"
	"github.com/wangxintong/yijing/backend/internal/service/session"
)

type stubSessionRepo struct {
	sessions map[string]*model.Session
}

func (s *stubSessionRepo) Upsert(_ context.Context, sessionKey, _ string) (*model.Session, error) {
	if sess, ok := s.sessions[sessionKey]; ok {
		return sess, nil
	}
	sess := &model.Session{ID: int64(len(s.sessions) + 1), SessionKey: sessionKey}
	s.sessions[sessionKey] = sess
	return sess, nil
}

func (s *stubSessionRepo) FindByKey(_ context.Context, sessionKey string) (*model.Session, error) {
	if sess, ok := s.sessions[sessionKey]; ok {
		return sess, nil
	}
	return nil, nil
}

type stubAnalysisRepo struct {
	records map[int64]*model.AnalysisRecord
	nextID  int64
}

func (s *stubAnalysisRepo) CreateWithFreeContent(_ context.Context, p repository.CreateAnalysisWithFreeContentParams) (int64, error) {
	s.nextID++
	record := &model.AnalysisRecord{
		ID:               s.nextID,
		SessionID:        p.SessionID,
		ModuleType:       p.ModuleType,
		AlgorithmVersion: p.AlgorithmVersion,
		InputPayload:     p.InputPayload,
		ResultPayload:    p.ResultPayload,
		FreeContent:      &p.FreeContent,
		UnlockStatus:     model.AnalysisUnlockStatusLocked,
		GenerationStatus: model.AnalysisGenerationStatusFreeDone,
		Status:           model.AnalysisStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if s.records == nil {
		s.records = map[int64]*model.AnalysisRecord{}
	}
	s.records[record.ID] = record
	return record.ID, nil
}

func (s *stubAnalysisRepo) FindOwnedByID(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
	record, ok := s.records[id]
	if !ok || record.SessionID != sessionID || record.Status != model.AnalysisStatusActive {
		return nil, nil
	}
	return record, nil
}

func (s *stubAnalysisRepo) ListBySession(_ context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error) {
	items := make([]model.AnalysisListItem, 0)
	for _, record := range s.records {
		if record.SessionID != sessionID || record.Status != model.AnalysisStatusActive {
			continue
		}
		if moduleType != nil && record.ModuleType != *moduleType {
			continue
		}
		items = append(items, model.AnalysisListItem{
			ID:               record.ID,
			ModuleType:       record.ModuleType,
			AlgorithmVersion: record.AlgorithmVersion,
			UnlockStatus:     record.UnlockStatus,
			GenerationStatus: record.GenerationStatus,
			CreatedAt:        record.CreatedAt,
		})
	}
	return items, int64(len(items)), 1, 20, nil
}

func (s *stubAnalysisRepo) DeleteOwnedByID(_ context.Context, id, sessionID int64) error {
	record, ok := s.records[id]
	if !ok || record.SessionID != sessionID || record.Status != model.AnalysisStatusActive {
		return repository.ErrAnalysisNotFound
	}
	delete(s.records, id)
	return nil
}

func (s *stubAnalysisRepo) UnlockWithFullContent(_ context.Context, id, sessionID int64, unlockType, fullContent string) error {
	record, ok := s.records[id]
	if !ok || record.SessionID != sessionID || record.Status != model.AnalysisStatusActive || record.UnlockStatus != model.AnalysisUnlockStatusLocked {
		return repository.ErrAnalysisNotFound
	}
	record.UnlockStatus = model.AnalysisUnlockStatusUnlocked
	record.UnlockType = &unlockType
	record.FullContent = &fullContent
	record.GenerationStatus = model.AnalysisGenerationStatusFullDone
	return nil
}

func newTestAnalysisHandler(t *testing.T) (*handler.AnalysisHandler, *stubSessionRepo) {
	t.Helper()
	sessions := &stubSessionRepo{sessions: map[string]*model.Session{}}
	analysisRepo := &stubAnalysisRepo{records: map[int64]*model.AnalysisRecord{}}
	sessionSvc := session.NewServiceWithRepo(sessions)
	baziSvc := bazi.NewServiceWithRepos(sessions, analysisRepo)
	analysisSvc := analysis.NewServiceWithRepo(analysisRepo)
	return handler.NewAnalysisHandler(baziSvc, analysisSvc, sessionSvc), sessions
}

const validBaziCreateBody = `{"session_key":"sess-a","birth_date":"1995-01-01","birth_hour_branch":"zi","birth_hour_unknown":false,"confirm_disclaimer":true}`

func TestCreateBaziSuccess(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)

	body := bytes.NewBufferString(validBaziCreateBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()

	h.CreateBazi(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateBaziRejectsSessionKeyConflict(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-b","birth_date":"1995-01-01","birth_hour_branch":"zi","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for session conflict, got %d", rec.Code)
	}
}

func TestCreateBaziRejectsUnknownFields(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"1995-01-01","birth_hour_branch":"zi","confirm_disclaimer":true,"gender":"male"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown field, got %d", rec.Code)
	}
}

func TestCreateBaziRejectsInvalidBirthDate(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"1995-02-30","birth_hour_branch":"zi","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid birth date, got %d", rec.Code)
	}
}

func TestCreateBaziRejectsMissingConfirmDisclaimer(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"1995-01-01","birth_hour_branch":"zi","birth_hour_unknown":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing confirm_disclaimer, got %d", rec.Code)
	}
}

func TestCreateBaziRejectsFalseConfirmDisclaimer(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"1995-01-01","birth_hour_branch":"zi","birth_hour_unknown":false,"confirm_disclaimer":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for false confirm_disclaimer, got %d", rec.Code)
	}
}

func TestCreateBaziAcceptsConfirmDisclaimerTrue(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(validBaziCreateBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetAnalysisRequiresHeaderSessionKey(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", bytes.NewBufferString(validBaziCreateBody))
	createRec := httptest.NewRecorder()
	h.CreateBazi(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create failed: %s", createRec.Body.String())
	}

	var createResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/"+strconv.FormatInt(createResp.Data.ID, 10)+"?session_key=sess-a", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for query session_key, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/analysis/"+strconv.FormatInt(createResp.Data.ID, 10), nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec = httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetAnalysisOtherSessionForbiddenAsNotFound(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}
	sessions.sessions["sess-b"] = &model.Session{ID: 11, SessionKey: "sess-b"}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", bytes.NewBufferString(validBaziCreateBody))
	createRec := httptest.NewRecorder()
	h.CreateBazi(createRec, createReq)

	var createResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/"+strconv.FormatInt(createResp.Data.ID, 10), nil)
	req.Header.Set(sessionkey.HeaderName, "sess-b")
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for other session, got %d", rec.Code)
	}
}

func TestListAnalysisReturnsSummaryOnly(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"1995-01-01","birth_hour_unknown":true,"confirm_disclaimer":true}`))
	createRec := httptest.NewRecorder()
	h.CreateBazi(createRec, createReq)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=bazi", nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listResp struct {
		Data struct {
			Items []map[string]any `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Data.Items) == 0 {
		t.Fatalf("expected list items")
	}
	for _, item := range listResp.Data.Items {
		for _, forbidden := range []string{"input_payload", "result_payload", "free_content", "full_content"} {
			if _, ok := item[forbidden]; ok {
				t.Fatalf("list item must not contain %s", forbidden)
			}
		}
	}
}

func TestListAnalysisRejectsQuerySessionKey(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=bazi&session_key=sess-a", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListUnknownSessionUsesPaginationValidation(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=bazi", nil)
	req.Header.Set(sessionkey.HeaderName, "unknown-session")
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.Page != 1 || resp.Data.PageSize != 20 || resp.Data.Total != 0 {
		t.Fatalf("unexpected default pagination: %#v", resp.Data)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=bazi&page_size=999999", nil)
	req.Header.Set(sessionkey.HeaderName, "unknown-session")
	rec = httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with capped page_size, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.PageSize != 100 {
		t.Fatalf("expected page_size capped to 100, got %d", resp.Data.PageSize)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=bazi&page=10001", nil)
	req.Header.Set(sessionkey.HeaderName, "unknown-session")
	rec = httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized page, got %d", rec.Code)
	}
}

func TestCreateBaziRejectsTooLongSessionKey(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"` + strings.Repeat("a", 65) + `","birth_date":"1995-01-01","birth_hour_branch":"zi","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for long session key, got %d", rec.Code)
	}
}

func createBaziRecord(t *testing.T, h *handler.AnalysisHandler, sessions *stubSessionRepo, sessionKey string) int64 {
	t.Helper()
	sessions.sessions[sessionKey] = &model.Session{ID: 10, SessionKey: sessionKey}
	body := bytes.NewBufferString(`{"session_key":"` + sessionKey + `","birth_date":"1995-01-01","birth_hour_branch":"zi","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create failed: %s", rec.Body.String())
	}
	var createResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	return createResp.Data.ID
}

func TestDeleteAnalysisSuccess(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/analysis/"+strconv.FormatInt(id, 10), nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "birth_date") || strings.Contains(rec.Body.String(), "input_payload") {
		t.Fatalf("delete response must not contain birth info: %s", rec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/"+strconv.FormatInt(id, 10), nil)
	getReq.Header.Set(sessionkey.HeaderName, "sess-a")
	getRec := httptest.NewRecorder()
	h.Get(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=bazi", nil)
	listReq.Header.Set(sessionkey.HeaderName, "sess-a")
	listRec := httptest.NewRecorder()
	h.List(listRec, listReq)
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

func TestDeleteAnalysisRejectsQuerySessionKey(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"?session_key=sess-a", nil)
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteAnalysisRequiresHeaderSessionKey(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/analysis/"+strconv.FormatInt(id, 10), nil)
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteAnalysisRejectsTooLongSessionKey(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/analysis/"+strconv.FormatInt(id, 10), nil)
	req.Header.Set(sessionkey.HeaderName, strings.Repeat("a", 65))
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteAnalysisOtherSessionNotFound(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-b"] = &model.Session{ID: 11, SessionKey: "sess-b"}
	id := createBaziRecord(t, h, sessions, "sess-a")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/analysis/"+strconv.FormatInt(id, 10), nil)
	req.Header.Set(sessionkey.HeaderName, "sess-b")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for other session delete, got %d", rec.Code)
	}
}

func TestDeleteAnalysisInvalidID(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/analysis/not-a-id", nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", rec.Code)
	}
}

func TestUnlockAnalysisSuccess(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	body := bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var unlockResp struct {
		Data struct {
			UnlockStatus int    `json:"unlock_status"`
			UnlockType   string `json:"unlock_type"`
			FullContent  string `json:"full_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &unlockResp); err != nil {
		t.Fatalf("decode unlock response: %v", err)
	}
	if unlockResp.Data.UnlockStatus != model.AnalysisUnlockStatusUnlocked {
		t.Fatalf("expected unlocked status")
	}
	if unlockResp.Data.UnlockType != model.UnlockTypeRewardedVideoMock {
		t.Fatalf("unexpected unlock type: %s", unlockResp.Data.UnlockType)
	}
	if strings.TrimSpace(unlockResp.Data.FullContent) == "" {
		t.Fatalf("expected full content")
	}
	if strings.Contains(rec.Body.String(), "birth_date") || strings.Contains(rec.Body.String(), "input_payload") {
		t.Fatalf("unlock response must not contain birth info: %s", rec.Body.String())
	}
}

func TestUnlockAnalysisRejectsInvalidUnlockType(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	body := bytes.NewBufferString(`{"unlock_type":"mock_button"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUnlockAnalysisRejectsQuerySessionKey(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	body := bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock?session_key=sess-a", body)
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUnlockAnalysisRequiresHeaderSessionKey(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	body := bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", body)
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUnlockAnalysisOtherSessionNotFound(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-b"] = &model.Session{ID: 11, SessionKey: "sess-b"}
	id := createBaziRecord(t, h, sessions, "sess-a")

	body := bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", body)
	req.Header.Set(sessionkey.HeaderName, "sess-b")
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUnlockAnalysisAlreadyUnlockedReturnsExistingContent(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	unlockOnce := func() string {
		body := bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", body)
		req.Header.Set(sessionkey.HeaderName, "sess-a")
		rec := httptest.NewRecorder()
		h.Unlock(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Data struct {
				FullContent string `json:"full_content"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode unlock response: %v", err)
		}
		return resp.Data.FullContent
	}

	first := unlockOnce()
	second := unlockOnce()
	if first == "" || second == "" {
		t.Fatalf("expected full content on repeated unlock")
	}
	if first != second {
		t.Fatalf("expected repeated unlock to return same full content")
	}
}

func unlockAnalysisRequest(t *testing.T, h *handler.AnalysisHandler, id int64, sessionKey, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", bytes.NewBufferString(body))
	if sessionKey != "" {
		req.Header.Set(sessionkey.HeaderName, sessionKey)
	}
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	return rec
}

func TestUnlockAnalysisRejectsTooLongSessionKey(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	rec := unlockAnalysisRequest(t, h, id, strings.Repeat("a", 65), `{"unlock_type":"rewarded_video_mock"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for long session key, got %d", rec.Code)
	}
}

func TestUnlockAnalysisRejectsEmptyUnlockType(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	rec := unlockAnalysisRequest(t, h, id, "sess-a", `{"unlock_type":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty unlock_type, got %d", rec.Code)
	}
}

func TestUnlockAnalysisRejectsDisallowedUnlockTypes(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	rejected := []string{
		"mock_ad",
		"rewarded_video",
		"paid",
		"admin",
		"unknown",
	}
	for _, unlockType := range rejected {
		t.Run(unlockType, func(t *testing.T) {
			body := `{"unlock_type":"` + unlockType + `"}`
			rec := unlockAnalysisRequest(t, h, id, "sess-a", body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for unlock_type=%q, got %d", unlockType, rec.Code)
			}
		})
	}
}

func TestUnlockAnalysisRejectsUnknownJSONFields(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	rec := unlockAnalysisRequest(t, h, id, "sess-a", `{"unlock_type":"rewarded_video_mock","session_key":"sess-a"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown json field, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUnlockAnalysisNotFound(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}

	rec := unlockAnalysisRequest(t, h, 99999, "sess-a", `{"unlock_type":"rewarded_video_mock"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing id, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUnlockAnalysisUnsupportedModuleForbidden(t *testing.T) {
	sessions := &stubSessionRepo{sessions: map[string]*model.Session{
		"sess-a": {ID: 10, SessionKey: "sess-a"},
	}}
	analysisRepo := &stubAnalysisRepo{records: map[int64]*model.AnalysisRecord{}, nextID: 99}
	free := "free"
	record := &model.AnalysisRecord{
		ID:               99,
		SessionID:        10,
		ModuleType:       model.ModuleTypeQimen,
		AlgorithmVersion: model.AlgorithmVersionQimenSimpleV1,
		InputPayload:     json.RawMessage(`{"question":"test"}`),
		ResultPayload:    json.RawMessage(`{"day_master":"甲","pillars":{"year":"甲子","month":"乙丑","day":"丙寅"},"five_elements":{"wood":1,"fire":1,"earth":1,"metal":1,"water":1}}`),
		FreeContent:      &free,
		UnlockStatus:     model.AnalysisUnlockStatusLocked,
		GenerationStatus: model.AnalysisGenerationStatusFreeDone,
		Status:           model.AnalysisStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	analysisRepo.records[record.ID] = record

	sessionSvc := session.NewServiceWithRepo(sessions)
	baziSvc := bazi.NewServiceWithRepos(sessions, analysisRepo)
	analysisSvc := analysis.NewServiceWithRepo(analysisRepo)
	customHandler := handler.NewAnalysisHandler(baziSvc, analysisSvc, sessionSvc)

	body := bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/99/unlock", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	customHandler.Unlock(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for qimen unlock, got %d body=%s", rec.Code, rec.Body.String())
	}
}
