package dailyfortune

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/divination"
	"github.com/wangxintong/yijing/backend/internal/service/interpretation"
	"github.com/wangxintong/yijing/backend/internal/service/session"
)

type Service struct {
	dailyFortuneRepo  *repository.DailyFortuneRepository
	sessionSvc        *session.Service
	divinationSvc     *divination.Service
	interpretationSvc *interpretation.Service
}

func NewService(
	dailyFortuneRepo *repository.DailyFortuneRepository,
	sessionSvc *session.Service,
	divinationSvc *divination.Service,
	interpretationSvc *interpretation.Service,
) *Service {
	return &Service{
		dailyFortuneRepo:  dailyFortuneRepo,
		sessionSvc:        sessionSvc,
		divinationSvc:     divinationSvc,
		interpretationSvc: interpretationSvc,
	}
}

type TodayInput struct {
	SessionKey string
	LocalDate  string
	ClientInfo string
}

type TodayResult struct {
	FortuneDate string
	IsExisting  bool
	Divination  *model.Divination
}

var (
	ErrSessionKeyEmpty = fmt.Errorf("session key empty")
	ErrInvalidDate     = fmt.Errorf("invalid local date")
)

func (s *Service) GetOrCreateToday(ctx context.Context, in TodayInput) (*TodayResult, error) {
	sessionKey := strings.TrimSpace(in.SessionKey)
	if sessionKey == "" {
		return nil, ErrSessionKeyEmpty
	}
	fortuneDate, err := parseLocalDate(in.LocalDate)
	if err != nil {
		return nil, ErrInvalidDate
	}

	sess, err := s.sessionSvc.CreateOrGet(ctx, sessionKey, in.ClientInfo)
	if err != nil {
		return nil, err
	}

	existing, err := s.dailyFortuneRepo.FindBySessionAndDate(ctx, sess.ID, fortuneDate)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		detail, err := s.loadDivinationWithFree(ctx, existing.DivinationID)
		if err != nil {
			return nil, err
		}
		return &TodayResult{
			FortuneDate: fortuneDate,
			IsExisting:  true,
			Divination:  detail,
		}, nil
	}

	created, err := s.divinationSvc.Create(ctx, divination.CreateInput{
		SessionKey:        sessionKey,
		CategoryID:        model.DailyFortuneCategoryID,
		Question:          model.DailyFortuneQuestion,
		ConfirmDisclaimer: true,
		ClientInfo:        in.ClientInfo,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.dailyFortuneRepo.Create(ctx, sess.ID, fortuneDate, created.ID)
	if err != nil {
		if errors.Is(err, repository.ErrDailyFortuneDuplicate) {
			existing, findErr := s.dailyFortuneRepo.FindBySessionAndDate(ctx, sess.ID, fortuneDate)
			if findErr != nil {
				return nil, findErr
			}
			if existing == nil {
				return nil, fmt.Errorf("daily fortune conflict but record not found")
			}
			detail, loadErr := s.loadDivinationWithFree(ctx, existing.DivinationID)
			if loadErr != nil {
				return nil, loadErr
			}
			return &TodayResult{
				FortuneDate: fortuneDate,
				IsExisting:  true,
				Divination:  detail,
			}, nil
		}
		return nil, err
	}

	return &TodayResult{
		FortuneDate: fortuneDate,
		IsExisting:  false,
		Divination:  created,
	}, nil
}

func (s *Service) loadDivinationWithFree(ctx context.Context, divinationID int64) (*model.Divination, error) {
	detail, err := s.divinationSvc.GetByID(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, fmt.Errorf("divination not found: %d", divinationID)
	}
	record, err := s.interpretationSvc.GetFree(ctx, divinationID)
	if err != nil {
		return nil, err
	}
	if record != nil {
		detail.FreeInterpretation = record.FreeContent
	}
	return detail, nil
}

func parseLocalDate(localDate string) (string, error) {
	localDate = strings.TrimSpace(localDate)
	t, err := time.Parse("2006-01-02", localDate)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}
