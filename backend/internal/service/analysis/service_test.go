package analysis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/analysis"
)

type mockAnalysisRepo struct {
	findFn func(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error)
	listFn func(ctx context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error)
}

func (m *mockAnalysisRepo) FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
	return m.findFn(ctx, id, sessionID)
}

func (m *mockAnalysisRepo) ListBySession(ctx context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error) {
	return m.listFn(ctx, sessionID, moduleType, page, pageSize)
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
