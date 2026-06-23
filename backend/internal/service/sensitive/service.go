package sensitive

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/wangxintong/yijing/backend/internal/repository"
)

const BlockMessage = "这个问题不适合用卦象方式解读。你可以换成更偏向自我反思、情绪整理或行动选择的问题。"

type Service struct {
	repo *repository.SensitiveRepository
}

func NewService(repo *repository.SensitiveRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CheckQuestion(ctx context.Context, question string) (bool, error) {
	keywords, err := s.repo.ListActiveKeywords(ctx)
	if err != nil {
		return false, err
	}

	normalized := strings.ToLower(strings.TrimSpace(question))
	for _, kw := range keywords {
		if kw == "" {
			continue
		}
		if strings.Contains(normalized, strings.ToLower(kw)) {
			return true, nil
		}
	}
	return false, nil
}

func ValidateQuestionLength(question string) bool {
	length := utf8.RuneCountInString(strings.TrimSpace(question))
	return length >= 5 && length <= 200
}
