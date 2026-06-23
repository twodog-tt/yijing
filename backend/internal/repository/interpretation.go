package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
)

type InterpretationRepository struct {
	db *sql.DB
}

func NewInterpretationRepository(db *sql.DB) *InterpretationRepository {
	return &InterpretationRepository{db: db}
}

func (r *InterpretationRepository) CreateFree(ctx context.Context, divinationID int64, freeContent string, generatedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO interpretation (
			divination_id, free_content, full_content, ai_provider,
			generation_status, generated_at
		) VALUES (?, ?, NULL, ?, ?, ?)
	`, divinationID, freeContent, model.AIProviderMock, model.GenerationStatusFreeDone, generatedAt)
	return err
}

func (r *InterpretationRepository) FindByDivinationID(ctx context.Context, divinationID int64) (*model.Interpretation, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, divination_id, free_content, full_content, ai_provider,
			generation_status, generated_at, created_at, updated_at
		FROM interpretation
		WHERE divination_id = ?
		LIMIT 1
	`, divinationID)

	var item model.Interpretation
	var fullContent sql.NullString
	var generatedAt sql.NullTime
	if err := row.Scan(
		&item.ID, &item.DivinationID, &item.FreeContent, &fullContent,
		&item.AIProvider, &item.GenerationStatus, &generatedAt,
		&item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if fullContent.Valid {
		v := fullContent.String
		item.FullContent = &v
	}
	if generatedAt.Valid {
		t := generatedAt.Time
		item.GeneratedAt = &t
	}
	return &item, nil
}

func (r *InterpretationRepository) UpdateFull(ctx context.Context, divinationID int64, fullContent, aiProvider string, generatedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE interpretation
		SET full_content = ?, ai_provider = ?, generation_status = ?, generated_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE divination_id = ?
	`, fullContent, aiProvider, model.GenerationStatusFullDone, generatedAt, divinationID)
	return err
}

func (r *InterpretationRepository) HasFullContent(ctx context.Context, divinationID int64) (bool, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT full_content FROM interpretation WHERE divination_id = ? LIMIT 1
	`, divinationID)
	var full sql.NullString
	if err := row.Scan(&full); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return full.Valid && strings.TrimSpace(full.String) != "", nil
}

type UnlockRepository struct {
	db *sql.DB
}

func NewUnlockRepository(db *sql.DB) *UnlockRepository {
	return &UnlockRepository{db: db}
}

type CreateUnlockParams struct {
	DivinationID      int64
	SessionID         int64
	UnlockType        string
	MockTransactionID string
}

func (r *UnlockRepository) Create(ctx context.Context, p CreateUnlockParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO unlock_record (
			divination_id, session_id, unlock_type, unlock_status, mock_transaction_id
		) VALUES (?, ?, ?, ?, ?)
	`, p.DivinationID, p.SessionID, p.UnlockType, 1, p.MockTransactionID)
	return err
}
