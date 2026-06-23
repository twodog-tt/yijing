package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
)

type AILogRepository struct {
	db *sql.DB
}

func NewAILogRepository(db *sql.DB) *AILogRepository {
	return &AILogRepository{db: db}
}

type CreateAILogInput struct {
	DivinationID    int64
	QuestionSummary string
	AIProvider      string
	ModelName       string
	Status          int
	DurationMs      int
	FallbackUsed    int
	ErrorMessage    string
}

func (r *AILogRepository) Create(ctx context.Context, in CreateAILogInput) error {
	var questionSummary sql.NullString
	if s := strings.TrimSpace(in.QuestionSummary); s != "" {
		questionSummary = sql.NullString{String: truncateRunes(s, 80), Valid: true}
	}
	var errorMessage sql.NullString
	if s := strings.TrimSpace(in.ErrorMessage); s != "" {
		errorMessage = sql.NullString{String: truncateRunes(s, 500), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ai_generation_logs (
			divination_id, question_summary, ai_provider, model_name,
			status, duration_ms, fallback_used, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, in.DivinationID, questionSummary, in.AIProvider, in.ModelName,
		in.Status, in.DurationMs, in.FallbackUsed, errorMessage)
	return err
}

type AILogListResult struct {
	Items    []model.AIGenerationLog
	Total    int
	Page     int
	PageSize int
}

func (r *AILogRepository) ListRecent(ctx context.Context, page, pageSize int) (*AILogListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_generation_logs`).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, divination_id, question_summary, ai_provider, model_name,
			status, duration_ms, fallback_used, error_message, created_at
		FROM ai_generation_logs
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.AIGenerationLog, 0, pageSize)
	for rows.Next() {
		var item model.AIGenerationLog
		var questionSummary sql.NullString
		var errorMessage sql.NullString
		if err := rows.Scan(
			&item.ID, &item.DivinationID, &questionSummary, &item.AIProvider, &item.ModelName,
			&item.Status, &item.DurationMs, &item.FallbackUsed, &errorMessage, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if questionSummary.Valid {
			item.QuestionSummary = &questionSummary.String
		}
		if errorMessage.Valid {
			item.ErrorMessage = &errorMessage.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AILogListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

type AILogStats struct {
	TotalCount      int        `json:"total_count"`
	SuccessCount    int        `json:"success_count"`
	FailCount       int        `json:"fail_count"`
	FallbackCount   int        `json:"fallback_count"`
	AvgDurationMs   float64    `json:"avg_duration_ms"`
	LatestCreatedAt *time.Time `json:"latest_created_at,omitempty"`
}

func (r *AILogRepository) GetStats(ctx context.Context) (*AILogStats, error) {
	stats := &AILogStats{}
	var latest sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS total_count,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS success_count,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS fail_count,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS fallback_count,
			COALESCE(AVG(duration_ms), 0) AS avg_duration_ms,
			MAX(created_at) AS latest_created_at
		FROM ai_generation_logs
	`, model.AILogStatusSuccess, model.AILogStatusFailed, model.AILogStatusFallbackSuccess).Scan(
		&stats.TotalCount,
		&stats.SuccessCount,
		&stats.FailCount,
		&stats.FallbackCount,
		&stats.AvgDurationMs,
		&latest,
	)
	if err != nil {
		return nil, err
	}
	if latest.Valid {
		stats.LatestCreatedAt = &latest.Time
	}
	return stats, nil
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
