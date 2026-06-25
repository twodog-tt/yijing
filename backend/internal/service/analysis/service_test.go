package analysis_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/analysis"
)

type mockAnalysisRepo struct {
	findFn   func(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error)
	listFn   func(ctx context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error)
	deleteFn func(ctx context.Context, id, sessionID int64) error
	unlockFn func(ctx context.Context, id, sessionID int64, unlockType, fullContent string) error
}

func (m *mockAnalysisRepo) FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
	return m.findFn(ctx, id, sessionID)
}

func (m *mockAnalysisRepo) ListBySession(ctx context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error) {
	return m.listFn(ctx, sessionID, moduleType, page, pageSize)
}

func (m *mockAnalysisRepo) DeleteOwnedByID(ctx context.Context, id, sessionID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, sessionID)
	}
	return nil
}

func (m *mockAnalysisRepo) UnlockWithFullContent(ctx context.Context, id, sessionID int64, unlockType, fullContent string) error {
	if m.unlockFn != nil {
		return m.unlockFn(ctx, id, sessionID, unlockType, fullContent)
	}
	return nil
}

func TestGetInvalidParams(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{})
	_, err := svc.Get(context.Background(), 0, 1)
	if !errors.Is(err, analysis.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestGetNotFound(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return nil, nil
		},
	})
	_, err := svc.Get(context.Background(), 10, 5)
	if !errors.Is(err, analysis.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetSuccess(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return &model.AnalysisRecord{ID: id, SessionID: sessionID}, nil
		},
	})
	record, err := svc.Get(context.Background(), 10, 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if record.ID != 5 || record.SessionID != 10 {
		t.Fatalf("unexpected record: %#v", record)
	}
}

func TestListRejectsUnknownModule(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{})
	unknown := 99
	_, err := svc.List(context.Background(), 10, &unknown, 1, 20)
	if !errors.Is(err, analysis.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestListMapsWrappedInvalidAnalysisParams(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		listFn: func(_ context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error) {
			return nil, 0, 0, 0, fmt.Errorf("%w: page exceeds limit", repository.ErrInvalidAnalysisParams)
		},
	})
	_, err := svc.List(context.Background(), 10, nil, model.MaxAnalysisPage+1, 20)
	if !errors.Is(err, analysis.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestListDefaultPagination(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		listFn: func(_ context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error) {
			return []model.AnalysisListItem{
				{ID: 1, ModuleType: model.ModuleTypeBazi, CreatedAt: time.Now()},
			}, 1, 1, 20, nil
		},
	})
	result, err := svc.List(context.Background(), 10, nil, 0, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.PageSize != 20 || result.Page != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected pagination result: %#v", result)
	}
}

func TestListCapsPageSizeAt100(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		listFn: func(_ context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error) {
			return nil, 0, 1, 100, nil
		},
	})
	result, err := svc.List(context.Background(), 10, nil, 1, 500)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.PageSize != 100 {
		t.Fatalf("expected page_size capped to 100, got %d", result.PageSize)
	}
}

func TestListUnknownSessionUsesSharedPaginationRules(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{})
	result, err := svc.List(context.Background(), 0, nil, 0, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Page != 1 || result.PageSize != 20 || len(result.Items) != 0 || result.Total != 0 {
		t.Fatalf("unexpected empty list result: %#v", result)
	}

	_, err = svc.List(context.Background(), 0, nil, model.MaxAnalysisPage+1, 20)
	if !errors.Is(err, analysis.ErrInvalidParams) {
		t.Fatalf("expected invalid params for oversized page, got %v", err)
	}

	result, err = svc.List(context.Background(), 0, nil, 1, 999999)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.PageSize != 100 {
		t.Fatalf("expected page_size capped to 100, got %d", result.PageSize)
	}
}

func TestDeleteSuccess(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		deleteFn: func(_ context.Context, id, sessionID int64) error {
			if id != 5 || sessionID != 10 {
				t.Fatalf("unexpected delete args id=%d sessionID=%d", id, sessionID)
			}
			return nil
		},
	})
	if err := svc.Delete(context.Background(), 10, 5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		deleteFn: func(_ context.Context, _, _ int64) error {
			return repository.ErrAnalysisNotFound
		},
	})
	err := svc.Delete(context.Background(), 10, 5)
	if !errors.Is(err, analysis.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestDeleteInvalidParams(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{})
	err := svc.Delete(context.Background(), 0, 5)
	if !errors.Is(err, analysis.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func sampleBaziRecord(id, sessionID int64) *model.AnalysisRecord {
	free := "free content"
	return &model.AnalysisRecord{
		ID:            id,
		SessionID:     sessionID,
		ModuleType:    model.ModuleTypeBazi,
		ResultPayload: json.RawMessage(`{"day_master":"甲","pillars":{"year":"乙亥","month":"己丑","day":"甲子","hour":"甲子"},"five_elements":{"wood":2,"fire":1,"earth":1,"metal":1,"water":1},"reflection_focus":"观察节奏","action_suggestions":["记录一周状态"]}`),
		FreeContent:   &free,
		UnlockStatus:  model.AnalysisUnlockStatusLocked,
	}
}

func TestUnlockSuccess(t *testing.T) {
	var unlockCalled bool
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return sampleBaziRecord(id, sessionID), nil
		},
		unlockFn: func(_ context.Context, id, sessionID int64, unlockType, fullContent string) error {
			unlockCalled = true
			if id != 5 || sessionID != 10 || unlockType != model.UnlockTypeRewardedVideoMock || fullContent == "" {
				t.Fatalf("unexpected unlock args id=%d sessionID=%d type=%q", id, sessionID, unlockType)
			}
			return nil
		},
	})

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if !unlockCalled {
		t.Fatalf("expected unlock repository call")
	}
	if result.UnlockStatus != model.AnalysisUnlockStatusUnlocked || result.FullContent == "" {
		t.Fatalf("unexpected unlock result: %#v", result)
	}
}

func TestUnlockRejectsInvalidUnlockType(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{})
	_, err := svc.Unlock(context.Background(), 10, 5, "mock_button")
	if !errors.Is(err, analysis.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestUnlockAlreadyUnlockedSkipsRepositoryUpdate(t *testing.T) {
	full := "existing full content"
	record := sampleBaziRecord(5, 10)
	record.UnlockStatus = model.AnalysisUnlockStatusUnlocked
	record.FullContent = &full

	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return record, nil
		},
		unlockFn: func(_ context.Context, _, _ int64, _, _ string) error {
			t.Fatalf("expected no repository unlock for already unlocked record")
			return nil
		},
	})

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if result.FullContent != full {
		t.Fatalf("expected existing full content, got %q", result.FullContent)
	}
}

func TestUnlockModuleNotSupported(t *testing.T) {
	record := sampleBaziRecord(5, 10)
	record.ModuleType = model.ModuleTypeQimen
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return record, nil
		},
	})
	_, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if !errors.Is(err, analysis.ErrModuleNotSupported) {
		t.Fatalf("expected module not supported, got %v", err)
	}
}
