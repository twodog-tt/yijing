package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/wangxintong/yijing/backend/internal/model"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) FindByKey(ctx context.Context, sessionKey string) (*model.Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, session_key, status, created_at
		FROM user_sessions
		WHERE session_key = ? AND status = ?
		LIMIT 1
	`, sessionKey, model.SessionStatusActive)

	var s model.Session
	if err := row.Scan(&s.ID, &s.SessionKey, &s.Status, &s.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) Create(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO user_sessions (session_key, client_info, status)
		VALUES (?, ?, ?)
	`, sessionKey, nullString(clientInfo), model.SessionStatusActive)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &model.Session{ID: id, SessionKey: sessionKey, Status: model.SessionStatusActive}, nil
}

func (r *SessionRepository) Upsert(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error) {
	existing, err := r.FindByKey(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	return r.Create(ctx, sessionKey, clientInfo)
}

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) ListActive(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, description, sort_order, status
		FROM matter_category
		WHERE status = ?
		ORDER BY sort_order ASC, id ASC
	`, model.CategoryStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.SortOrder, &c.Status); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *CategoryRepository) FindActiveByID(ctx context.Context, id int64) (*model.Category, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, description, sort_order, status
		FROM matter_category
		WHERE id = ? AND status = ?
		LIMIT 1
	`, id, model.CategoryStatusActive)

	var c model.Category
	if err := row.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.SortOrder, &c.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

type HexagramRepository struct {
	db *sql.DB
}

func NewHexagramRepository(db *sql.DB) *HexagramRepository {
	return &HexagramRepository{db: db}
}

func (r *HexagramRepository) FindByBinaryCode(ctx context.Context, binaryCode string) (*model.Hexagram, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, number, name, full_name, upper_trigram, lower_trigram, binary_code, summary
		FROM hexagram
		WHERE binary_code = ? AND status = 1
		LIMIT 1
	`, binaryCode)

	var h model.Hexagram
	if err := row.Scan(&h.ID, &h.Number, &h.Name, &h.FullName, &h.UpperTrigram, &h.LowerTrigram, &h.BinaryCode, &h.Summary); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("hexagram not found for binary_code %s", binaryCode)
		}
		return nil, err
	}
	return &h, nil
}

func (r *HexagramRepository) FindByID(ctx context.Context, id int64) (*model.Hexagram, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, number, name, full_name, upper_trigram, lower_trigram, binary_code, summary
		FROM hexagram
		WHERE id = ? AND status = 1
		LIMIT 1
	`, id)

	var h model.Hexagram
	if err := row.Scan(&h.ID, &h.Number, &h.Name, &h.FullName, &h.UpperTrigram, &h.LowerTrigram, &h.BinaryCode, &h.Summary); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &h, nil
}

type SensitiveRepository struct {
	db *sql.DB
}

func NewSensitiveRepository(db *sql.DB) *SensitiveRepository {
	return &SensitiveRepository{db: db}
}

func (r *SensitiveRepository) ListActiveKeywords(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT keyword FROM sensitive_keyword WHERE status = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []string
	for rows.Next() {
		var kw string
		if err := rows.Scan(&kw); err != nil {
			return nil, err
		}
		keywords = append(keywords, kw)
	}
	return keywords, rows.Err()
}

type DivinationRepository struct {
	db *sql.DB
}

func NewDivinationRepository(db *sql.DB) *DivinationRepository {
	return &DivinationRepository{db: db}
}

type CreateDivinationParams struct {
	SessionID         int64
	CategoryID        int64
	Question          string
	Method            string
	PrimaryHexagramID int64
	ChangedHexagramID int64
	MovingLinesJSON   string
	LineSnapshotJSON  string
	Seed              string
	Lines             []model.Line
}

func (r *DivinationRepository) Create(ctx context.Context, p CreateDivinationParams) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO divination_record (
			session_id, category_id, question, method,
			primary_hexagram_id, changed_hexagram_id,
			moving_lines, line_snapshot, seed,
			unlock_status, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.SessionID, p.CategoryID, p.Question, p.Method,
		p.PrimaryHexagramID, p.ChangedHexagramID,
		p.MovingLinesJSON, p.LineSnapshotJSON, p.Seed,
		model.UnlockStatusLocked, model.DivinationStatusActive)
	if err != nil {
		return 0, err
	}

	divinationID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, line := range p.Lines {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO divination_line (divination_id, line_position, line_value, is_yang, is_moving)
			VALUES (?, ?, ?, ?, ?)
		`, divinationID, line.Position, line.Value, line.IsYang, line.IsMoving)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return divinationID, nil
}

func (r *DivinationRepository) FindByID(ctx context.Context, id int64) (*model.Divination, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, session_id, category_id, question, method,
			primary_hexagram_id, changed_hexagram_id,
			moving_lines, line_snapshot, seed,
			unlock_status, status, created_at
		FROM divination_record
		WHERE id = ? AND status = ?
	`, id, model.DivinationStatusActive)

	var d model.Divination
	if err := row.Scan(
		&d.ID, &d.SessionID, &d.CategoryID, &d.Question, &d.Method,
		&d.PrimaryHexagramID, &d.ChangedHexagramID,
		&d.MovingLines, &d.LineSnapshot, &d.Seed,
		&d.UnlockStatus, &d.Status, &d.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DivinationRepository) ListBySession(ctx context.Context, sessionID int64, page, pageSize int) ([]model.Divination, int64, error) {
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

	var total int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM divination_record
		WHERE session_id = ? AND status = ?
	`, sessionID, model.DivinationStatusActive).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, session_id, category_id, question, method,
			primary_hexagram_id, changed_hexagram_id,
			moving_lines, line_snapshot, seed,
			unlock_status, status, created_at
		FROM divination_record
		WHERE session_id = ? AND status = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, sessionID, model.DivinationStatusActive, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Divination
	for rows.Next() {
		var d model.Divination
		if err := rows.Scan(
			&d.ID, &d.SessionID, &d.CategoryID, &d.Question, &d.Method,
			&d.PrimaryHexagramID, &d.ChangedHexagramID,
			&d.MovingLines, &d.LineSnapshot, &d.Seed,
			&d.UnlockStatus, &d.Status, &d.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, rows.Err()
}

func (r *DivinationRepository) UpdateUnlockStatus(ctx context.Context, id int64, unlockStatus int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE divination_record
		SET unlock_status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = ?
	`, unlockStatus, id, model.DivinationStatusActive)
	return err
}

var (
	ErrInvalidDivinationParams = errors.New("invalid divination params")
	ErrDivinationNotFound      = errors.New("divination not found")
)

func (r *DivinationRepository) SoftDeleteOwnedByID(ctx context.Context, id, sessionID int64) error {
	if id <= 0 || sessionID <= 0 {
		return ErrInvalidDivinationParams
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE divination_record
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND session_id = ? AND status = ?
	`, model.DivinationStatusDeleted, id, sessionID, model.DivinationStatusActive)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrDivinationNotFound
	}
	return nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
