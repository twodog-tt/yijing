package category

import (
	"context"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/repository"
)

type Service struct {
	repo *repository.CategoryRepository
}

func NewService(repo *repository.CategoryRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListActive(ctx context.Context) ([]model.Category, error) {
	items, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]model.Category, 0, len(items))
	for _, item := range items {
		result = append(result, model.Category{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return result, nil
}
