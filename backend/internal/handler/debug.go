package handler

import (
	"net/http"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/pkg/response"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/ai"
)

// DebugHandler 仅用于本地 MVP 调试。
// 生产环境必须为 /api/v1/debug/* 增加权限保护（如管理员鉴权、内网访问限制）。
type DebugHandler struct {
	aiLogRepo *repository.AILogRepository
	aiRouter  *ai.Router
}

func NewDebugHandler(aiLogRepo *repository.AILogRepository, aiRouter *ai.Router) *DebugHandler {
	return &DebugHandler{aiLogRepo: aiLogRepo, aiRouter: aiRouter}
}

func (h *DebugHandler) ListAILogs(w http.ResponseWriter, r *http.Request) {
	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)

	result, err := h.aiLogRepo.ListRecent(r.Context(), page, pageSize)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "list ai logs failed")
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		row := map[string]any{
			"id":            item.ID,
			"divination_id": item.DivinationID,
			"ai_provider":   item.AIProvider,
			"model_name":    item.ModelName,
			"status":        item.Status,
			"duration_ms":   item.DurationMs,
			"fallback_used": item.FallbackUsed,
			"created_at":    clock.FormatRFC3339(item.CreatedAt),
		}
		if item.ErrorMessage != nil {
			row["error_message"] = *item.ErrorMessage
		}
		items = append(items, row)
	}

	response.OK(w, map[string]any{
		"items":     items,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	})
}

func (h *DebugHandler) AIHealth(w http.ResponseWriter, r *http.Request) {
	cfg := h.aiRouter.Config()
	provider := h.aiRouter.ConfiguredProvider()

	data := map[string]any{
		"provider": provider,
	}

	switch provider {
	case "deepseek":
		data["api_key_configured"] = strings.TrimSpace(cfg.DeepSeekAPIKey) != ""
		data["model"] = cfg.DeepSeekModel
		data["base_url"] = cfg.DeepSeekBaseURL
		data["timeout_seconds"] = cfg.DeepSeekTimeoutSeconds
	default:
		data["api_key_configured"] = false
		data["model"] = "mock"
		data["base_url"] = ""
		data["timeout_seconds"] = 0
	}

	response.OK(w, data)
}

func (h *DebugHandler) AIStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.aiLogRepo.GetStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "get ai stats failed")
		return
	}

	data := map[string]any{
		"total_count":     stats.TotalCount,
		"success_count":   stats.SuccessCount,
		"fail_count":      stats.FailCount,
		"fallback_count":  stats.FallbackCount,
		"avg_duration_ms": stats.AvgDurationMs,
	}
	if stats.LatestCreatedAt != nil {
		data["latest_created_at"] = clock.FormatRFC3339(*stats.LatestCreatedAt)
	}
	response.OK(w, data)
}
