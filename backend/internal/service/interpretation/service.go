package interpretation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/ai"
)

type Service struct {
	repo     *repository.InterpretationRepository
	aiRouter *ai.Router
}

func NewService(repo *repository.InterpretationRepository, aiRouter *ai.Router) *Service {
	return &Service{repo: repo, aiRouter: aiRouter}
}

func (s *Service) CreateFree(ctx context.Context, in GenerateInput, divinationID int64) (string, error) {
	freeContent := BuildFreeContent(in)
	now := clock.Now()
	if err := s.repo.CreateFree(ctx, divinationID, freeContent, now); err != nil {
		return "", err
	}
	return freeContent, nil
}

func (s *Service) GetFree(ctx context.Context, divinationID int64) (*model.Interpretation, error) {
	return s.repo.FindByDivinationID(ctx, divinationID)
}

type FullResult struct {
	Report   *model.FullReport
	Provider string
}

func (s *Service) GenerateAndSaveFull(ctx context.Context, in GenerateInput, divinationID int64) (*FullResult, error) {
	existing, err := s.repo.FindByDivinationID(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if hasExistingFullContent(existing) {
		report, err := parseFullReport(*existing.FullContent)
		if err != nil {
			return nil, err
		}
		return &FullResult{Report: report, Provider: existing.AIProvider}, nil
	}

	if in.FreeContent == "" && existing != nil {
		in.FreeContent = existing.FreeContent
	}
	in.DivinationID = divinationID

	output, err := s.aiRouter.GenerateFullInterpretation(ctx, toAIInput(in))
	if err != nil {
		return nil, err
	}

	now := clock.Now()
	if err := s.repo.UpdateFull(ctx, divinationID, output.Content, output.Provider, now); err != nil {
		return nil, err
	}

	report, err := parseFullReport(output.Content)
	if err != nil {
		return nil, err
	}
	return &FullResult{Report: report, Provider: output.Provider}, nil
}

func (s *Service) GetFullContent(ctx context.Context, divinationID int64) (*model.FullReport, error) {
	result, err := s.GetFullWithMeta(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.Report, nil
}

func (s *Service) GetFullWithMeta(ctx context.Context, divinationID int64) (*FullResult, error) {
	record, err := s.repo.FindByDivinationID(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if record == nil || record.FullContent == nil || *record.FullContent == "" {
		return nil, nil
	}
	report, err := parseFullReport(*record.FullContent)
	if err != nil {
		return nil, err
	}
	return &FullResult{Report: report, Provider: record.AIProvider}, nil
}

func parseFullReport(content string) (*model.FullReport, error) {
	var report model.FullReport
	if err := json.Unmarshal([]byte(content), &report); err != nil {
		return nil, fmt.Errorf("parse full report: %w", err)
	}
	return &report, nil
}

func toAIInput(in GenerateInput) ai.GenerateInput {
	primary := ai.HexagramInfo{}
	changed := ai.HexagramInfo{}
	if in.PrimaryHexagram != nil {
		primary = ai.HexagramInfo{
			Name: in.PrimaryHexagram.Name, FullName: in.PrimaryHexagram.FullName,
			Summary: in.PrimaryHexagram.Summary, BinaryCode: in.PrimaryHexagram.BinaryCode,
		}
	}
	if in.ChangedHexagram != nil {
		changed = ai.HexagramInfo{
			Name: in.ChangedHexagram.Name, FullName: in.ChangedHexagram.FullName,
			Summary: in.ChangedHexagram.Summary, BinaryCode: in.ChangedHexagram.BinaryCode,
		}
	}
	return ai.GenerateInput{
		DivinationID:    in.DivinationID,
		Question:        in.Question,
		CategoryName:    in.CategoryName,
		PrimaryHexagram: primary,
		ChangedHexagram: changed,
		MovingLines:     in.MovingLines,
		LineSnapshot:    in.LineSnapshot,
		FreeContent:     in.FreeContent,
	}
}

func hasExistingFullContent(record *model.Interpretation) bool {
	return record != nil && record.FullContent != nil && strings.TrimSpace(*record.FullContent) != ""
}
