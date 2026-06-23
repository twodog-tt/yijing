package ai

import (
	"context"
	"encoding/json"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/jsonutil"
	"github.com/wangxintong/yijing/backend/internal/pkg/report"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return model.AIProviderMock
}

func (p *MockProvider) GenerateFullInterpretation(ctx context.Context, input GenerateInput) (*GenerateOutput, error) {
	_ = ctx
	full := report.BuildFull(report.FullInput{
		Question:       input.Question,
		CategoryName:   input.CategoryName,
		PrimaryName:    input.PrimaryHexagram.FullName,
		PrimarySummary: input.PrimaryHexagram.Summary,
		ChangedName:    input.ChangedHexagram.FullName,
		ChangedSummary: input.ChangedHexagram.Summary,
		MovingLines:    input.MovingLines,
	})
	raw, err := json.Marshal(full)
	if err != nil {
		return nil, err
	}
	content, err := jsonutil.EnsureRequiredFields(string(raw))
	if err != nil {
		return nil, err
	}
	return &GenerateOutput{
		Provider:    model.AIProviderMock,
		Content:     content,
		RawResponse: content,
		ModelName:   "mock",
	}, nil
}
