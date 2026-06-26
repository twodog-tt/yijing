package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
)

var ErrInvalidAnalysisParams = errors.New("invalid analysis params")

var ErrAnalysisNotFound = errors.New("analysis not found")

type AnalysisRepository struct {
	db *sql.DB
}

func NewAnalysisRepository(db *sql.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

type CreateAnalysisParams struct {
	SessionID        int64
	ModuleType       int
	AlgorithmVersion string
	CategoryID       *int64
	Question         *string
	InputPayload     json.RawMessage
	ResultPayload    json.RawMessage
}

type CreateAnalysisWithFreeContentParams struct {
	CreateAnalysisParams
	FreeContent string
}

func (r *AnalysisRepository) Create(ctx context.Context, p CreateAnalysisParams) (int64, error) {
	algorithmVersion, err := validateCreateAnalysisParams(p)
	if err != nil {
		return 0, err
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO analysis_records (
			session_id, module_type, algorithm_version, category_id, question,
			input_payload, result_payload,
			free_content, full_content, unlock_status, unlock_type, ai_provider, generation_status, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULL, NULL, 0, NULL, NULL, 0, 1)
	`,
		p.SessionID,
		p.ModuleType,
		algorithmVersion,
		nullInt64Ptr(p.CategoryID),
		nullStringPtr(p.Question),
		p.InputPayload,
		nullJSON(p.ResultPayload),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *AnalysisRepository) CreateWithFreeContent(ctx context.Context, p CreateAnalysisWithFreeContentParams) (int64, error) {
	algorithmVersion, err := validateCreateAnalysisParams(p.CreateAnalysisParams)
	if err != nil {
		return 0, err
	}
	freeContent := strings.TrimSpace(p.FreeContent)
	if freeContent == "" {
		return 0, fmt.Errorf("%w: free_content required", ErrInvalidAnalysisParams)
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO analysis_records (
			session_id, module_type, algorithm_version, category_id, question,
			input_payload, result_payload,
			free_content, full_content, unlock_status, unlock_type, ai_provider, generation_status, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL, 0, NULL, NULL, ?, 1)
	`,
		p.SessionID,
		p.ModuleType,
		algorithmVersion,
		nullInt64Ptr(p.CategoryID),
		nullStringPtr(p.Question),
		p.InputPayload,
		nullJSON(p.ResultPayload),
		freeContent,
		model.AnalysisGenerationStatusFreeDone,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *AnalysisRepository) FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
	if id <= 0 || sessionID <= 0 {
		return nil, nil
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, session_id, module_type, algorithm_version, category_id, question,
			input_payload, result_payload, free_content, full_content,
			unlock_status, unlock_type, ai_provider, generation_status, status,
			created_at, updated_at
		FROM analysis_records
		WHERE id = ? AND session_id = ? AND status = ?
		LIMIT 1
	`, id, sessionID, model.AnalysisStatusActive)

	return scanAnalysisRecord(row)
}

func (r *AnalysisRepository) UnlockWithFullContent(
	ctx context.Context,
	id, sessionID int64,
	unlockType, fullContent, aiProvider string,
) error {
	if id <= 0 || sessionID <= 0 {
		return ErrInvalidAnalysisParams
	}
	unlockType = strings.TrimSpace(unlockType)
	fullContent = strings.TrimSpace(fullContent)
	aiProvider = strings.TrimSpace(aiProvider)
	if unlockType == "" || fullContent == "" || aiProvider == "" {
		return ErrInvalidAnalysisParams
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE analysis_records
	 SET unlock_status = ?, unlock_type = ?, full_content = ?, generation_status = ?, ai_provider = ?, updated_at = NOW()
		WHERE id = ? AND session_id = ? AND status = ? AND unlock_status = ?
	`, model.AnalysisUnlockStatusUnlocked, unlockType, fullContent, model.AnalysisGenerationStatusFullDone, aiProvider,
		id, sessionID, model.AnalysisStatusActive, model.AnalysisUnlockStatusLocked)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAnalysisNotFound
	}
	return nil
}

func (r *AnalysisRepository) UpdateUnlockedFullContent(
	ctx context.Context,
	id, sessionID int64,
	fullContent, aiProvider string,
) error {
	if id <= 0 || sessionID <= 0 {
		return ErrInvalidAnalysisParams
	}
	fullContent = strings.TrimSpace(fullContent)
	aiProvider = strings.TrimSpace(aiProvider)
	if fullContent == "" || aiProvider == "" {
		return ErrInvalidAnalysisParams
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE analysis_records
		SET full_content = ?, ai_provider = ?, generation_status = ?, updated_at = NOW()
		WHERE id = ? AND session_id = ? AND status = ? AND unlock_status = ?
		  AND (full_content IS NULL OR full_content = '')
	`, fullContent, aiProvider, model.AnalysisGenerationStatusFullDone,
		id, sessionID, model.AnalysisStatusActive, model.AnalysisUnlockStatusUnlocked)
	return err
}

func (r *AnalysisRepository) DeleteOwnedByID(ctx context.Context, id, sessionID int64) error {
	if id <= 0 || sessionID <= 0 {
		return ErrInvalidAnalysisParams
	}

	res, err := r.db.ExecContext(ctx, `
		DELETE FROM analysis_records
		WHERE id = ? AND session_id = ? AND status = ?
	`, id, sessionID, model.AnalysisStatusActive)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAnalysisNotFound
	}
	return nil
}

func (r *AnalysisRepository) UpdateFreeContent(ctx context.Context, id, sessionID int64, freeContent string) error {
	if id <= 0 || sessionID <= 0 {
		return ErrInvalidAnalysisParams
	}
	freeContent = strings.TrimSpace(freeContent)
	if freeContent == "" {
		return fmt.Errorf("%w: free_content required", ErrInvalidAnalysisParams)
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE analysis_records
		SET free_content = ?, generation_status = ?, updated_at = NOW()
		WHERE id = ? AND session_id = ? AND status = ?
	`, freeContent, model.AnalysisGenerationStatusFreeDone, id, sessionID, model.AnalysisStatusActive)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrInvalidAnalysisParams
	}
	return nil
}

func (r *AnalysisRepository) ListBySession(
	ctx context.Context,
	sessionID int64,
	moduleType *int,
	page, pageSize int,
) ([]model.AnalysisListItem, int64, int, int, error) {
	if sessionID <= 0 {
		return nil, 0, 0, 0, ErrInvalidAnalysisParams
	}
	if moduleType != nil {
		if err := model.ValidateModuleType(*moduleType); err != nil {
			return nil, 0, 0, 0, err
		}
	}

	normalizedPage, normalizedPageSize, offset, err := computeAnalysisPagination(page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	countSQL := `
		SELECT COUNT(*)
		FROM analysis_records
		WHERE session_id = ? AND status = ?
	`
	countArgs := []any{sessionID, model.AnalysisStatusActive}
	if moduleType != nil {
		countSQL += ` AND module_type = ?`
		countArgs = append(countArgs, *moduleType)
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, 0, 0, err
	}

	listSQL := `
		SELECT id, module_type, algorithm_version, category_id, question,
			unlock_status, generation_status, created_at
		FROM analysis_records
		WHERE session_id = ? AND status = ?
	`
	listArgs := []any{sessionID, model.AnalysisStatusActive}
	if moduleType != nil {
		listSQL += ` AND module_type = ?`
		listArgs = append(listArgs, *moduleType)
	}
	listSQL += `
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`
	listArgs = append(listArgs, normalizedPageSize, offset)

	rows, err := r.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	defer rows.Close()

	items := make([]model.AnalysisListItem, 0)
	for rows.Next() {
		var item model.AnalysisListItem
		var categoryID sql.NullInt64
		var question sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.ModuleType,
			&item.AlgorithmVersion,
			&categoryID,
			&question,
			&item.UnlockStatus,
			&item.GenerationStatus,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, 0, 0, err
		}
		if categoryID.Valid {
			v := categoryID.Int64
			item.CategoryID = &v
		}
		if question.Valid {
			v := question.String
			item.Question = &v
		}
		items = append(items, item)
	}
	return items, total, normalizedPage, normalizedPageSize, rows.Err()
}

func validateCreateAnalysisParams(p CreateAnalysisParams) (string, error) {
	if p.SessionID <= 0 {
		return "", fmt.Errorf("%w: session_id required", ErrInvalidAnalysisParams)
	}
	if err := model.ValidateModuleType(p.ModuleType); err != nil {
		return "", err
	}
	algorithmVersion := strings.TrimSpace(p.AlgorithmVersion)
	if err := model.ValidateAlgorithmVersion(p.ModuleType, algorithmVersion); err != nil {
		return "", err
	}
	if err := model.ValidateJSONObjectPayload(p.InputPayload, model.MaxAnalysisPayloadBytes); err != nil {
		return "", fmt.Errorf("%w: input_payload must be a JSON object", ErrInvalidAnalysisParams)
	}
	if len(p.ResultPayload) > 0 {
		if err := model.ValidateJSONObjectPayload(p.ResultPayload, model.MaxAnalysisPayloadBytes); err != nil {
			return "", fmt.Errorf("%w: result_payload must be a JSON object", ErrInvalidAnalysisParams)
		}
	}
	return algorithmVersion, nil
}

func computeAnalysisPagination(page, pageSize int) (normalizedPage, normalizedPageSize int, offset int64, err error) {
	if page < 1 {
		page = 1
	}
	if page > model.MaxAnalysisPage {
		return 0, 0, 0, fmt.Errorf("%w: page exceeds limit", ErrInvalidAnalysisParams)
	}
	if pageSize < 1 {
		pageSize = model.DefaultAnalysisPageSize
	}
	if pageSize > model.MaxAnalysisPageSize {
		pageSize = model.MaxAnalysisPageSize
	}

	page64 := int64(page - 1)
	size64 := int64(pageSize)
	if page64 > math.MaxInt64/size64 {
		return 0, 0, 0, fmt.Errorf("%w: pagination overflow", ErrInvalidAnalysisParams)
	}
	return page, pageSize, page64 * size64, nil
}

// ValidateAnalysisPagination normalizes page/page_size using the shared analysis list rules.
func ValidateAnalysisPagination(page, pageSize int) (normalizedPage, normalizedPageSize int, err error) {
	normalizedPage, normalizedPageSize, _, err = computeAnalysisPagination(page, pageSize)
	return normalizedPage, normalizedPageSize, err
}

func scanAnalysisRecord(row *sql.Row) (*model.AnalysisRecord, error) {
	var record model.AnalysisRecord
	var categoryID sql.NullInt64
	var question sql.NullString
	var resultPayload []byte
	var freeContent sql.NullString
	var fullContent sql.NullString
	var unlockType sql.NullString
	var aiProvider sql.NullString

	if err := row.Scan(
		&record.ID,
		&record.SessionID,
		&record.ModuleType,
		&record.AlgorithmVersion,
		&categoryID,
		&question,
		&record.InputPayload,
		&resultPayload,
		&freeContent,
		&fullContent,
		&record.UnlockStatus,
		&unlockType,
		&aiProvider,
		&record.GenerationStatus,
		&record.Status,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if categoryID.Valid {
		v := categoryID.Int64
		record.CategoryID = &v
	}
	if question.Valid {
		v := question.String
		record.Question = &v
	}
	if len(resultPayload) > 0 {
		record.ResultPayload = json.RawMessage(resultPayload)
	}
	if freeContent.Valid {
		v := freeContent.String
		record.FreeContent = &v
	}
	if fullContent.Valid {
		v := fullContent.String
		record.FullContent = &v
	}
	if unlockType.Valid {
		v := unlockType.String
		record.UnlockType = &v
	}
	if aiProvider.Valid {
		v := aiProvider.String
		record.AIProvider = &v
	}
	return &record, nil
}

func nullInt64Ptr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullStringPtr(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func nullJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
