package session

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

var ErrSessionKeyEmpty = errors.New("session_key is required")

type sessionRepository interface {
	Upsert(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error)
	FindByKey(ctx context.Context, sessionKey string) (*model.Session, error)
}

type Service struct {
	repo sessionRepository
}

func NewService(repo *repository.SessionRepository) *Service {
	return &Service{repo: repo}
}

func NewServiceWithRepo(repo sessionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrGet(ctx context.Context, sessionKey, clientInfo string) (*model.Session, error) {
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		sessionKey = uuid.NewString()
	}
	return s.repo.Upsert(ctx, sessionKey, clientInfo)
}

// SessionIDByKey resolves an existing session id. Returns (0, nil) when the key is unknown.
func (s *Service) SessionIDByKey(ctx context.Context, sessionKey string) (int64, error) {
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		return 0, ErrSessionKeyEmpty
	}
	sess, err := s.repo.FindByKey(ctx, sessionKey)
	if err != nil {
		return 0, err
	}
	if sess == nil {
		return 0, nil
	}
	return sess.ID, nil
}
