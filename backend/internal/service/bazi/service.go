package bazi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/sessionkey"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

var (
	ErrInvalidParams     = fmt.Errorf("invalid params")
	ErrSessionKeyEmpty   = fmt.Errorf("session_key is required")
	ErrSessionKeyTooLong = fmt.Errorf("session_key exceeds max length")
)

type sessionRepository interface {
	Upsert(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error)
}

type analysisRepository interface {
	CreateWithFreeContent(ctx context.Context, p repository.CreateAnalysisWithFreeContentParams) (int64, error)
	FindOwnedByID(ctx context.Context, id, sessionID int64) (*model.AnalysisRecord, error)
}

type Service struct {
	sessions sessionRepository
	analysis analysisRepository
}

func NewService(sessions *repository.SessionRepository, analysis *repository.AnalysisRepository) *Service {
	return &Service{sessions: sessions, analysis: analysis}
}

func NewServiceWithRepos(sessions sessionRepository, analysis analysisRepository) *Service {
	return &Service{sessions: sessions, analysis: analysis}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*model.AnalysisRecord, error) {
	sessionKey := strings.TrimSpace(input.SessionKey)
	if sessionKey == "" {
		return nil, ErrSessionKeyEmpty
	}
	if err := sessionkey.ValidateLength(sessionKey); err != nil {
		if errors.Is(err, sessionkey.ErrTooLong) {
			return nil, ErrSessionKeyTooLong
		}
		return nil, err
	}

	calc, err := Calculate(input.BirthDate, input.BirthHourBranch, input.BirthHourUnknown)
	if err != nil {
		return nil, err
	}

	freeContent := BuildFreeContent(calc)
	if strings.TrimSpace(freeContent) == "" {
		return nil, fmt.Errorf("%w: free_content generation failed", ErrInvalidParams)
	}

	inputPayload, err := calc.InputPayload()
	if err != nil {
		return nil, err
	}
	resultPayload, err := calc.ResultPayload()
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.Upsert(ctx, sessionKey, input.ClientInfo)
	if err != nil {
		return nil, err
	}

	id, err := s.analysis.CreateWithFreeContent(ctx, repository.CreateAnalysisWithFreeContentParams{
		CreateAnalysisParams: repository.CreateAnalysisParams{
			SessionID:        session.ID,
			ModuleType:       model.ModuleTypeBazi,
			AlgorithmVersion: model.AlgorithmVersionBaziSimpleV1,
			InputPayload:     inputPayload,
			ResultPayload:    resultPayload,
		},
		FreeContent: freeContent,
	})
	if err != nil {
		if errors.Is(err, repository.ErrInvalidAnalysisParams) {
			return nil, ErrInvalidParams
		}
		return nil, err
	}

	record, err := s.analysis.FindOwnedByID(ctx, id, session.ID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("analysis record not found after create")
	}
	return record, nil
}

func ValidateHourBranch(branch string) bool {
	_, ok := validHourBranches[strings.TrimSpace(strings.ToLower(branch))]
	return ok
}
