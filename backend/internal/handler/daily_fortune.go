package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/wangxintong/yijing/backend/internal/pkg/response"
	"github.com/wangxintong/yijing/backend/internal/service/dailyfortune"
)

type DailyFortuneHandler struct {
	svc *dailyfortune.Service
}

func NewDailyFortuneHandler(svc *dailyfortune.Service) *DailyFortuneHandler {
	return &DailyFortuneHandler{svc: svc}
}

type dailyFortuneTodayRequest struct {
	SessionKey string `json:"session_key"`
	LocalDate  string `json:"local_date"`
}

func (h *DailyFortuneHandler) Today(w http.ResponseWriter, r *http.Request) {
	var req dailyFortuneTodayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid json body")
		return
	}

	result, err := h.svc.GetOrCreateToday(r.Context(), dailyfortune.TodayInput{
		SessionKey: req.SessionKey,
		LocalDate:  req.LocalDate,
		ClientInfo: r.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, dailyfortune.ErrSessionKeyEmpty):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "session_key is required")
		case errors.Is(err, dailyfortune.ErrInvalidDate):
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "local_date must be YYYY-MM-DD")
		default:
			response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "daily fortune failed")
		}
		return
	}

	response.OK(w, map[string]any{
		"daily_fortune": map[string]any{
			"fortune_date": result.FortuneDate,
			"is_existing":  result.IsExisting,
		},
		"divination": toDivinationResponse(result.Divination),
	})
}
