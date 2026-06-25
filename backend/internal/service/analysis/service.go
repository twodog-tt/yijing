package analysis

import (
	"context"
	"errors"
	"fmt"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrInvalidParams = fmt.Errorf("invalid params")
)

type recordRepository interface {
	FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error)
	ListBySession(ctx context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error)
}

type Service struct {
	repo recordRepository
}

func NewService(repo *repository.AnalysisRepository) *Service {
	return &Service{repo: repo}
}

func NewServiceWithRepo(repo recordRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, sessionID, id int64) (*model.AnalysisRecord, error) {
	if sessionID <= 0 || id <= 0 {
		return nil, ErrInvalidParams
	}

	record, err := s.repo.FindOwnedByID(ctx, id, sessionID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrNotFound
	}
	return record, nil
}

func (s *Service) List(
	ctx context.Context,
	sessionID int64,
	moduleType *int,
	page, pageSize int,
) (*model.PaginatedAnalysisList, error) {
	if moduleType != nil {
		if err := model.ValidateModuleType(*moduleType); err != nil {
			return nil, ErrInvalidParams
		}
	}

	if sessionID <= 0 {
		normalizedPage, normalizedPageSize, err := repository.ValidateAnalysisPagination(page, pageSize)
		if err != nil {
			if errors.Is(err, repository.ErrInvalidAnalysisParams) {
				return nil, ErrInvalidParams
			}
			return nil, err
		}
		return &model.PaginatedAnalysisList{
			Items:    []model.AnalysisListItem{},
			Page:     normalizedPage,
			PageSize: normalizedPageSize,
			Total:    0,
		}, nil
	}

	items, total, normalizedPage, normalizedPageSize, err := s.repo.ListBySession(ctx, sessionID, moduleType, page, pageSize)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidAnalysisParams) {
			return nil, ErrInvalidParams
		}
		return nil, err
	}
	if items == nil {
		items = []model.AnalysisListItem{}
	}

	return &model.PaginatedAnalysisList{
		Items:    items,
		Page:     normalizedPage,
		PageSize: normalizedPageSize,
		Total:    total,
	}, nil
}
