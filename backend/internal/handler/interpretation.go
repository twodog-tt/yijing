package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/response"
	"github.com/wangxintong/yijing/backend/internal/service/interpretation"
	"github.com/wangxintong/yijing/backend/internal/service/unlock"
)

type InterpretationHandler struct {
	interpretationSvc *interpretation.Service
	unlockSvc         *unlock.Service
}

func NewInterpretationHandler(interpretationSvc *interpretation.Service, unlockSvc *unlock.Service) *InterpretationHandler {
	return &InterpretationHandler{
		interpretationSvc: interpretationSvc,
		unlockSvc:         unlockSvc,
	}
}

func (h *InterpretationHandler) GetFree(w http.ResponseWriter, r *http.Request) {
	id, err := parseDivinationIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid divination id")
		return
	}

	record, err := h.interpretationSvc.GetFree(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "get free interpretation failed")
		return
	}
	if record == nil {
		response.Error(w, http.StatusNotFound, response.CodeNotFound, "interpretation not found")
		return
	}

	response.OK(w, map[string]any{
		"divination_id":      record.DivinationID,
		"free_content":       record.FreeContent,
		"ai_provider":        record.AIProvider,
		"generation_status":  record.GenerationStatus,
		"generated_at":       record.GeneratedAt,
	})
}

func (h *InterpretationHandler) GetFull(w http.ResponseWriter, r *http.Request) {
	id, err := parseDivinationIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid divination id")
		return
	}

	sessionKey := strings.TrimSpace(r.URL.Query().Get("session_key"))
	result, err := h.unlockSvc.GetFullInterpretationWithMeta(r.Context(), id, sessionKey)
	if err != nil {
		switch {
		case errors.Is(err, unlock.ErrInvalidParams):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
		case errors.Is(err, unlock.ErrNotFound):
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "divination not found")
		case errors.Is(err, unlock.ErrForbidden):
			response.Error(w, http.StatusForbidden, response.CodeForbidden, "forbidden")
		case errors.Is(err, unlock.ErrNotUnlocked):
			response.Error(w, http.StatusForbidden, response.CodeForbidden, "full interpretation not unlocked")
		default:
			response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "get full interpretation failed")
		}
		return
	}

	response.OK(w, map[string]any{
		"divination_id": id,
		"full_content":  result.Report,
		"ai_provider":   result.Provider,
	})
}

type UnlockHandler struct {
	svc *unlock.Service
}

func NewUnlockHandler(svc *unlock.Service) *UnlockHandler {
	return &UnlockHandler{svc: svc}
}

func (h *UnlockHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	id, err := parseDivinationIDFromPath(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid divination id")
		return
	}

	var req model.UnlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid json body")
		return
	}

	result, err := h.svc.Unlock(r.Context(), id, req.SessionKey, req.UnlockType)
	if err != nil {
		switch {
		case errors.Is(err, unlock.ErrInvalidParams):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid params: session_key and unlock_type required")
		case errors.Is(err, unlock.ErrNotFound):
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "divination not found")
		case errors.Is(err, unlock.ErrForbidden):
			response.Error(w, http.StatusForbidden, response.CodeForbidden, "forbidden")
		default:
			response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "unlock failed")
		}
		return
	}

	response.OK(w, map[string]any{
		"divination_id":       result.DivinationID,
		"unlock_status":       result.UnlockStatus,
		"mock_transaction_id": result.MockTransactionID,
		"full_interpretation": result.FullInterpretation,
	})
}

func parseDivinationIDFromPath(path string) (int64, error) {
	const prefix = "/api/v1/divinations/"
	if !strings.HasPrefix(path, prefix) {
		return 0, errors.New("invalid path")
	}
	rest := strings.TrimPrefix(path, prefix)
	idPart := strings.SplitN(rest, "/", 2)[0]
	return strconvParseInt(idPart)
}
