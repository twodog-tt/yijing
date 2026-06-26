package model

import "time"

const (
	SessionStatusActive    = 1
	CategoryStatusActive   = 1
	DivinationStatusDeleted = 0
	DivinationStatusActive  = 1
	UnlockStatusLocked     = 0
	UnlockStatusUnlocked   = 1
	MethodCoinThree        = "coin_three"

	AIProviderMock             = "mock"
	AIProviderDeepSeek         = "deepseek"
	AIProviderMockFallback     = "mock_fallback"
	AIProviderTemplateFallback = "template_fallback"

	GenerationStatusPending  = 0
	GenerationStatusFreeDone = 1
	GenerationStatusFullDone = 2
	GenerationStatusFailed   = 3

	UnlockTypeMockAd            = "mock_ad"
	UnlockTypeMockButton        = "mock_button"
	UnlockTypeFreeUnlock        = "free_unlock"
	UnlockTypeRewardedVideoMock = "rewarded_video_mock"
	UnlockTypeRewardedVideo     = "rewarded_video"

	AILogStatusSuccess         = 1
	AILogStatusFailed          = 2
	AILogStatusFallbackSuccess = 3

	DailyFortuneCategoryID    int64 = 6
	DailyFortuneCategoryName        = "今日运势"
	DailyFortuneQuestion            = "我今天的整体状态和行动节奏如何？"
	DailyFortuneStatusActive        = 1
	DailyFortuneStatusDeleted       = 0
)

type Session struct {
	ID         int64     `json:"session_id"`
	SessionKey string    `json:"session_key"`
	Status     int       `json:"-"`
	CreatedAt  time.Time `json:"-"`
}

type Category struct {
	ID          int64  `json:"id"`
	Code        string `json:"code,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"-"`
	Status      int    `json:"-"`
}

type Hexagram struct {
	ID           int64  `json:"id"`
	Number       int    `json:"number"`
	Name         string `json:"name"`
	FullName     string `json:"full_name"`
	UpperTrigram string `json:"upper_trigram,omitempty"`
	LowerTrigram string `json:"lower_trigram,omitempty"`
	BinaryCode   string `json:"binary_code"`
	Summary      string `json:"summary,omitempty"`
}

type Line struct {
	Position int `json:"position"`
	Value    int `json:"value"`
	IsYang   int `json:"is_yang"`
	IsMoving int `json:"is_moving"`
}

type Divination struct {
	ID                int64     `json:"id"`
	SessionID         int64     `json:"-"`
	CategoryID        int64     `json:"-"`
	Question          string    `json:"question"`
	Method            string    `json:"method,omitempty"`
	PrimaryHexagramID int64     `json:"-"`
	ChangedHexagramID int64     `json:"-"`
	MovingLines       string    `json:"-"`
	LineSnapshot      string    `json:"-"`
	Seed              string    `json:"-"`
	UnlockStatus      int       `json:"unlock_status"`
	Status            int       `json:"-"`
	CreatedAt         time.Time `json:"created_at"`

	Category           *Category `json:"category,omitempty"`
	PrimaryHexagram    *Hexagram `json:"primary_hexagram,omitempty"`
	ChangedHexagram    *Hexagram `json:"changed_hexagram,omitempty"`
	Lines              []Line    `json:"lines,omitempty"`
	MovingLinesArray   []int     `json:"moving_lines,omitempty"`
	FreeInterpretation string    `json:"free_interpretation,omitempty"`
}

type DivinationListItem struct {
	ID              int64     `json:"id"`
	Question        string    `json:"question"`
	Category        *Category `json:"category"`
	PrimaryHexagram *Hexagram `json:"primary_hexagram"`
	ChangedHexagram *Hexagram `json:"changed_hexagram"`
	MovingLines     []int     `json:"moving_lines"`
	UnlockStatus    int       `json:"unlock_status"`
	CreatedAt       string    `json:"created_at"`
}

type PaginatedDivinations struct {
	Items    []DivinationListItem `json:"items"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Total    int64                `json:"total"`
}

type CreateDivinationRequest struct {
	SessionKey        string `json:"session_key"`
	CategoryID        int64  `json:"category_id"`
	Question          string `json:"question"`
	ConfirmDisclaimer bool   `json:"confirm_disclaimer"`
}

type CreateSessionRequest struct {
	SessionKey string `json:"session_key"`
}

type Interpretation struct {
	ID               int64      `json:"id"`
	DivinationID     int64      `json:"divination_id"`
	FreeContent      string     `json:"free_content"`
	FullContent      *string    `json:"full_content,omitempty"`
	AIProvider       string     `json:"ai_provider"`
	GenerationStatus int        `json:"generation_status"`
	GeneratedAt      *time.Time `json:"generated_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at,omitempty"`
}

type FullReport struct {
	Summary             string   `json:"summary"`
	Overall             string   `json:"overall"`
	CurrentState        string   `json:"current_state"`
	Opportunity         string   `json:"opportunity"`
	Risk                string   `json:"risk"`
	ActionSteps         []string `json:"action_steps"`
	EmotionReminder     string   `json:"emotion_reminder"`
	ReflectionQuestions []string `json:"reflection_questions"`
	Disclaimer          string   `json:"disclaimer"`
}

type UnlockRequest struct {
	SessionKey string `json:"session_key"`
	UnlockType string `json:"unlock_type"`
}

type UnlockResult struct {
	DivinationID       int64  `json:"divination_id"`
	UnlockStatus       int    `json:"unlock_status"`
	MockTransactionID  string `json:"mock_transaction_id"`
	FullInterpretation any    `json:"full_interpretation"`
}

type AIGenerationLog struct {
	ID              int64     `json:"id"`
	DivinationID    int64     `json:"divination_id"`
	QuestionSummary *string   `json:"question_summary,omitempty"`
	AIProvider      string    `json:"ai_provider"`
	ModelName       string    `json:"model_name"`
	Status          int       `json:"status"`
	DurationMs      int       `json:"duration_ms"`
	FallbackUsed    int       `json:"fallback_used"`
	ErrorMessage    *string   `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type DailyFortune struct {
	ID           int64     `json:"id"`
	SessionID    int64     `json:"session_id"`
	FortuneDate  string    `json:"fortune_date"`
	DivinationID int64     `json:"divination_id"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
