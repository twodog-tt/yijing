package ai

import (
	"context"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/config"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

type Router struct {
	cfg      *config.Config
	mock     Provider
	deepseek Provider
	logRepo  *repository.AILogRepository
}

func NewRouter(cfg *config.Config, logRepo *repository.AILogRepository) *Router {
	mock := NewMockProvider()
	return &Router{
		cfg:      cfg,
		mock:     mock,
		deepseek: NewDeepSeekProvider(cfg, mock),
		logRepo:  logRepo,
	}
}

func (r *Router) GenerateFullInterpretation(ctx context.Context, input GenerateInput) (*GenerateOutput, error) {
	start := time.Now()
	provider := strings.ToLower(strings.TrimSpace(r.cfg.AIProvider))

	var out *GenerateOutput
	var err error
	switch provider {
	case model.AIProviderDeepSeek:
		out, err = r.deepseek.GenerateFullInterpretation(ctx, input)
	default:
		out, err = r.mock.GenerateFullInterpretation(ctx, input)
	}

	if r.logRepo != nil && input.DivinationID > 0 {
		r.recordLog(ctx, input, out, err, start)
	}
	return out, err
}

func (r *Router) recordLog(ctx context.Context, input GenerateInput, out *GenerateOutput, callErr error, start time.Time) {
	entry := repository.CreateAILogInput{
		DivinationID:    input.DivinationID,
		QuestionSummary: input.Question,
		DurationMs:      int(time.Since(start).Milliseconds()),
	}

	if callErr != nil {
		entry.AIProvider = r.ConfiguredProvider()
		entry.ModelName = r.configuredModelName()
		entry.Status = model.AILogStatusFailed
		entry.ErrorMessage = callErr.Error()
		_ = r.logRepo.Create(ctx, entry)
		return
	}

	if out == nil {
		return
	}

	entry.AIProvider = out.Provider
	entry.ModelName = out.ModelName
	entry.FallbackUsed = out.FallbackUsed
	entry.ErrorMessage = out.ErrorMessage
	if out.DurationMs > 0 {
		entry.DurationMs = int(out.DurationMs)
	}

	switch out.Provider {
	case model.AIProviderMockFallback:
		entry.Status = model.AILogStatusFallbackSuccess
	default:
		entry.Status = model.AILogStatusSuccess
	}
	_ = r.logRepo.Create(ctx, entry)
}

func (r *Router) MockProvider() Provider {
	return r.mock
}

func (r *Router) DeepSeekProvider() Provider {
	return r.deepseek
}

func (r *Router) ConfiguredProvider() string {
	p := strings.ToLower(strings.TrimSpace(r.cfg.AIProvider))
	if p == model.AIProviderDeepSeek {
		return model.AIProviderDeepSeek
	}
	return model.AIProviderMock
}

func (r *Router) configuredModelName() string {
	if r.ConfiguredProvider() == model.AIProviderDeepSeek {
		return r.cfg.DeepSeekModel
	}
	return "mock"
}

func (r *Router) Config() *config.Config {
	return r.cfg
}
