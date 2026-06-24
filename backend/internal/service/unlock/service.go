package unlock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/interpretation"
)

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrForbidden     = fmt.Errorf("forbidden")
	ErrNotUnlocked   = fmt.Errorf("not unlocked")
	ErrInvalidParams = fmt.Errorf("invalid params")
)

type Service struct {
	divinationRepo    *repository.DivinationRepository
	sessionRepo       *repository.SessionRepository
	unlockRepo        *repository.UnlockRepository
	interpretationSvc *interpretation.Service
	hexagramRepo      *repository.HexagramRepository
	categoryRepo      *repository.CategoryRepository
}

func NewService(
	divinationRepo *repository.DivinationRepository,
	sessionRepo *repository.SessionRepository,
	unlockRepo *repository.UnlockRepository,
	interpretationSvc *interpretation.Service,
	hexagramRepo *repository.HexagramRepository,
	categoryRepo *repository.CategoryRepository,
) *Service {
	return &Service{
		divinationRepo:    divinationRepo,
		sessionRepo:       sessionRepo,
		unlockRepo:        unlockRepo,
		interpretationSvc: interpretationSvc,
		hexagramRepo:      hexagramRepo,
		categoryRepo:      categoryRepo,
	}
}

func (s *Service) Unlock(ctx context.Context, divinationID int64, sessionKey, unlockType string) (*model.UnlockResult, error) {
	sessionKey = strings.TrimSpace(sessionKey)
	unlockType = strings.TrimSpace(unlockType)

	if sessionKey == "" || divinationID <= 0 {
		return nil, ErrInvalidParams
	}
	if err := validateUnlockType(unlockType); err != nil {
		return nil, err
	}

	divination, err := s.divinationRepo.FindByID(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if divination == nil {
		return nil, ErrNotFound
	}

	session, err := s.sessionRepo.FindByKey(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if err := validateSessionAccess(divination.SessionID, session); err != nil {
		return nil, err
	}

	if divination.UnlockStatus == model.UnlockStatusUnlocked {
		full, err := s.interpretationSvc.GetFullContent(ctx, divinationID)
		if err != nil {
			return nil, err
		}
		return &model.UnlockResult{
			DivinationID:       divinationID,
			UnlockStatus:       model.UnlockStatusUnlocked,
			MockTransactionID:  "",
			FullInterpretation: full,
		}, nil
	}

	in, err := s.buildGenerateInput(ctx, divination)
	if err != nil {
		return nil, err
	}
	if interp, err := s.interpretationSvc.GetFree(ctx, divinationID); err == nil && interp != nil {
		in.FreeContent = interp.FreeContent
	}

	report, err := s.interpretationSvc.GenerateAndSaveFull(ctx, in, divinationID)
	if err != nil {
		return nil, err
	}

	mockTxnID := fmt.Sprintf("MOCK-%s-%d", clock.Now().Format("20060102"), clock.Now().UnixNano()%100000)

	if err := s.unlockRepo.Create(ctx, repository.CreateUnlockParams{
		DivinationID:        divinationID,
		SessionID:           session.ID,
		UnlockType:          unlockType,
		MockTransactionID:   mockTxnID,
	}); err != nil {
		return nil, err
	}

	if err := s.divinationRepo.UpdateUnlockStatus(ctx, divinationID, model.UnlockStatusUnlocked); err != nil {
		return nil, err
	}

	return &model.UnlockResult{
		DivinationID:       divinationID,
		UnlockStatus:       model.UnlockStatusUnlocked,
		MockTransactionID:  mockTxnID,
		FullInterpretation: report.Report,
	}, nil
}

func (s *Service) GetFullInterpretation(ctx context.Context, divinationID int64, sessionKey string) (*model.FullReport, error) {
	result, err := s.GetFullInterpretationWithMeta(ctx, divinationID, sessionKey)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ErrNotUnlocked
	}
	return result.Report, nil
}

func (s *Service) GetFullInterpretationWithMeta(ctx context.Context, divinationID int64, sessionKey string) (*interpretation.FullResult, error) {
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" || divinationID <= 0 {
		return nil, ErrInvalidParams
	}

	divination, err := s.divinationRepo.FindByID(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if divination == nil {
		return nil, ErrNotFound
	}

	session, err := s.sessionRepo.FindByKey(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if err := validateSessionAccess(divination.SessionID, session); err != nil {
		return nil, err
	}

	if err := validateUnlocked(divination.UnlockStatus); err != nil {
		return nil, err
	}

	result, err := s.interpretationSvc.GetFullWithMeta(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ErrNotUnlocked
	}
	return result, nil
}

func validateSessionAccess(divinationSessionID int64, session *model.Session) error {
	if session == nil || session.ID != divinationSessionID {
		return ErrForbidden
	}
	return nil
}

func validateUnlockType(unlockType string) error {
	switch unlockType {
	case model.UnlockTypeMockAd,
		model.UnlockTypeMockButton,
		model.UnlockTypeRewardedVideoMock:
		return nil
	default:
		return ErrInvalidParams
	}
}

func validateUnlocked(unlockStatus int) error {
	if unlockStatus != model.UnlockStatusUnlocked {
		return ErrNotUnlocked
	}
	return nil
}

func (s *Service) buildGenerateInput(ctx context.Context, d *model.Divination) (interpretation.GenerateInput, error) {
	category, err := s.categoryRepo.FindActiveByID(ctx, d.CategoryID)
	if err != nil {
		return interpretation.GenerateInput{}, err
	}
	primary, err := s.hexagramRepo.FindByID(ctx, d.PrimaryHexagramID)
	if err != nil {
		return interpretation.GenerateInput{}, err
	}
	changed, err := s.hexagramRepo.FindByID(ctx, d.ChangedHexagramID)
	if err != nil {
		return interpretation.GenerateInput{}, err
	}

	var lines []model.Line
	var movingLines []int
	if d.LineSnapshot != "" {
		if err := json.Unmarshal([]byte(d.LineSnapshot), &lines); err != nil {
			return interpretation.GenerateInput{}, err
		}
	}
	if d.MovingLines != "" {
		if err := json.Unmarshal([]byte(d.MovingLines), &movingLines); err != nil {
			return interpretation.GenerateInput{}, err
		}
	}

	categoryName := ""
	if category != nil {
		categoryName = category.Name
	}

	return interpretation.GenerateInput{
		DivinationID:    d.ID,
		Question:        d.Question,
		CategoryName:    categoryName,
		PrimaryHexagram: primary,
		ChangedHexagram: changed,
		MovingLines:     movingLines,
		Lines:           lines,
		LineSnapshot:    d.LineSnapshot,
	}, nil
}
