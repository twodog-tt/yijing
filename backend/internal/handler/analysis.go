package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/pkg/response"
	"github.com/wangxintong/yijing/backend/internal/pkg/sessionkey"
	"github.com/wangxintong/yijing/backend/internal/service/analysis"
	"github.com/wangxintong/yijing/backend/internal/service/bazi"
	"github.com/wangxintong/yijing/backend/internal/service/session"
)

type AnalysisHandler struct {
	baziSvc     *bazi.Service
	analysisSvc *analysis.Service
	sessionSvc  sessionResolver
}

type sessionResolver interface {
	SessionIDByKey(ctx context.Context, sessionKey string) (int64, error)
}

func NewAnalysisHandler(baziSvc *bazi.Service, analysisSvc *analysis.Service, sessionSvc sessionResolver) *AnalysisHandler {
	return &AnalysisHandler{
		baziSvc:     baziSvc,
		analysisSvc: analysisSvc,
		sessionSvc:  sessionSvc,
	}
}

type createBaziRequest struct {
	SessionKey        string `json:"session_key"`
	BirthDate         string `json:"birth_date"`
	BirthHourBranch   string `json:"birth_hour_branch"`
	BirthHourUnknown  bool   `json:"birth_hour_unknown"`
	ConfirmDisclaimer bool   `json:"confirm_disclaimer"`
}

func (h *AnalysisHandler) CreateBazi(w http.ResponseWriter, r *http.Request) {
	var req createBaziRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid json body")
		return
	}

	sessionKey, err := sessionkey.ResolveForCreate(sessionkey.FromHeader(r), req.SessionKey)
	if err != nil {
		switch {
		case errors.Is(err, sessionkey.ErrConflict):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key conflict between header and body")
		case errors.Is(err, sessionkey.ErrTooLong):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key exceeds max length")
		default:
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid session_key")
		}
		return
	}
	if !req.ConfirmDisclaimer {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "confirm_disclaimer must be true")
		return
	}

	record, err := h.baziSvc.Create(r.Context(), bazi.CreateInput{
		SessionKey:       sessionKey,
		BirthDate:        req.BirthDate,
		BirthHourBranch:  req.BirthHourBranch,
		BirthHourUnknown: req.BirthHourUnknown,
		ClientInfo:       r.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, bazi.ErrSessionKeyEmpty):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
		case errors.Is(err, bazi.ErrSessionKeyTooLong):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key exceeds max length")
		case errors.Is(err, bazi.ErrInvalidParams):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid params: check birth_date and birth_hour_branch")
		default:
			response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "create bazi analysis failed")
		}
		return
	}

	response.OK(w, toAnalysisResponse(record))
}

func (h *AnalysisHandler) Get(w http.ResponseWriter, r *http.Request) {
	if sessionkey.FromQuery(r) != "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key must be sent via X-Session-Key header")
		return
	}

	sessionKey := sessionkey.FromHeader(r)
	sessionID, err := h.sessionSvc.SessionIDByKey(r.Context(), sessionKey)
	if err != nil {
		if errors.Is(err, session.ErrSessionKeyEmpty) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "resolve session failed")
		return
	}
	if sessionID == 0 {
		response.Error(w, http.StatusNotFound, response.CodeNotFound, "analysis not found")
		return
	}

	id, err := parseAnalysisIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid analysis id")
		return
	}

	record, err := h.analysisSvc.Get(r.Context(), sessionID, id)
	if err != nil {
		if errors.Is(err, analysis.ErrInvalidParams) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid params")
			return
		}
		if errors.Is(err, analysis.ErrNotFound) {
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "analysis not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "get analysis failed")
		return
	}

	response.OK(w, toAnalysisResponse(record))
}

func (h *AnalysisHandler) List(w http.ResponseWriter, r *http.Request) {
	if sessionkey.FromQuery(r) != "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key must be sent via X-Session-Key header")
		return
	}

	sessionKey := sessionkey.FromHeader(r)
	sessionID, err := h.sessionSvc.SessionIDByKey(r.Context(), sessionKey)
	if err != nil {
		if errors.Is(err, session.ErrSessionKeyEmpty) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "resolve session failed")
		return
	}

	moduleType, err := parseAnalysisModuleFilter(r.URL.Query().Get("module"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid module filter")
		return
	}

	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)

	result, err := h.analysisSvc.List(r.Context(), sessionID, moduleType, page, pageSize)
	if err != nil {
		if errors.Is(err, analysis.ErrInvalidParams) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid params")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "list analysis failed")
		return
	}

	response.OK(w, result)
}

func toAnalysisResponse(record *model.AnalysisRecord) map[string]any {
	resp := map[string]any{
		"id":                record.ID,
		"module_type":       record.ModuleType,
		"algorithm_version": record.AlgorithmVersion,
		"input_payload":     json.RawMessage(record.InputPayload),
		"result_payload":    json.RawMessage(record.ResultPayload),
		"unlock_status":     record.UnlockStatus,
		"generation_status": record.GenerationStatus,
		"created_at":        clock.FormatRFC3339(record.CreatedAt),
		"updated_at":        clock.FormatRFC3339(record.UpdatedAt),
	}
	if record.FreeContent != nil {
		resp["free_content"] = *record.FreeContent
	}
	return resp
}

func parseAnalysisIDFromPath(path string) (int64, error) {
	const prefix = "/api/v1/analysis/"
	if !strings.HasPrefix(path, prefix) {
		return 0, errors.New("invalid path")
	}
	rest := strings.TrimPrefix(path, prefix)
	idPart := strings.SplitN(rest, "/", 2)[0]
	return strconvParseInt(idPart)
}

func parseAnalysisModuleFilter(module string) (*int, error) {
	module = strings.TrimSpace(strings.ToLower(module))
	if module == "" {
		return nil, nil
	}
	if module != "bazi" {
		return nil, errors.New("unsupported module")
	}
	moduleType := model.ModuleTypeBazi
	return &moduleType, nil
}
