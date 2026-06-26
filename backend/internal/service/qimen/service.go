package qimen

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/pkg/sessionkey"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

var (
	ErrInvalidParams      = fmt.Errorf("invalid params")
	ErrSessionKeyEmpty    = fmt.Errorf("session_key is required")
	ErrSessionKeyTooLong  = fmt.Errorf("session_key exceeds max length")
	ErrSensitiveBlocked   = fmt.Errorf("sensitive blocked")
	ErrDisclaimerRequired = fmt.Errorf("confirm_disclaimer required")
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
	risk     questionRiskChecker
}

func NewService(
	sessions *repository.SessionRepository,
	analysis *repository.AnalysisRepository,
	risk questionRiskChecker,
) *Service {
	return &Service{sessions: sessions, analysis: analysis, risk: risk}
}

func NewServiceWithRepos(
	sessions sessionRepository,
	analysis analysisRepository,
	risk questionRiskChecker,
) *Service {
	return &Service{sessions: sessions, analysis: analysis, risk: risk}
}

type CreateParams struct {
	SessionKey        string
	Question          string
	Category          string
	ConfirmDisclaimer bool
	AlgorithmVersion  string
	ClientInfo        string
}

func (s *Service) Create(ctx context.Context, input CreateParams) (*model.AnalysisRecord, error) {
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
	if !input.ConfirmDisclaimer {
		return nil, ErrDisclaimerRequired
	}

	question := strings.TrimSpace(input.Question)
	if !ValidateQuestionLength(question) {
		return nil, ErrInvalidParams
	}

	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "general"
	} else {
		category = strings.ToLower(category)
		if !ValidateCategory(category) {
			return nil, ErrInvalidParams
		}
	}

	blocked, err := checkQuestionRisk(ctx, s.risk, question)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrSensitiveBlocked
	}

	algorithmVersion, err := ResolveAlgorithmVersion(input.AlgorithmVersion)
	if err != nil {
		return nil, err
	}

	now := clock.Now()
	v1Calc, err := Calculate(question, category, now)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidParams, err)
	}

	var (
		freeContent   string
		resultPayload []byte
	)

	switch algorithmVersion {
	case AlgorithmVersionQimenV2POC:
		v2Calc, calcErr := CalculateV2(CalculateInputV2{
			Category: category,
			Now:      now,
		})
		if calcErr != nil {
			return nil, calcErr
		}
		merged := MergeV1InterpretationWithV2(v1Calc, v2Calc)
		freeContent = BuildFreeContentForVersion(merged, algorithmVersion)
		if strings.TrimSpace(freeContent) == "" {
			return nil, fmt.Errorf("%w: free_content generation failed", ErrInvalidParams)
		}
		resultPayload, err = BuildV2APIResultPayload(v1Calc, v2Calc)
		if err != nil {
			return nil, err
		}
	default:
		freeContent = BuildFreeContent(v1Calc)
		if strings.TrimSpace(freeContent) == "" {
			return nil, fmt.Errorf("%w: free_content generation failed", ErrInvalidParams)
		}
		resultPayload, err = v1Calc.ResultPayload()
		if err != nil {
			return nil, err
		}
	}

	inputPayload, err := v1Calc.InputPayload()
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.Upsert(ctx, sessionKey, input.ClientInfo)
	if err != nil {
		return nil, err
	}

	questionCopy := question
	id, err := s.analysis.CreateWithFreeContent(ctx, repository.CreateAnalysisWithFreeContentParams{
		CreateAnalysisParams: repository.CreateAnalysisParams{
			SessionID:        session.ID,
			ModuleType:       model.ModuleTypeQimen,
			AlgorithmVersion: algorithmVersion,
			Question:         &questionCopy,
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
