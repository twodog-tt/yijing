package qimen_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

type mockSessionRepo struct {
	upsertFn func(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error)
}

func (m *mockSessionRepo) Upsert(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error) {
	return m.upsertFn(ctx, sessionKey, clientInfo)
}

type mockAnalysisRepo struct {
	createWithFreeContentFn func(ctx context.Context, p repository.CreateAnalysisWithFreeContentParams) (int64, error)
	findOwnedByIDFn         func(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error)
	lastCreateParams        repository.CreateAnalysisWithFreeContentParams
	createCalls             int
}

func (m *mockAnalysisRepo) CreateWithFreeContent(ctx context.Context, p repository.CreateAnalysisWithFreeContentParams) (int64, error) {
	m.createCalls++
	m.lastCreateParams = p
	if m.createWithFreeContentFn != nil {
		return m.createWithFreeContentFn(ctx, p)
	}
	return 2001, nil
}

func (m *mockAnalysisRepo) FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
	if m.findOwnedByIDFn != nil {
		return m.findOwnedByIDFn(ctx, id, sessionID)
	}
	free := m.lastCreateParams.FreeContent
	question := m.lastCreateParams.Question
	return &model.AnalysisRecord{
		ID:               id,
		SessionID:        sessionID,
		ModuleType:       model.ModuleTypeQimen,
		AlgorithmVersion: model.AlgorithmVersionQimenSimpleV1,
		Question:         question,
		InputPayload:     m.lastCreateParams.InputPayload,
		ResultPayload:    m.lastCreateParams.ResultPayload,
		FreeContent:      &free,
		UnlockStatus:     model.AnalysisUnlockStatusLocked,
		GenerationStatus: model.AnalysisGenerationStatusFreeDone,
	}, nil
}

type mockRiskChecker struct {
	blocked bool
	err     error
}

func (m *mockRiskChecker) CheckQuestion(_ context.Context, _ string) (bool, error) {
	return m.blocked, m.err
}

func validCreateParams() qimen.CreateParams {
	return qimen.CreateParams{
		SessionKey:        "sess-qimen",
		Question:          "我最近适合推进这个计划吗？",
		Category:          "career",
		ConfirmDisclaimer: true,
	}
}

func TestCreateSuccess(t *testing.T) {
	analysisRepo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, sessionKey, _ string) (*model.Session, error) {
			return &model.Session{ID: 10, SessionKey: sessionKey}, nil
		},
	}, analysisRepo, nil)

	record, err := svc.Create(context.Background(), validCreateParams())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if record == nil || record.ID != 2001 {
		t.Fatalf("unexpected record: %#v", record)
	}
	if analysisRepo.createCalls != 1 {
		t.Fatalf("expected single insert, got %d", analysisRepo.createCalls)
	}
	if analysisRepo.lastCreateParams.ModuleType != model.ModuleTypeQimen {
		t.Fatalf("expected qimen module, got %d", analysisRepo.lastCreateParams.ModuleType)
	}
	if analysisRepo.lastCreateParams.AlgorithmVersion != model.AlgorithmVersionQimenSimpleV1 {
		t.Fatalf("unexpected algorithm version")
	}
	if strings.TrimSpace(analysisRepo.lastCreateParams.FreeContent) == "" {
		t.Fatalf("expected free_content")
	}
	if !strings.Contains(analysisRepo.lastCreateParams.FreeContent, "qimen-simple-v1") {
		t.Fatalf("expected disclaimer in free_content")
	}

	var input map[string]any
	if err := json.Unmarshal(analysisRepo.lastCreateParams.InputPayload, &input); err != nil {
		t.Fatalf("input payload: %v", err)
	}
	if input["question"] == "" || input["confirm_disclaimer"] != true {
		t.Fatalf("unexpected input payload: %#v", input)
	}
	for _, forbidden := range []string{"name", "phone", "gender", "address", "id_card"} {
		if _, ok := input[forbidden]; ok {
			t.Fatalf("must not store %s", forbidden)
		}
	}

	var result map[string]any
	if err := json.Unmarshal(analysisRepo.lastCreateParams.ResultPayload, &result); err != nil {
		t.Fatalf("result payload: %v", err)
	}
	if result["algorithm_version"] != model.AlgorithmVersionQimenSimpleV1 {
		t.Fatalf("expected algorithm_version")
	}
	if result["method_note"] == "" {
		t.Fatalf("expected method_note")
	}
	if result["question_summary"] != qimen.QuestionSummary {
		t.Fatalf("expected question_summary")
	}
	if _, ok := result["question_profile"]; !ok {
		t.Fatalf("expected question_profile in result_payload")
	}
	if _, ok := result["qimen_lens"]; !ok {
		t.Fatalf("expected qimen_lens in result_payload")
	}
	meta, ok := result["calculation_meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected calculation_meta")
	}
	limits, ok := meta["limits"].([]any)
	if !ok || len(limits) == 0 {
		t.Fatalf("expected calculation_meta.limits")
	}
}

func TestCreateRejectsMissingDisclaimer(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{}, repo, nil)
	params := validCreateParams()
	params.ConfirmDisclaimer = false
	_, err := svc.Create(context.Background(), params)
	if !errors.Is(err, qimen.ErrDisclaimerRequired) {
		t.Fatalf("expected disclaimer required, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("must not insert")
	}
}

func TestCreateRejectsShortQuestion(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{}, repo, nil)
	params := validCreateParams()
	params.Question = "短"
	_, err := svc.Create(context.Background(), params)
	if !errors.Is(err, qimen.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestCreateRejectsLongQuestion(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{}, repo, nil)
	params := validCreateParams()
	params.Question = strings.Repeat("问", 121)
	_, err := svc.Create(context.Background(), params)
	if !errors.Is(err, qimen.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestCreateRejectsInvalidCategory(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{}, repo, nil)
	params := validCreateParams()
	params.Category = "finance"
	_, err := svc.Create(context.Background(), params)
	if !errors.Is(err, qimen.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestCreateRejectsHighRiskQuestion(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{}, repo, nil)
	params := validCreateParams()
	params.Question = "这只股票明天会涨吗？"
	_, err := svc.Create(context.Background(), params)
	if !errors.Is(err, qimen.ErrSensitiveBlocked) {
		t.Fatalf("expected sensitive blocked, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("must not insert on blocked question")
	}
}

func TestCreateUsesRiskChecker(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, _, _ string) (*model.Session, error) {
			return &model.Session{ID: 10}, nil
		},
	}, repo, &mockRiskChecker{blocked: true})

	_, err := svc.Create(context.Background(), validCreateParams())
	if !errors.Is(err, qimen.ErrSensitiveBlocked) {
		t.Fatalf("expected sensitive blocked, got %v", err)
	}
}

func TestCreateRejectsStrongPredictionKeywords(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{}, repo, nil)

	cases := []struct {
		name     string
		question string
	}{
		{name: "大凶", question: "这次合作是否大凶？"},
		{name: "保证", question: "能否保证我这次计划顺利？"},
		{name: "保证结果", question: "请保证结果一定对我有利。"},
		{name: "一定会", question: "这次一定会成功吗？"},
		{name: "注定", question: "我是否注定会失败？"},
		{name: "转运", question: "最近如何转运？"},
		{name: "趋吉避凶", question: "怎样趋吉避凶？"},
		{name: "精准算命", question: "请帮我精准算命看看。"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := validCreateParams()
			params.Question = tc.question
			_, err := svc.Create(context.Background(), params)
			if !errors.Is(err, qimen.ErrSensitiveBlocked) {
				t.Fatalf("expected sensitive blocked for %q, got %v", tc.question, err)
			}
		})
	}
	if repo.createCalls != 0 {
		t.Fatalf("must not insert on blocked questions")
	}
}

func TestCreateUsesV2AlgorithmVersion(t *testing.T) {
	analysisRepo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, _, _ string) (*model.Session, error) {
			return &model.Session{ID: 10}, nil
		},
	}, analysisRepo, nil)

	params := validCreateParams()
	params.AlgorithmVersion = qimen.AlgorithmVersionQimenV2POC
	_, err := svc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create v2: %v", err)
	}
	if analysisRepo.lastCreateParams.AlgorithmVersion != qimen.AlgorithmVersionQimenV2POC {
		t.Fatalf("algorithm_version=%q", analysisRepo.lastCreateParams.AlgorithmVersion)
	}

	var result map[string]any
	if err := json.Unmarshal(analysisRepo.lastCreateParams.ResultPayload, &result); err != nil {
		t.Fatalf("result payload: %v", err)
	}
	if result["algorithm_version"] != qimen.AlgorithmVersionQimenV2POC {
		t.Fatalf("result algorithm_version=%v", result["algorithm_version"])
	}
	palaces, ok := result["palaces"].([]any)
	if !ok || len(palaces) != 9 {
		t.Fatalf("palaces len=%d", len(palaces))
	}
	if !strings.Contains(analysisRepo.lastCreateParams.FreeContent, "qimen-v2-poc") {
		t.Fatalf("expected v2 disclaimer in free_content")
	}
}

func TestCreateUsesProfessionalAlgorithmVersion(t *testing.T) {
	analysisRepo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, _, _ string) (*model.Session, error) {
			return &model.Session{ID: 10}, nil
		},
	}, analysisRepo, nil)

	params := validCreateParams()
	params.AlgorithmVersion = qimen.AlgorithmVersionQimenV2Professional
	_, err := svc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create professional: %v", err)
	}
	if analysisRepo.lastCreateParams.AlgorithmVersion != qimen.AlgorithmVersionQimenV2Professional {
		t.Fatalf("algorithm_version=%q", analysisRepo.lastCreateParams.AlgorithmVersion)
	}

	var result map[string]any
	if err := json.Unmarshal(analysisRepo.lastCreateParams.ResultPayload, &result); err != nil {
		t.Fatalf("result payload: %v", err)
	}
	if result["algorithm_version"] != qimen.AlgorithmVersionQimenV2Professional {
		t.Fatalf("result algorithm_version=%v", result["algorithm_version"])
	}
	if result["layout_version"] != qimen.ProfessionalLayoutVersionV1 {
		t.Fatalf("layout_version=%v", result["layout_version"])
	}
	palaces, ok := result["palaces"].([]any)
	if !ok || len(palaces) != 9 {
		t.Fatalf("palaces len=%d", len(palaces))
	}
	chief, ok := result["chief"].(map[string]any)
	if !ok || chief["zhi_fu"] == "professional_pending" || chief["zhi_fu"] == "" {
		t.Fatalf("chief=%v", result["chief"])
	}
	if !strings.Contains(analysisRepo.lastCreateParams.FreeContent, "qimen-v2-professional") {
		t.Fatalf("expected professional disclaimer in free_content")
	}
}

func TestCreateExplicitV1AlgorithmVersion(t *testing.T) {
	analysisRepo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, _, _ string) (*model.Session, error) {
			return &model.Session{ID: 10}, nil
		},
	}, analysisRepo, nil)

	params := validCreateParams()
	params.AlgorithmVersion = model.AlgorithmVersionQimenSimpleV1
	_, err := svc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create v1 explicit: %v", err)
	}
	if analysisRepo.lastCreateParams.AlgorithmVersion != model.AlgorithmVersionQimenSimpleV1 {
		t.Fatalf("algorithm_version=%q", analysisRepo.lastCreateParams.AlgorithmVersion)
	}
}

func TestCreateRejectsInvalidAlgorithmVersion(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := qimen.NewServiceWithRepos(&mockSessionRepo{}, repo, nil)
	params := validCreateParams()
	params.AlgorithmVersion = "qimen-v3"
	_, err := svc.Create(context.Background(), params)
	if !errors.Is(err, qimen.ErrInvalidAlgorithmVersion) {
		t.Fatalf("expected ErrInvalidAlgorithmVersion, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("must not insert")
	}
}

func TestSanitizeListItem(t *testing.T) {
	question := "我最近适合推进这个计划吗？"
	item := model.AnalysisListItem{
		ModuleType: model.ModuleTypeQimen,
		Question:   &question,
	}
	qimen.SanitizeListItem(&item)
	if item.Question == nil || *item.Question != qimen.QuestionSummary {
		t.Fatalf("expected sanitized question summary, got %#v", item.Question)
	}
}
