package divination

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/interpretation"
	"github.com/wangxintong/yijing/backend/internal/service/sensitive"
)

type Service struct {
	divinationRepo    *repository.DivinationRepository
	hexagramRepo      *repository.HexagramRepository
	categoryRepo      *repository.CategoryRepository
	sessionRepo       *repository.SessionRepository
	sensitiveSvc      *sensitive.Service
	interpretationSvc *interpretation.Service
}

func NewService(
	divinationRepo *repository.DivinationRepository,
	hexagramRepo *repository.HexagramRepository,
	categoryRepo *repository.CategoryRepository,
	sessionRepo *repository.SessionRepository,
	sensitiveSvc *sensitive.Service,
	interpretationSvc *interpretation.Service,
) *Service {
	return &Service{
		divinationRepo:    divinationRepo,
		hexagramRepo:      hexagramRepo,
		categoryRepo:      categoryRepo,
		sessionRepo:       sessionRepo,
		sensitiveSvc:      sensitiveSvc,
		interpretationSvc: interpretationSvc,
	}
}

type CreateInput struct {
	SessionKey        string
	CategoryID        int64
	Question          string
	ConfirmDisclaimer bool
	ClientInfo        string
}

var (
	ErrInvalidParams     = fmt.Errorf("invalid params")
	ErrSensitiveBlocked  = fmt.Errorf("sensitive blocked")
	ErrCategoryNotFound  = fmt.Errorf("category not found")
	ErrSessionKeyEmpty   = fmt.Errorf("session key empty")
)

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.Divination, error) {
	question := strings.TrimSpace(in.Question)
	sessionKey := strings.TrimSpace(in.SessionKey)

	if sessionKey == "" {
		return nil, ErrSessionKeyEmpty
	}
	if !in.ConfirmDisclaimer {
		return nil, ErrInvalidParams
	}
	if !sensitive.ValidateQuestionLength(question) {
		return nil, ErrInvalidParams
	}

	blocked, err := s.sensitiveSvc.CheckQuestion(ctx, question)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrSensitiveBlocked
	}

	category, err := s.categoryRepo.FindActiveByID(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}

	session, err := s.sessionRepo.Upsert(ctx, sessionKey, in.ClientInfo)
	if err != nil {
		return nil, err
	}

	seed := buildSeed(sessionKey, question)
	lines := tossLines(seed)
	primaryBinary := linesToBinary(lines)
	changedBinary := changedBinaryFromLines(lines)

	primaryHexagram, err := s.hexagramRepo.FindByBinaryCode(ctx, primaryBinary)
	if err != nil {
		return nil, fmt.Errorf("lookup primary hexagram: %w", err)
	}
	changedHexagram, err := s.hexagramRepo.FindByBinaryCode(ctx, changedBinary)
	if err != nil {
		return nil, fmt.Errorf("lookup changed hexagram: %w", err)
	}

	movingLines := collectMovingLines(lines)
	movingLinesJSON, err := json.Marshal(movingLines)
	if err != nil {
		return nil, err
	}
	lineSnapshotJSON, err := json.Marshal(lines)
	if err != nil {
		return nil, err
	}

	divinationID, err := s.divinationRepo.Create(ctx, repository.CreateDivinationParams{
		SessionID:         session.ID,
		CategoryID:        category.ID,
		Question:          question,
		Method:            model.MethodCoinThree,
		PrimaryHexagramID: primaryHexagram.ID,
		ChangedHexagramID: changedHexagram.ID,
		MovingLinesJSON:   string(movingLinesJSON),
		LineSnapshotJSON:  string(lineSnapshotJSON),
		Seed:              seed,
		Lines:             lines,
	})
	if err != nil {
		return nil, err
	}

	freeContent, err := s.interpretationSvc.CreateFree(ctx, interpretation.GenerateInput{
		Question:        question,
		CategoryName:    category.Name,
		PrimaryHexagram: primaryHexagram,
		ChangedHexagram: changedHexagram,
		MovingLines:     movingLines,
		Lines:           lines,
	}, divinationID)
	if err != nil {
		return nil, fmt.Errorf("create free interpretation: %w", err)
	}

	detail, err := s.buildDetail(ctx, &model.Divination{
		ID:                divinationID,
		SessionID:         session.ID,
		CategoryID:        category.ID,
		Question:          question,
		Method:            model.MethodCoinThree,
		PrimaryHexagramID: primaryHexagram.ID,
		ChangedHexagramID: changedHexagram.ID,
		MovingLines:       string(movingLinesJSON),
		LineSnapshot:      string(lineSnapshotJSON),
		Seed:              seed,
		UnlockStatus:      model.UnlockStatusLocked,
		CreatedAt:         clock.Now(),
	})
	if err != nil {
		return nil, err
	}
	detail.FreeInterpretation = freeContent
	return detail, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*model.Divination, error) {
	record, err := s.divinationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return s.buildDetail(ctx, record)
}

func (s *Service) ListHistory(ctx context.Context, sessionKey string, page, pageSize int) (*model.PaginatedDivinations, error) {
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		return nil, ErrSessionKeyEmpty
	}

	session, err := s.sessionRepo.FindByKey(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return &model.PaginatedDivinations{
			Items:    []model.DivinationListItem{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
		}, nil
	}

	records, total, err := s.divinationRepo.ListBySession(ctx, session.ID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]model.DivinationListItem, 0, len(records))
	for i := range records {
		item, err := s.buildListItem(ctx, &records[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return &model.PaginatedDivinations{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *Service) buildDetail(ctx context.Context, record *model.Divination) (*model.Divination, error) {
	category, err := s.categoryRepo.FindActiveByID(ctx, record.CategoryID)
	if err != nil {
		return nil, err
	}
	primary, err := s.hexagramRepo.FindByID(ctx, record.PrimaryHexagramID)
	if err != nil {
		return nil, err
	}
	changed, err := s.hexagramRepo.FindByID(ctx, record.ChangedHexagramID)
	if err != nil {
		return nil, err
	}

	lines, movingLines, err := parseLineSnapshot(record.LineSnapshot, record.MovingLines)
	if err != nil {
		return nil, err
	}

	record.Category = &model.Category{ID: category.ID, Name: category.Name}
	record.PrimaryHexagram = trimHexagram(primary)
	record.ChangedHexagram = trimHexagram(changed)
	record.Lines = lines
	record.MovingLinesArray = movingLines
	return record, nil
}

func (s *Service) buildListItem(ctx context.Context, record *model.Divination) (*model.DivinationListItem, error) {
	category, err := s.categoryRepo.FindActiveByID(ctx, record.CategoryID)
	if err != nil {
		return nil, err
	}
	primary, err := s.hexagramRepo.FindByID(ctx, record.PrimaryHexagramID)
	if err != nil {
		return nil, err
	}
	changed, err := s.hexagramRepo.FindByID(ctx, record.ChangedHexagramID)
	if err != nil {
		return nil, err
	}
	_, movingLines, err := parseLineSnapshot(record.LineSnapshot, record.MovingLines)
	if err != nil {
		return nil, err
	}

	return &model.DivinationListItem{
		ID:              record.ID,
		Question:        record.Question,
		Category:        &model.Category{ID: category.ID, Name: category.Name},
		PrimaryHexagram: trimHexagram(primary),
		ChangedHexagram: trimHexagram(changed),
		MovingLines:     movingLines,
		UnlockStatus:    record.UnlockStatus,
		CreatedAt:       clock.FormatRFC3339(record.CreatedAt),
	}, nil
}

func trimHexagram(h *model.Hexagram) *model.Hexagram {
	if h == nil {
		return nil
	}
	return &model.Hexagram{
		ID:         h.ID,
		Number:     h.Number,
		Name:       h.Name,
		FullName:   h.FullName,
		BinaryCode: h.BinaryCode,
	}
}

func parseLineSnapshot(snapshot, movingLinesJSON string) ([]model.Line, []int, error) {
	var lines []model.Line
	if snapshot != "" {
		if err := json.Unmarshal([]byte(snapshot), &lines); err != nil {
			return nil, nil, err
		}
	}
	var movingLines []int
	if movingLinesJSON != "" {
		if err := json.Unmarshal([]byte(movingLinesJSON), &movingLines); err != nil {
			return nil, nil, err
		}
	}
	if movingLines == nil {
		movingLines = []int{}
	}
	return lines, movingLines, nil
}

func buildSeed(sessionKey, question string) string {
	sum := sha256.Sum256([]byte(sessionKey + "|" + question + "|" + clock.Now().Format(time.RFC3339Nano)))
	return hex.EncodeToString(sum[:16])
}

func tossLines(seed string) []model.Line {
	r := rand.New(rand.NewSource(hashSeed(seed)))
	lines := make([]model.Line, 0, 6)
	for pos := 1; pos <= 6; pos++ {
		heads := 0
		for i := 0; i < 3; i++ {
			if r.Intn(2) == 1 {
				heads++
			}
		}
		line := model.Line{Position: pos}
		switch heads {
		case 0:
			line.Value, line.IsYang, line.IsMoving = 6, 0, 1
		case 1:
			line.Value, line.IsYang, line.IsMoving = 7, 1, 0
		case 2:
			line.Value, line.IsYang, line.IsMoving = 8, 0, 0
		case 3:
			line.Value, line.IsYang, line.IsMoving = 9, 1, 1
		}
		lines = append(lines, line)
	}
	return lines
}

func hashSeed(seed string) int64 {
	var h int64
	for i := 0; i < len(seed); i++ {
		h = h*31 + int64(seed[i])
	}
	if h < 0 {
		h = -h
	}
	return h
}

func linesToBinary(lines []model.Line) string {
	b := make([]byte, 6)
	for i, line := range lines {
		if line.IsYang == 1 {
			b[i] = '1'
		} else {
			b[i] = '0'
		}
	}
	return string(b)
}

func changedBinaryFromLines(lines []model.Line) string {
	b := make([]byte, 6)
	for i, line := range lines {
		yang := line.IsYang == 1
		if line.IsMoving == 1 {
			yang = !yang
		}
		if yang {
			b[i] = '1'
		} else {
			b[i] = '0'
		}
	}
	return string(b)
}

func collectMovingLines(lines []model.Line) []int {
	var moving []int
	for _, line := range lines {
		if line.IsMoving == 1 {
			moving = append(moving, line.Position)
		}
	}
	if moving == nil {
		moving = []int{}
	}
	return moving
}
