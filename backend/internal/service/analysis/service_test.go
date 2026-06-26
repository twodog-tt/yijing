package analysis_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/analysis"
	"github.com/wangxintong/yijing/backend/internal/service/bazi"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

type mockAnalysisRepo struct {
	findFn   func(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error)
	listFn   func(ctx context.Context, sessionID int64, moduleType *int, page, pageSize int) ([]model.AnalysisListItem, int64, int, int, error)
	deleteFn func(ctx context.Context, id, sessionID int64) error
	unlockFn func(ctx context.Context, id, sessionID int64, unlockType, fullContent, aiProvider string) error
	repairFn func(ctx context.Context, id, sessionID int64, fullContent, aiProvider string) error
}

type stubFullReportGenerator struct {
	generateFn func(ctx context.Context, analysisID int64, resultPayload json.RawMessage, freeContent string) (string, string, error)
	calls      int
}

func (s *stubFullReportGenerator) Generate(ctx context.Context, analysisID int64, resultPayload json.RawMessage, freeContent string) (string, string, error) {
	s.calls++
	if s.generateFn != nil {
		return s.generateFn(ctx, analysisID, resultPayload, freeContent)
	}
	content, err := bazi.BuildFullContent(resultPayload, freeContent)
	if err != nil {
		return "", "", err
	}
	return content, model.AIProviderTemplateFallback, nil
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

func (m *mockAnalysisRepo) UnlockWithFullContent(ctx context.Context, id, sessionID int64, unlockType, fullContent, aiProvider string) error {
	if m.unlockFn != nil {
		return m.unlockFn(ctx, id, sessionID, unlockType, fullContent, aiProvider)
	}
	return nil
}

func (m *mockAnalysisRepo) UpdateUnlockedFullContent(ctx context.Context, id, sessionID int64, fullContent, aiProvider string) error {
	if m.repairFn != nil {
		return m.repairFn(ctx, id, sessionID, fullContent, aiProvider)
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

func TestUnlockSuccessUsesTemplateFallbackByDefault(t *testing.T) {
	var savedProvider string
	gen := &stubFullReportGenerator{}
	svc := analysis.NewServiceWithRepoAndGenerator(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return sampleBaziRecord(id, sessionID), nil
		},
		unlockFn: func(_ context.Context, id, sessionID int64, unlockType, fullContent, aiProvider string) error {
			savedProvider = aiProvider
			if id != 5 || sessionID != 10 || unlockType != model.UnlockTypeRewardedVideoMock || fullContent == "" {
				t.Fatalf("unexpected unlock args id=%d sessionID=%d type=%q", id, sessionID, unlockType)
			}
			return nil
		},
	}, gen)

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if gen.calls != 1 {
		t.Fatalf("expected generator called once, got %d", gen.calls)
	}
	if savedProvider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback provider, got %q", savedProvider)
	}
	if result.AIProvider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback in result, got %q", result.AIProvider)
	}
	if result.UnlockStatus != model.AnalysisUnlockStatusUnlocked || result.FullContent == "" {
		t.Fatalf("unexpected unlock result: %#v", result)
	}
}

func TestUnlockSuccessUsesDeepSeekProvider(t *testing.T) {
	aiContent := strings.Repeat("完整报告内容。", 30) + "\n免责声明：仅供学习参考。"
	gen := &stubFullReportGenerator{
		generateFn: func(_ context.Context, _ int64, _ json.RawMessage, _ string) (string, string, error) {
			return aiContent, model.AIProviderDeepSeek, nil
		},
	}
	var savedProvider string
	svc := analysis.NewServiceWithRepoAndGenerator(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return sampleBaziRecord(id, sessionID), nil
		},
		unlockFn: func(_ context.Context, _, _ int64, _, fullContent, aiProvider string) error {
			savedProvider = aiProvider
			if fullContent != aiContent {
				t.Fatalf("unexpected full content saved")
			}
			return nil
		},
	}, gen)

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if savedProvider != model.AIProviderDeepSeek {
		t.Fatalf("expected deepseek provider, got %q", savedProvider)
	}
	if result.AIProvider != model.AIProviderDeepSeek {
		t.Fatalf("expected deepseek in result, got %q", result.AIProvider)
	}
	if result.FullContent != aiContent {
		t.Fatalf("unexpected full content in result")
	}
}

func TestUnlockDeepSeekFailureFallsBackToTemplate(t *testing.T) {
	gen := &stubFullReportGenerator{
		generateFn: func(_ context.Context, _ int64, resultPayload json.RawMessage, freeContent string) (string, string, error) {
			content, err := bazi.BuildFullContent(resultPayload, freeContent)
			if err != nil {
				return "", "", err
			}
			return content, model.AIProviderTemplateFallback, nil
		},
	}
	var savedProvider string
	svc := analysis.NewServiceWithRepoAndGenerator(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return sampleBaziRecord(id, sessionID), nil
		},
		unlockFn: func(_ context.Context, _, _ int64, _, _, aiProvider string) error {
			savedProvider = aiProvider
			return nil
		},
	}, gen)

	_, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if savedProvider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback provider, got %q", savedProvider)
	}
}

func TestUnlockRejectsInvalidUnlockType(t *testing.T) {
	svc := analysis.NewServiceWithRepo(&mockAnalysisRepo{})
	_, err := svc.Unlock(context.Background(), 10, 5, "mock_button")
	if !errors.Is(err, analysis.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestUnlockAlreadyUnlockedSkipsRepositoryUpdateAndGenerator(t *testing.T) {
	full := "existing full content"
	record := sampleBaziRecord(5, 10)
	record.UnlockStatus = model.AnalysisUnlockStatusUnlocked
	record.FullContent = &full
	provider := model.AIProviderDeepSeek
	record.AIProvider = &provider
	gen := &stubFullReportGenerator{
		generateFn: func(_ context.Context, _ int64, _ json.RawMessage, _ string) (string, string, error) {
			t.Fatalf("expected no generator call for already unlocked record")
			return "", "", nil
		},
	}

	svc := analysis.NewServiceWithRepoAndGenerator(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return record, nil
		},
		unlockFn: func(_ context.Context, _, _ int64, _, _, _ string) error {
			t.Fatalf("expected no repository unlock for already unlocked record")
			return nil
		},
	}, gen)

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if result.FullContent != full {
		t.Fatalf("expected existing full content, got %q", result.FullContent)
	}
	if result.AIProvider != model.AIProviderDeepSeek {
		t.Fatalf("expected existing ai_provider, got %q", result.AIProvider)
	}
}

func TestNewServiceNilGeneratorUsesTemplateFallback(t *testing.T) {
	svc := analysis.NewServiceWithRepoAndGenerator(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return sampleBaziRecord(id, sessionID), nil
		},
		unlockFn: func(_ context.Context, _, _ int64, _, _, _ string) error { return nil },
	}, nil)
	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if result.AIProvider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", result.AIProvider)
	}
}

func TestUnlockModuleNotSupported(t *testing.T) {
	record := sampleBaziRecord(5, 10)
	record.ModuleType = 99
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

func TestUnlockQimenSuccessUsesTemplateFallbackByDefault(t *testing.T) {
	qimenGen := &stubFullReportGenerator{
		generateFn: func(_ context.Context, _ int64, resultPayload json.RawMessage, freeContent string) (string, string, error) {
			content, err := qimen.BuildFullContent(resultPayload, freeContent)
			if err != nil {
				return "", "", err
			}
			return content, model.AIProviderTemplateFallback, nil
		},
	}
	svc := analysis.NewServiceWithFullReportGenerators(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return sampleQimenRecord(id, sessionID), nil
		},
		unlockFn: func(_ context.Context, id, sessionID int64, unlockType, fullContent, aiProvider string) error {
			if id != 5 || sessionID != 10 || unlockType != model.UnlockTypeRewardedVideoMock || fullContent == "" {
				t.Fatalf("unexpected unlock args id=%d sessionID=%d", id, sessionID)
			}
			if aiProvider != model.AIProviderTemplateFallback {
				t.Fatalf("expected template_fallback, got %q", aiProvider)
			}
			return nil
		},
	}, nil, qimenGen)

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if qimenGen.calls != 1 {
		t.Fatalf("expected generator called once, got %d", qimenGen.calls)
	}
	if result.FullContent == "" || result.AIProvider != model.AIProviderTemplateFallback {
		t.Fatalf("unexpected unlock result: %#v", result)
	}
}

func TestUnlockQimenRepairsMissingFullContentWhenAlreadyUnlocked(t *testing.T) {
	record := sampleQimenRecord(5, 10)
	record.UnlockStatus = model.AnalysisUnlockStatusUnlocked
	qimenGen := &stubFullReportGenerator{
		generateFn: func(_ context.Context, _ int64, resultPayload json.RawMessage, freeContent string) (string, string, error) {
			content, err := qimen.BuildFullContent(resultPayload, freeContent)
			if err != nil {
				return "", "", err
			}
			return content, model.AIProviderTemplateFallback, nil
		},
	}
	repairCalled := false
	svc := analysis.NewServiceWithFullReportGenerators(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return record, nil
		},
		unlockFn: func(_ context.Context, _, _ int64, _, _, _ string) error {
			t.Fatalf("expected no repository unlock for already unlocked record")
			return nil
		},
		repairFn: func(_ context.Context, id, sessionID int64, fullContent, aiProvider string) error {
			repairCalled = true
			if id != 5 || sessionID != 10 || fullContent == "" {
				t.Fatalf("unexpected repair args id=%d sessionID=%d", id, sessionID)
			}
			if aiProvider != model.AIProviderTemplateFallback {
				t.Fatalf("expected template_fallback, got %q", aiProvider)
			}
			return nil
		},
	}, nil, qimenGen)

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if !repairCalled {
		t.Fatalf("expected repair persistence call")
	}
	if result.FullContent == "" || !strings.Contains(result.FullContent, "【2. 局势梳理展开】") {
		t.Fatalf("expected repaired full content, got %q", result.FullContent)
	}
}

func TestUnlockQimenAlreadyUnlockedSkipsGenerator(t *testing.T) {
	full := "existing qimen full content"
	record := sampleQimenRecord(5, 10)
	record.UnlockStatus = model.AnalysisUnlockStatusUnlocked
	record.FullContent = &full
	provider := model.AIProviderDeepSeek
	record.AIProvider = &provider
	qimenGen := &stubFullReportGenerator{
		generateFn: func(_ context.Context, _ int64, _ json.RawMessage, _ string) (string, string, error) {
			t.Fatalf("expected no generator call for already unlocked qimen record")
			return "", "", nil
		},
	}
	svc := analysis.NewServiceWithFullReportGenerators(&mockAnalysisRepo{
		findFn: func(_ context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
			return record, nil
		},
		unlockFn: func(_ context.Context, _, _ int64, _, _, _ string) error {
			t.Fatalf("expected no repository unlock for already unlocked record")
			return nil
		},
	}, nil, qimenGen)

	result, err := svc.Unlock(context.Background(), 10, 5, model.UnlockTypeRewardedVideoMock)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if result.FullContent != full {
		t.Fatalf("expected existing full content, got %q", result.FullContent)
	}
}

func sampleQimenRecord(id, sessionID int64) *model.AnalysisRecord {
	free := "【局势梳理】\n当前局势整理。"
	return &model.AnalysisRecord{
		ID:         id,
		SessionID:  sessionID,
		ModuleType: model.ModuleTypeQimen,
		ResultPayload: json.RawMessage(`{
			"algorithm_version":"qimen-simple-v1",
			"method_note":"简化奇门规则",
			"question_summary":"用户问题已用于本次局势梳理",
			"category":"career",
			"time_context":{"time_bucket":"day"},
			"situation_overview":"当前局势更像是在整理方向与节奏。",
			"risk_observations":["过度依赖单一结论。"],
			"action_pacing":"建议分步推进。",
			"reflection_questions":["我真正想推进的核心目标是什么？"],
			"action_suggestions":["用一页纸写下现状。"]
		}`),
		FreeContent:  &free,
		UnlockStatus: model.AnalysisUnlockStatusLocked,
	}
}

func TestGenerateFullReportQimenUsesTemplateFallbackByDefault(t *testing.T) {
	svc := analysis.NewServiceWithFullReportGenerators(&mockAnalysisRepo{}, nil, nil)
	content, provider, err := svc.GenerateFullReport(context.Background(), sampleQimenRecord(1, 10))
	if err != nil {
		t.Fatalf("GenerateFullReport: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【2. 局势梳理展开】") {
		t.Fatalf("expected qimen template sections, got %q", content)
	}
}

func TestGenerateFullReportRejectsMalformedQimenPayload(t *testing.T) {
	record := sampleQimenRecord(1, 10)
	record.ResultPayload = json.RawMessage(`{"category":"career"}`)
	svc := analysis.NewServiceWithFullReportGenerators(&mockAnalysisRepo{}, nil, nil)
	_, _, err := svc.GenerateFullReport(context.Background(), record)
	if err == nil {
		t.Fatalf("expected error for malformed qimen payload")
	}
}

func TestGenerateFullReportBaziStillWorks(t *testing.T) {
	svc := analysis.NewServiceWithFullReportGenerators(&mockAnalysisRepo{}, nil, nil)
	content, provider, err := svc.GenerateFullReport(context.Background(), sampleBaziRecord(1, 10))
	if err != nil {
		t.Fatalf("GenerateFullReport: %v", err)
	}
	if provider != model.AIProviderTemplateFallback {
		t.Fatalf("expected template_fallback, got %q", provider)
	}
	if !strings.Contains(content, "【1. 简化干支示意】") {
		t.Fatalf("expected bazi template sections")
	}
}
