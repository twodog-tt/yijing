package qimen

import (
	"encoding/json"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

// CalculateInputV2 is the isolated qimen-v2 POC input; not wired to Create API by default.
type CalculateInputV2 struct {
	Category string
	Now      time.Time
}

// CalculationResultV2 is the isolated qimen-v2 POC result.
type CalculationResultV2 struct {
	Category       string
	TimeContext    TimeContext
	CalendarBasis  CalendarBasis
	Dun            Dun
	Xun            Xun
	Chief          Chief
	Palaces        []Palace
	MethodNote     string
	Limits         []string
}

// CalculateV2 builds a structured nine-palace POC payload without replacing qimen-simple-v1.
func CalculateV2(input CalculateInputV2) (CalculationResultV2, error) {
	now := input.Now
	if now.IsZero() {
		now = clock.Now()
	}
	now = normalizeMoment(now)
	category := NormalizeCategory(input.Category)
	if category != "" && !ValidateCategory(category) {
		category = "general"
	}
	if category == "" {
		category = "general"
	}

	basis := calendarBasisFor(now)
	dun := dunForMoment(now, category)
	palaces := buildPalaces(dun, category, now)
	chief := chiefFor(dun, palaces)
	xun := xunForMoment(now, category)

	return CalculationResultV2{
		Category: category,
		TimeContext: TimeContext{
			CreatedAt:  now.Format(time.RFC3339),
			TimeBucket: timeBucketFor(now),
		},
		CalendarBasis: basis,
		Dun:           dun,
		Xun:           xun,
		Chief:         chief,
		Palaces:       palaces,
		MethodNote:    MethodNoteV2,
		Limits:        append([]string(nil), calculationLimitsV2...),
	}, nil
}

func (c CalculationResultV2) ResultPayload() (json.RawMessage, error) {
	payload := map[string]any{
		"algorithm_version": AlgorithmVersionQimenV2POC,
		"calendar_basis":    c.CalendarBasis,
		"dun":               c.Dun,
		"xun":               c.Xun,
		"chief":             c.Chief,
		"palaces":           c.Palaces,
		"method_note":       c.MethodNote,
		"limits":            c.Limits,
		"category":          c.Category,
		"time_context":      c.TimeContext,
		"calculation_meta": map[string]any{
			"limits": c.Limits,
		},
	}
	return json.Marshal(payload)
}
