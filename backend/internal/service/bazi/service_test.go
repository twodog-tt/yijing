package bazi_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/bazi"
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
	return 1001, nil
}

func (m *mockAnalysisRepo) FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error) {
	if m.findOwnedByIDFn != nil {
		return m.findOwnedByIDFn(ctx, id, sessionID)
	}
	free := m.lastCreateParams.FreeContent
	return &model.AnalysisRecord{
		ID:               id,
		SessionID:        sessionID,
		FreeContent:      &free,
		GenerationStatus: model.AnalysisGenerationStatusFreeDone,
	}, nil
}

func TestCreateSuccessSingleInsertWithFreeContent(t *testing.T) {
	analysisRepo := &mockAnalysisRepo{}
	svc := bazi.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, sessionKey, _ string) (*model.Session, error) {
			return &model.Session{ID: 10, SessionKey: sessionKey}, nil
		},
	}, analysisRepo)

	record, err := svc.Create(context.Background(), bazi.CreateInput{
		SessionKey:       "sess-1",
		BirthDate:        "1995-01-01",
		BirthHourBranch:  "zi",
		BirthHourUnknown: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if record == nil || record.ID != 1001 {
		t.Fatalf("unexpected record: %#v", record)
	}
	if analysisRepo.createCalls != 1 {
		t.Fatalf("expected single insert, got %d calls", analysisRepo.createCalls)
	}
	if strings.TrimSpace(analysisRepo.lastCreateParams.FreeContent) == "" {
		t.Fatalf("expected free_content in single insert")
	}

	var input map[string]any
	if err := json.Unmarshal(analysisRepo.lastCreateParams.InputPayload, &input); err != nil {
		t.Fatalf("input payload: %v", err)
	}
	if input["birth_date"] != "1995-01-01" || input["birth_hour_branch"] != "zi" {
		t.Fatalf("unexpected input payload: %#v", input)
	}
	if _, ok := input["gender"]; ok {
		t.Fatalf("must not store gender")
	}

	var result map[string]any
	if err := json.Unmarshal(analysisRepo.lastCreateParams.ResultPayload, &result); err != nil {
		t.Fatalf("result payload: %v", err)
	}
	if result["algorithm_version"] != model.AlgorithmVersionBaziSimpleV1 {
		t.Fatalf("expected algorithm_version, got %#v", result["algorithm_version"])
	}
	if result["method_note"] == "" {
		t.Fatalf("expected method_note")
	}
	if _, ok := result["birth"]; ok {
		t.Fatalf("result_payload must not contain birth")
	}
	pillars := result["pillars"].(map[string]any)
	if pillars["hour"] == "" {
		t.Fatalf("expected hour pillar for known hour")
	}
}

func TestCreateSuccessUnknownHourOmitsHourKey(t *testing.T) {
	analysisRepo := &mockAnalysisRepo{}
	svc := bazi.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, _, _ string) (*model.Session, error) {
			return &model.Session{ID: 10}, nil
		},
	}, analysisRepo)

	_, err := svc.Create(context.Background(), bazi.CreateInput{
		SessionKey:       "sess-1",
		BirthDate:        "1995-01-01",
		BirthHourUnknown: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(analysisRepo.lastCreateParams.ResultPayload, &result); err != nil {
		t.Fatalf("result payload: %v", err)
	}
	pillars := result["pillars"].(map[string]any)
	if _, ok := pillars["hour"]; ok {
		t.Fatalf("result pillars must not contain hour key")
	}
}

func TestCreateRejectsInvalidDate(t *testing.T) {
	repo := &mockAnalysisRepo{}
	svc := bazi.NewServiceWithRepos(&mockSessionRepo{}, repo)
	_, err := svc.Create(context.Background(), bazi.CreateInput{
		SessionKey: "sess-1",
		BirthDate:  "1995-02-30",
	})
	if !errors.Is(err, bazi.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("must not insert on invalid date")
	}
}

func TestCreateRejectsInvalidHourBranch(t *testing.T) {
	svc := bazi.NewServiceWithRepos(&mockSessionRepo{}, &mockAnalysisRepo{})
	_, err := svc.Create(context.Background(), bazi.CreateInput{
		SessionKey:       "sess-1",
		BirthDate:        "1995-01-01",
		BirthHourBranch:  "invalid",
		BirthHourUnknown: false,
	})
	if !errors.Is(err, bazi.ErrInvalidParams) {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestCreateDoesNotPassRawClientJSONToRepository(t *testing.T) {
	analysisRepo := &mockAnalysisRepo{}
	svc := bazi.NewServiceWithRepos(&mockSessionRepo{
		upsertFn: func(_ context.Context, _, _ string) (*model.Session, error) {
			return &model.Session{ID: 10}, nil
		},
	}, analysisRepo)

	_, err := svc.Create(context.Background(), bazi.CreateInput{
		SessionKey:       "sess-1",
		BirthDate:        "1995-01-01",
		BirthHourBranch:  "wu",
		BirthHourUnknown: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	raw := string(analysisRepo.lastCreateParams.InputPayload)
	if strings.Contains(raw, "gender") || strings.Contains(raw, "phone") {
		t.Fatalf("unexpected sensitive fields in input payload: %s", raw)
	}
	if analysisRepo.lastCreateParams.ModuleType != model.ModuleTypeBazi {
		t.Fatalf("expected bazi module type")
	}
}

func TestCreateRejectsTooLongSessionKey(t *testing.T) {
	svc := bazi.NewServiceWithRepos(&mockSessionRepo{}, &mockAnalysisRepo{})
	_, err := svc.Create(context.Background(), bazi.CreateInput{
		SessionKey: strings.Repeat("a", 65),
		BirthDate:  "1995-01-01",
	})
	if !errors.Is(err, bazi.ErrSessionKeyTooLong) {
		t.Fatalf("expected session key too long, got %v", err)
	}
}

func TestCalculateGeneratesFreeContent(t *testing.T) {
	calc, err := bazi.Calculate("1995-01-01", "zi", false)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	content := bazi.BuildFreeContent(calc)
	for _, section := range []string{"【一句话总结】", "【五行倾向学习】", "【性格/行动倾向】", "【自我反思与行动建议】", "【免责声明】"} {
		if !strings.Contains(content, section) {
			t.Fatalf("missing section %s in free content", section)
		}
	}
	if strings.Contains(content, "当前示意四柱") {
		t.Fatalf("must use 简化干支示意 wording")
	}
	if strings.Contains(content, "今日/近期反思建议") {
		t.Fatalf("must use 自我反思与行动建议 wording")
	}
	for _, banned := range []string{"精准算命", "一生命运", "婚姻财运预测", "保证发财", "疾病寿命", "改运化解"} {
		if strings.Contains(content, banned) {
			t.Fatalf("free content must not contain %q", banned)
		}
	}
}

func TestCalculateUnknownHourFreeContentMentionsMissingHour(t *testing.T) {
	calc, err := bazi.Calculate("1995-01-01", "", true)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	content := bazi.BuildFreeContent(calc)
	if !strings.Contains(content, "不生成时柱") {
		t.Fatalf("expected unknown hour note in free content")
	}
}

func TestCalculateIncludesMethodNote(t *testing.T) {
	calc, err := bazi.Calculate("1995-01-01", "zi", false)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["method_note"] == "" {
		t.Fatalf("expected method_note")
	}
}
