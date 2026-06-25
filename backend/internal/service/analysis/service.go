package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/bazi"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

var (
	ErrNotFound           = fmt.Errorf("not found")
	ErrInvalidParams      = fmt.Errorf("invalid params")
	ErrModuleNotSupported = fmt.Errorf("module not supported")
)

type recordRepository interface {
	FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error)
	ListBySession(ctx context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error)
	DeleteOwnedByID(ctx context.Context, id, sessionID int64) error
	UnlockWithFullContent(ctx context.Context, id, sessionID int64, unlockType, fullContent, aiProvider string) error
}

type fullReportGenerator interface {
	Generate(ctx context.Context, analysisID int64, resultPayload json.RawMessage, freeContent string) (content string, aiProvider string, err error)
}

type Service struct {
	repo       recordRepository
	fullReport fullReportGenerator
}

func NewService(repo *repository.AnalysisRepository, generator *bazi.FullReportGenerator) *Service {
	var gen fullReportGenerator
	if generator == nil {
		gen = bazi.NewFullReportGenerator(nil)
	} else {
		gen = generator
	}
	return &Service{repo: repo, fullReport: gen}
}

func NewServiceWithRepo(repo recordRepository) *Service {
	return NewServiceWithRepoAndGenerator(repo, bazi.NewFullReportGenerator(nil))
}

func NewServiceWithRepoAndGenerator(repo recordRepository, generator fullReportGenerator) *Service {
	if generator == nil {
		generator = bazi.NewFullReportGenerator(nil)
	}
	return &Service{repo: repo, fullReport: generator}
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

func (s *Service) Unlock(ctx context.Context, sessionID, id int64, unlockType string) (*model.AnalysisUnlockResult, error) {
	if sessionID <= 0 || id <= 0 {
		return nil, ErrInvalidParams
	}
	unlockType = strings.TrimSpace(unlockType)
	if unlockType != model.UnlockTypeRewardedVideoMock {
		return nil, ErrInvalidParams
	}

	record, err := s.repo.FindOwnedByID(ctx, id, sessionID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrNotFound
	}
	if record.ModuleType != model.ModuleTypeBazi {
		return nil, ErrModuleNotSupported
	}

	if record.UnlockStatus == model.AnalysisUnlockStatusUnlocked {
		fullContent := strings.TrimSpace(derefString(record.FullContent))
		aiProvider := derefString(record.AIProvider)
		if fullContent == "" {
			generated, provider, genErr := s.fullReport.Generate(ctx, id, record.ResultPayload, derefString(record.FreeContent))
			if genErr != nil {
				return nil, genErr
			}
			fullContent = generated
			aiProvider = provider
		}
		return &model.AnalysisUnlockResult{
			ID:               record.ID,
			UnlockStatus:     model.AnalysisUnlockStatusUnlocked,
			UnlockType:       unlockType,
			FullContent:      fullContent,
			GenerationStatus: record.GenerationStatus,
			AIProvider:       aiProvider,
		}, nil
	}

	fullContent, aiProvider, err := s.fullReport.Generate(ctx, id, record.ResultPayload, derefString(record.FreeContent))
	if err != nil {
		return nil, ErrInvalidParams
	}

	if err := s.repo.UnlockWithFullContent(ctx, id, sessionID, unlockType, fullContent, aiProvider); err != nil {
		if errors.Is(err, repository.ErrAnalysisNotFound) {
			return nil, ErrNotFound
		}
		if errors.Is(err, repository.ErrInvalidAnalysisParams) {
			return nil, ErrInvalidParams
		}
		return nil, err
	}

	return &model.AnalysisUnlockResult{
		ID:               id,
		UnlockStatus:     model.AnalysisUnlockStatusUnlocked,
		UnlockType:       unlockType,
		FullContent:      fullContent,
		GenerationStatus: model.AnalysisGenerationStatusFullDone,
		AIProvider:       aiProvider,
	}, nil
}

func (s *Service) Delete(ctx context.Context, sessionID, id int64) error {
	if sessionID <= 0 || id <= 0 {
		return ErrInvalidParams
	}

	err := s.repo.DeleteOwnedByID(ctx, id, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrAnalysisNotFound) {
			return ErrNotFound
		}
		if errors.Is(err, repository.ErrInvalidAnalysisParams) {
			return ErrInvalidParams
		}
		return err
	}
	return nil
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
	for i := range items {
		qimen.SanitizeListItem(&items[i])
	}

	return &model.PaginatedAnalysisList{
		Items:    items,
		Page:     normalizedPage,
		PageSize: normalizedPageSize,
		Total:    total,
	}, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
