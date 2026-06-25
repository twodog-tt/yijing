package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	ModuleTypeBazi  = 1
	ModuleTypeQimen = 2

	AnalysisStatusActive  = 1
	AnalysisStatusDeleted = 0

	AnalysisUnlockStatusLocked   = 0
	AnalysisUnlockStatusUnlocked = 1

	AnalysisGenerationStatusPending        = 0
	AnalysisGenerationStatusFreeDone       = 1
	AnalysisGenerationStatusFullGenerating = 2
	AnalysisGenerationStatusFullDone       = 3
	AnalysisGenerationStatusFreeFailed     = 4
	AnalysisGenerationStatusFullFailed     = 5

	AlgorithmVersionBaziSimpleV1  = "bazi-simple-v1"
	AlgorithmVersionQimenSimpleV1 = "qimen-simple-v1"

	MaxAnalysisPayloadBytes = 65536

	DefaultAnalysisPageSize = 20
	MaxAnalysisPageSize     = 100
	MaxAnalysisPage         = 10000
)

var (
	ErrInvalidModuleType       = errors.New("invalid module type")
	ErrInvalidAlgorithmVersion = errors.New("invalid algorithm version")
	ErrInvalidJSONPayload      = errors.New("invalid json payload")
)

func ValidateModuleType(moduleType int) error {
	switch moduleType {
	case ModuleTypeBazi, ModuleTypeQimen:
		return nil
	default:
		return ErrInvalidModuleType
	}
}

func ValidateAlgorithmVersion(moduleType int, version string) error {
	version = strings.TrimSpace(version)
	if version == "" {
		return ErrInvalidAlgorithmVersion
	}
	switch moduleType {
	case ModuleTypeBazi:
		if version == AlgorithmVersionBaziSimpleV1 {
			return nil
		}
	case ModuleTypeQimen:
		if version == AlgorithmVersionQimenSimpleV1 {
			return nil
		}
	default:
		return ErrInvalidAlgorithmVersion
	}
	return ErrInvalidAlgorithmVersion
}

// ValidateJSONObjectPayload ensures payload is a non-empty JSON object within size limits.
func ValidateJSONObjectPayload(raw json.RawMessage, maxBytes int) error {
	if len(raw) == 0 {
		return ErrInvalidJSONPayload
	}
	if maxBytes > 0 && len(raw) > maxBytes {
		return ErrInvalidJSONPayload
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return ErrInvalidJSONPayload
	}
	if !json.Valid(trimmed) {
		return ErrInvalidJSONPayload
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return ErrInvalidJSONPayload
	}
	return nil
}

type AnalysisRecord struct {
	ID               int64           `json:"id"`
	SessionID        int64           `json:"-"`
	ModuleType       int             `json:"module_type"`
	AlgorithmVersion string          `json:"algorithm_version"`
	CategoryID       *int64          `json:"category_id,omitempty"`
	Question         *string         `json:"question,omitempty"`
	InputPayload     json.RawMessage `json:"input_payload"`
	ResultPayload    json.RawMessage `json:"result_payload,omitempty"`
	FreeContent      *string         `json:"free_content,omitempty"`
	FullContent      *string         `json:"full_content,omitempty"`
	UnlockStatus     int             `json:"unlock_status"`
	UnlockType       *string         `json:"unlock_type,omitempty"`
	AIProvider       *string         `json:"ai_provider,omitempty"`
	GenerationStatus int             `json:"generation_status"`
	Status           int             `json:"-"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// AnalysisListItem is a minimal list projection without sensitive birth data or large blobs.
type AnalysisListItem struct {
	ID               int64     `json:"id"`
	ModuleType       int       `json:"module_type"`
	AlgorithmVersion string    `json:"algorithm_version"`
	CategoryID       *int64    `json:"category_id,omitempty"`
	Question         *string   `json:"question,omitempty"`
	UnlockStatus     int       `json:"unlock_status"`
	GenerationStatus int       `json:"generation_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type PaginatedAnalysisList struct {
	Items    []AnalysisListItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

type AnalysisUnlockResult struct {
	ID               int64  `json:"id"`
	UnlockStatus     int    `json:"unlock_status"`
	UnlockType       string `json:"unlock_type"`
	FullContent      string `json:"full_content"`
	GenerationStatus int    `json:"generation_status"`
}
