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
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
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
			Question:         record.Question,
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

func (s *stubAnalysisRepo) UnlockWithFullContent(_ context.Context, id, sessionID int64, unlockType, fullContent, aiProvider string) error {
	record, ok := s.records[id]
	if !ok || record.SessionID != sessionID || record.Status != model.AnalysisStatusActive || record.UnlockStatus != model.AnalysisUnlockStatusLocked {
		return repository.ErrAnalysisNotFound
	}
	record.UnlockStatus = model.AnalysisUnlockStatusUnlocked
	record.UnlockType = &unlockType
	record.FullContent = &fullContent
	record.AIProvider = &aiProvider
	record.GenerationStatus = model.AnalysisGenerationStatusFullDone
	return nil
}

func (s *stubAnalysisRepo) UpdateUnlockedFullContent(_ context.Context, id, sessionID int64, fullContent, aiProvider string) error {
	record, ok := s.records[id]
	if !ok || record.SessionID != sessionID || record.Status != model.AnalysisStatusActive || record.UnlockStatus != model.AnalysisUnlockStatusUnlocked {
		return repository.ErrAnalysisNotFound
	}
	if record.FullContent != nil && strings.TrimSpace(*record.FullContent) != "" {
		return nil
	}
	record.FullContent = &fullContent
	record.AIProvider = &aiProvider
	record.GenerationStatus = model.AnalysisGenerationStatusFullDone
	return nil
}

func newTestAnalysisHandler(t *testing.T) (*handler.AnalysisHandler, *stubSessionRepo) {
	t.Helper()
	sessions := &stubSessionRepo{sessions: map[string]*model.Session{}}
	analysisRepo := &stubAnalysisRepo{records: map[int64]*model.AnalysisRecord{}}
	sessionSvc := session.NewServiceWithRepo(sessions)
	baziSvc := bazi.NewServiceWithRepos(sessions, analysisRepo)
	qimenSvc := qimen.NewServiceWithRepos(sessions, analysisRepo, nil)
	analysisSvc := analysis.NewServiceWithRepo(analysisRepo)
	return handler.NewAnalysisHandler(baziSvc, qimenSvc, analysisSvc, sessionSvc), sessions
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

func TestCreateBaziDefaultsToSimpleV1(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", bytes.NewBufferString(validBaziCreateBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			AlgorithmVersion string `json:"algorithm_version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.AlgorithmVersion != model.AlgorithmVersionBaziSimpleV1 {
		t.Fatalf("algorithm_version=%q", resp.Data.AlgorithmVersion)
	}
}

func TestCreateBaziUsesV2AlgorithmVersion(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"2024-02-05","birth_hour_branch":"wu","birth_hour_unknown":false,"confirm_disclaimer":true,"algorithm_version":"bazi-v2-poc"}`)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			AlgorithmVersion string          `json:"algorithm_version"`
			ResultPayload    json.RawMessage `json:"result_payload"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.AlgorithmVersion != bazi.AlgorithmVersionBaziV2POC {
		t.Fatalf("algorithm_version=%q", resp.Data.AlgorithmVersion)
	}
	raw := string(resp.Data.ResultPayload)
	if !strings.Contains(raw, `"algorithm_version":"bazi-v2-poc"`) {
		t.Fatalf("result_payload missing v2 marker: %s", raw)
	}
	for _, forbidden := range []string{"birth_date", "session_key", "prompt"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("result_payload must not contain %q", forbidden)
		}
	}
}

func TestCreateBaziRejectsInvalidAlgorithmVersion(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"1995-01-01","birth_hour_branch":"zi","confirm_disclaimer":true,"algorithm_version":"bazi-v3"}`)
	rec := httptest.NewRecorder()
	h.CreateBazi(rec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "algorithm_version must be bazi-simple-v1 or bazi-v2-poc") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestUnlockAnalysisV2FreeUnlockSuccess(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}
	body := bytes.NewBufferString(`{"session_key":"sess-a","birth_date":"2024-02-05","birth_hour_branch":"wu","birth_hour_unknown":false,"confirm_disclaimer":true,"algorithm_version":"bazi-v2-poc"}`)
	createRec := httptest.NewRecorder()
	h.CreateBazi(createRec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/bazi", body))
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

	unlockRec := httptest.NewRecorder()
	unlockReq := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(createResp.Data.ID, 10)+"/unlock", bytes.NewBufferString(`{"unlock_type":"free_unlock"}`))
	unlockReq.Header.Set(sessionkey.HeaderName, "sess-a")
	h.Unlock(unlockRec, unlockReq)
	if unlockRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", unlockRec.Code, unlockRec.Body.String())
	}
	var unlockResp struct {
		Data struct {
			UnlockType  string `json:"unlock_type"`
			FullContent string `json:"full_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(unlockRec.Body.Bytes(), &unlockResp); err != nil {
		t.Fatalf("decode unlock response: %v", err)
	}
	if unlockResp.Data.UnlockType != model.UnlockTypeFreeUnlock {
		t.Fatalf("unexpected unlock type: %s", unlockResp.Data.UnlockType)
	}
	if strings.TrimSpace(unlockResp.Data.FullContent) == "" {
		t.Fatalf("expected full content")
	}
	if !strings.Contains(unlockResp.Data.FullContent, "bazi-v2-poc") {
		t.Fatalf("expected v2 full content marker")
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
			AIProvider   string `json:"ai_provider"`
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
	if unlockResp.Data.AIProvider != model.AIProviderTemplateFallback {
		t.Fatalf("expected ai_provider template_fallback, got %q", unlockResp.Data.AIProvider)
	}
	if strings.Contains(rec.Body.String(), "birth_date") || strings.Contains(rec.Body.String(), "input_payload") {
		t.Fatalf("unlock response must not contain birth info: %s", rec.Body.String())
	}
}

func TestUnlockAnalysisFreeUnlockSuccess(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	body := bytes.NewBufferString(`{"unlock_type":"free_unlock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var unlockResp struct {
		Data struct {
			UnlockType  string `json:"unlock_type"`
			FullContent string `json:"full_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &unlockResp); err != nil {
		t.Fatalf("decode unlock response: %v", err)
	}
	if unlockResp.Data.UnlockType != model.UnlockTypeFreeUnlock {
		t.Fatalf("unexpected unlock type: %s", unlockResp.Data.UnlockType)
	}
	if strings.TrimSpace(unlockResp.Data.FullContent) == "" {
		t.Fatalf("expected full content")
	}
}

func TestUnlockAnalysisQimenFreeUnlockSuccess(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createQimenRecord(t, h, sessions, "sess-a")

	body := bytes.NewBufferString(`{"unlock_type":"free_unlock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.Unlock(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var unlockResp struct {
		Data struct {
			UnlockType  string `json:"unlock_type"`
			FullContent string `json:"full_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &unlockResp); err != nil {
		t.Fatalf("decode unlock response: %v", err)
	}
	if unlockResp.Data.UnlockType != model.UnlockTypeFreeUnlock {
		t.Fatalf("unexpected unlock type: %s", unlockResp.Data.UnlockType)
	}
	if strings.TrimSpace(unlockResp.Data.FullContent) == "" {
		t.Fatalf("expected full content")
	}
}

func TestGetAnalysisUnlockedIncludesAIProvider(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createBaziRecord(t, h, sessions, "sess-a")

	unlockReq := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/"+strconv.FormatInt(id, 10)+"/unlock", bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`))
	unlockReq.Header.Set(sessionkey.HeaderName, "sess-a")
	unlockRec := httptest.NewRecorder()
	h.Unlock(unlockRec, unlockReq)
	if unlockRec.Code != http.StatusOK {
		t.Fatalf("unlock expected 200, got %d", unlockRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/"+strconv.FormatInt(id, 10), nil)
	getReq.Header.Set(sessionkey.HeaderName, "sess-a")
	getRec := httptest.NewRecorder()
	h.Get(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d body=%s", getRec.Code, getRec.Body.String())
	}

	var getResp struct {
		Data struct {
			UnlockStatus int    `json:"unlock_status"`
			AIProvider   string `json:"ai_provider"`
			FullContent  string `json:"full_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getResp.Data.UnlockStatus != model.AnalysisUnlockStatusUnlocked {
		t.Fatalf("expected unlocked status")
	}
	if getResp.Data.AIProvider != model.AIProviderTemplateFallback {
		t.Fatalf("expected ai_provider template_fallback, got %q", getResp.Data.AIProvider)
	}
	if strings.TrimSpace(getResp.Data.FullContent) == "" {
		t.Fatalf("expected full_content in detail response")
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
		ModuleType:       99,
		AlgorithmVersion: model.AlgorithmVersionBaziSimpleV1,
		InputPayload:     json.RawMessage(`{"birth_date":"1995-01-01"}`),
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
	qimenSvc := qimen.NewServiceWithRepos(sessions, analysisRepo, nil)
	analysisSvc := analysis.NewServiceWithRepo(analysisRepo)
	customHandler := handler.NewAnalysisHandler(baziSvc, qimenSvc, analysisSvc, sessionSvc)

	body := bytes.NewBufferString(`{"unlock_type":"rewarded_video_mock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/99/unlock", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	customHandler.Unlock(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unsupported module unlock, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func createQimenRecord(t *testing.T, h *handler.AnalysisHandler, sessions *stubSessionRepo, sessionKey string) int64 {
	t.Helper()
	sessions.sessions[sessionKey] = &model.Session{ID: 10, SessionKey: sessionKey}
	body := bytes.NewBufferString(`{"session_key":"` + sessionKey + `","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create qimen failed: %s", rec.Body.String())
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

func TestUnlockQimenAnalysisSuccess(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createQimenRecord(t, h, sessions, "sess-a")

	rec := unlockAnalysisRequest(t, h, id, "sess-a", `{"unlock_type":"rewarded_video_mock"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for qimen unlock, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			UnlockStatus int    `json:"unlock_status"`
			FullContent  string `json:"full_content"`
			AIProvider   string `json:"ai_provider"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode unlock response: %v", err)
	}
	if resp.Data.UnlockStatus != model.AnalysisUnlockStatusUnlocked {
		t.Fatalf("expected unlocked status")
	}
	if strings.TrimSpace(resp.Data.FullContent) == "" {
		t.Fatalf("expected full_content")
	}
	if resp.Data.AIProvider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", resp.Data.AIProvider)
	}
	if strings.Contains(rec.Body.String(), "我最近适合推进") {
		t.Fatalf("unlock response must not contain original question")
	}
}

func TestUnlockQimenAnalysisWrongSessionNotFound(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createQimenRecord(t, h, sessions, "sess-a")
	sessions.sessions["sess-b"] = &model.Session{ID: 11, SessionKey: "sess-b"}

	rec := unlockAnalysisRequest(t, h, id, "sess-b", `{"unlock_type":"rewarded_video_mock"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for other session, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUnlockQimenAnalysisAlreadyUnlockedReturnsExistingContent(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createQimenRecord(t, h, sessions, "sess-a")

	unlockOnce := func() string {
		rec := unlockAnalysisRequest(t, h, id, "sess-a", `{"unlock_type":"rewarded_video_mock"}`)
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
		t.Fatalf("expected full content on repeated qimen unlock")
	}
	if first != second {
		t.Fatalf("expected repeated qimen unlock to return same full content")
	}
}

func TestUnlockQimenAnalysisRejectsMockButton(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	id := createQimenRecord(t, h, sessions, "sess-a")

	rec := unlockAnalysisRequest(t, h, id, "sess-a", `{"unlock_type":"mock_button"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for mock_button on qimen, got %d body=%s", rec.Code, rec.Body.String())
	}
}

const validQimenCreateBody = `{"session_key":"sess-a","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true}`

func TestCreateQimenSuccess(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)

	body := bytes.NewBufferString(validQimenCreateBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object")
	}
	if int(data["module_type"].(float64)) != model.ModuleTypeQimen {
		t.Fatalf("expected qimen module_type")
	}
	if data["algorithm_version"] != model.AlgorithmVersionQimenSimpleV1 {
		t.Fatalf("expected qimen algorithm version")
	}
}

func TestCreateQimenUsesV2AlgorithmVersion(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true,"algorithm_version":"qimen-v2-poc"}`)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			AlgorithmVersion string          `json:"algorithm_version"`
			ResultPayload    json.RawMessage `json:"result_payload"`
			FreeContent      string          `json:"free_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.AlgorithmVersion != qimen.AlgorithmVersionQimenV2POC {
		t.Fatalf("algorithm_version=%q", resp.Data.AlgorithmVersion)
	}
	raw := string(resp.Data.ResultPayload)
	if !strings.Contains(raw, `"algorithm_version":"qimen-v2-poc"`) {
		t.Fatalf("result_payload missing v2 marker: %s", raw)
	}
	if !strings.Contains(raw, `"palaces"`) {
		t.Fatalf("result_payload missing palaces: %s", raw)
	}
	for _, forbidden := range []string{"session_key", "prompt", "input_payload"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("result_payload must not contain %q", forbidden)
		}
	}
	if strings.Contains(rec.Body.String(), "session_key") {
		t.Fatalf("response must not contain session_key")
	}
}

func TestCreateQimenExplicitV1AlgorithmVersion(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true,"algorithm_version":"qimen-simple-v1"}`)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			AlgorithmVersion string `json:"algorithm_version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.AlgorithmVersion != model.AlgorithmVersionQimenSimpleV1 {
		t.Fatalf("algorithm_version=%q", resp.Data.AlgorithmVersion)
	}
}

func TestCreateQimenRejectsInvalidAlgorithmVersion(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true,"algorithm_version":"qimen-v3"}`)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "algorithm_version must be qimen-simple-v1 or qimen-v2-poc") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestUnlockQimenV2FreeUnlockSuccess(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}
	body := bytes.NewBufferString(`{"session_key":"sess-a","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true,"algorithm_version":"qimen-v2-poc"}`)
	createRec := httptest.NewRecorder()
	h.CreateQimen(createRec, httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body))
	if createRec.Code != http.StatusOK {
		t.Fatalf("create v2 qimen failed: %s", createRec.Body.String())
	}
	var createResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	rec := unlockAnalysisRequest(t, h, createResp.Data.ID, "sess-a", `{"unlock_type":"free_unlock"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for v2 qimen unlock, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			FullContent string `json:"full_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode unlock: %v", err)
	}
	if strings.TrimSpace(resp.Data.FullContent) == "" {
		t.Fatalf("expected full_content")
	}
	if !strings.Contains(resp.Data.FullContent, "九宫") && !strings.Contains(resp.Data.FullContent, "宫位") {
		t.Fatalf("expected v2 palace observation in full_content")
	}
	if !strings.Contains(resp.Data.FullContent, "POC") {
		t.Fatalf("expected POC note in full_content")
	}
	if strings.Contains(rec.Body.String(), "session_key") || strings.Contains(rec.Body.String(), "prompt") {
		t.Fatalf("unlock response must not expose session_key or prompt")
	}
}

func TestCreateQimenRejectsMissingDisclaimer(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","question":"我最近适合推进这个计划吗？","category":"career"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateQimenRejectsUnknownFields(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true,"phone":"13800000000"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown field, got %d", rec.Code)
	}
}

func TestCreateQimenRejectsHighRiskQuestion(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-a","question":"这只股票明天会涨吗？","category":"decision","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for sensitive question, got %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "这只股票明天会涨吗") {
		t.Fatalf("error response must not echo full question")
	}
}

func TestCreateQimenRejectsSessionKeyInQuery(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(validQimenCreateBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen?session_key=sess-a", body)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for query session_key, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "session_key must be sent via X-Session-Key header") {
		t.Fatalf("expected query session_key rejection message")
	}
}

func TestCreateQimenRejectsSessionKeyConflict(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"sess-b","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for session conflict, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateQimenRejectsTooLongSessionKey(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	body := bytes.NewBufferString(`{"session_key":"` + strings.Repeat("a", 65) + `","question":"我最近适合推进这个计划吗？","category":"career","confirm_disclaimer":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", body)
	rec := httptest.NewRecorder()
	h.CreateQimen(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for long session key, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListAnalysisQimenReturnsSummaryOnly(t *testing.T) {
	h, sessions := newTestAnalysisHandler(t)
	sessions.sessions["sess-a"] = &model.Session{ID: 10, SessionKey: "sess-a"}

	createBody := bytes.NewBufferString(validQimenCreateBody)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/qimen", createBody)
	createRec := httptest.NewRecorder()
	h.CreateQimen(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create qimen failed: %d %s", createRec.Code, createRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=qimen", nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if strings.Contains(body, "我最近适合推进这个计划吗？") {
		t.Fatalf("list response must not contain full question")
	}
	if !strings.Contains(body, qimen.QuestionSummary) {
		t.Fatalf("expected question summary in list response")
	}
	if strings.Contains(body, "input_payload") || strings.Contains(body, "result_payload") || strings.Contains(body, "free_content") {
		t.Fatalf("list response must not contain payloads or free_content")
	}
}

func TestListAnalysisRejectsInvalidModuleQimenTypo(t *testing.T) {
	h, _ := newTestAnalysisHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis?module=qiemen", nil)
	req.Header.Set(sessionkey.HeaderName, "sess-a")
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid module, got %d", rec.Code)
	}
}
