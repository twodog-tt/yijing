package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

type DailyFortuneRepository struct {
	db *sql.DB
}

func NewDailyFortuneRepository(db *sql.DB) *DailyFortuneRepository {
	return &DailyFortuneRepository{db: db}
}

func (r *DailyFortuneRepository) FindBySessionAndDate(ctx context.Context, sessionID int64, fortuneDate string) (*model.DailyFortune, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, session_id, fortune_date, divination_id, status, created_at, updated_at
		FROM daily_fortunes
		WHERE session_id = ? AND fortune_date = ? AND status = ?
		LIMIT 1
	`, sessionID, fortuneDate, model.DailyFortuneStatusActive)

	var item model.DailyFortune
	var date time.Time
	if err := row.Scan(
		&item.ID, &item.SessionID, &date, &item.DivinationID, &item.Status, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	item.FortuneDate = date.Format("2006-01-02")
	return &item, nil
}

func (r *DailyFortuneRepository) Create(ctx context.Context, sessionID int64, fortuneDate string, divinationID int64) (*model.DailyFortune, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO daily_fortunes (session_id, fortune_date, divination_id, status)
		VALUES (?, ?, ?, ?)
	`, sessionID, fortuneDate, divinationID, model.DailyFortuneStatusActive)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrDailyFortuneDuplicate
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &model.DailyFortune{
		ID:           id,
		SessionID:    sessionID,
		FortuneDate:  fortuneDate,
		DivinationID: divinationID,
		Status:       model.DailyFortuneStatusActive,
		CreatedAt:    clock.Now(),
		UpdatedAt:    clock.Now(),
	}, nil
}

var ErrDailyFortuneDuplicate = errors.New("daily fortune duplicate")
