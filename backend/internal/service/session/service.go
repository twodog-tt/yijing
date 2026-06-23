package session

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

type Service struct {
	repo *repository.SessionRepository
}

func NewService(repo *repository.SessionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrGet(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error) {
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		sessionKey = uuid.NewString()
	}
	return s.repo.Upsert(ctx, sessionKey, clientInfo)
}
