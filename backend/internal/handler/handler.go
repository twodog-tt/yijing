package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/pkg/response"
	"github.com/wangxintong/yijing/backend/internal/pkg/sessionkey"
	"github.com/wangxintong/yijing/backend/internal/service/category"
	"github.com/wangxintong/yijing/backend/internal/service/divination"
	"github.com/wangxintong/yijing/backend/internal/service/session"
)

type HealthHandler struct {
	DB *sql.DB
}

type healthPayload struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
	DB        string `json:"db,omitempty"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	if h.DB != nil {
		if err := h.DB.PingContext(r.Context()); err != nil {
			dbStatus = "down"
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthPayload{
		Status:    "ok",
		Service:   "yijing-backend",
		Timestamp: clock.FormatRFC3339(clock.Now()),
		DB:        dbStatus,
	})
}

type SessionHandler struct {
	svc *session.Service
}

func NewSessionHandler(svc *session.Service) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, http.ErrBodyNotAllowed) {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid json body")
		return
	}

	s, err := h.svc.CreateOrGet(r.Context(), req.SessionKey, r.UserAgent())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "create session failed")
		return
	}

	response.OK(w, map[string]any{
		"session_id":  s.ID,
		"session_key": s.SessionKey,
	})
}

type CategoryHandler struct {
	svc *category.Service
}

func NewCategoryHandler(svc *category.Service) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListActive(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "list categories failed")
		return
	}
	response.OK(w, items)
}

type DivinationHandler struct {
	svc *divination.Service
}

func NewDivinationHandler(svc *divination.Service) *DivinationHandler {
	return &DivinationHandler{svc: svc}
}

func (h *DivinationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDivinationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid json body")
		return
	}

	result, err := h.svc.Create(r.Context(), divination.CreateInput{
		SessionKey:        req.SessionKey,
		CategoryID:        req.CategoryID,
		Question:          req.Question,
		ConfirmDisclaimer: req.ConfirmDisclaimer,
		ClientInfo:        r.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, divination.ErrSensitiveBlocked):
			response.Error(w, http.StatusBadRequest, response.CodeSensitiveBlock,
				"这个问题不适合用卦象方式解读。你可以换成更偏向自我反思、情绪整理或行动选择的问题。")
		case errors.Is(err, divination.ErrCategoryNotFound):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "category not found or disabled")
		case errors.Is(err, divination.ErrSessionKeyEmpty):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
		case errors.Is(err, divination.ErrInvalidParams):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid params: check question length (5-200) and confirm_disclaimer")
		default:
			response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "create divination failed")
		}
		return
	}

	response.OK(w, toDivinationResponse(result))
}

func (h *DivinationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseDivinationIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid divination id")
		return
	}

	result, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "get divination failed")
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, response.CodeNotFound, "divination not found")
		return
	}

	response.OK(w, toDivinationResponse(result))
}

func (h *DivinationHandler) List(w http.ResponseWriter, r *http.Request) {
	sessionKey := strings.TrimSpace(r.URL.Query().Get("session_key"))
	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)

	if sessionKey == "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
		return
	}

	result, err := h.svc.ListHistory(r.Context(), sessionKey, page, pageSize)
	if err != nil {
		if errors.Is(err, divination.ErrSessionKeyEmpty) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "list divinations failed")
		return
	}

	response.OK(w, result)
}

func (h *DivinationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if sessionkey.FromQuery(r) != "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key must be sent via X-Session-Key header")
		return
	}

	sessionKey := sessionkey.FromHeader(r)
	if strings.TrimSpace(sessionKey) == "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
		return
	}
	if err := sessionkey.ValidateLength(sessionKey); err != nil {
		if errors.Is(err, sessionkey.ErrTooLong) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key exceeds max length")
			return
		}
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid session_key")
		return
	}

	id, err := parseDivinationIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid divination id")
		return
	}

	if err := h.svc.Delete(r.Context(), sessionKey, id); err != nil {
		switch {
		case errors.Is(err, divination.ErrSessionKeyEmpty):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
		case errors.Is(err, divination.ErrInvalidParams):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid params")
		case errors.Is(err, divination.ErrNotFound):
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "divination not found")
		default:
			response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "delete divination failed")
		}
		return
	}

	response.OK(w, nil)
}

func toDivinationResponse(d *model.Divination) map[string]any {
	resp := map[string]any{
		"id":               d.ID,
		"question":         d.Question,
		"category":         d.Category,
		"primary_hexagram": d.PrimaryHexagram,
		"changed_hexagram": d.ChangedHexagram,
		"lines":            d.Lines,
		"moving_lines":     d.MovingLinesArray,
		"unlock_status":    d.UnlockStatus,
		"created_at":       clock.FormatRFC3339(d.CreatedAt),
	}
	if d.FreeInterpretation != "" {
		resp["free_interpretation"] = d.FreeInterpretation
	}
	return resp
}

func strconvParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseIDFromPath(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.TrimSuffix(idStr, "/")
	return strconv.ParseInt(idStr, 10, 64)
}

func parseIntDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
