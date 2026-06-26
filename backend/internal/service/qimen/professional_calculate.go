package qimen

import (
	"encoding/json"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

const (
	chiefStatusProfessionalPending  = "professional_pending"
	palaceStatusProfessionalPending = "professional_pending"
)

// CalculateInputProfessional is input for professional preview (not wired to Create API).
type CalculateInputProfessional struct {
	Category string
	Now      time.Time
	Provider SolarTermProvider
}

// CalculateProfessionalPreview builds a professional foundation payload without full plate rotation.
// It does not replace CalculateV2 or affect qimen-v2-poc behavior.
func CalculateProfessionalPreview(input CalculateInputProfessional) (*CalculationResultV2Professional, error) {
	now := input.Now
	if now.IsZero() {
		now = clock.Now()
	}
	now = normalizeProfessionalMoment(now)
	category := NormalizeCategory(input.Category)
	if category != "" && !ValidateCategory(category) {
		category = "general"
	}
	if category == "" {
		category = "general"
	}

	provider := input.Provider
	if provider == nil {
		provider = defaultSolarTermProvider()
	}

	basis := ResolveProfessionalCalendarBasis(now, provider)
	dun := ResolveProfessionalDun(now, provider)
	ganzhi := ResolveProfessionalGanZhi(now)
	xun := ResolveXunFromGanZhi(ganzhi.Day, ganzhi.Hour)

	result := &CalculationResultV2Professional{
		Category:      category,
		CalendarBasis: basis,
		Dun:           dun,
		Ganzhi:        ganzhi,
		Xun:           xun,
		Chief: ProfessionalChief{
			ZhiFu:        chiefStatusProfessionalPending,
			ZhiShi:       chiefStatusProfessionalPending,
			ZhiFuPalace:  chiefStatusProfessionalPending,
			ZhiShiPalace: chiefStatusProfessionalPending,
		},
		Palaces:    nil,
		MethodNote: MethodNoteV2Professional,
		Limits:     CalculationLimitsV2Professional(),
	}
	return result, nil
}

// ResultPayload serializes the professional preview for tests and future internal tooling.
func (c CalculationResultV2Professional) ResultPayload() (json.RawMessage, error) {
	payload := map[string]any{
		"algorithm_version": AlgorithmVersionQimenV2Professional,
		"calendar_basis":    c.CalendarBasis,
		"dun":               c.Dun,
		"ganzhi":            c.Ganzhi,
		"xun":               c.Xun,
		"chief":             c.Chief,
		"palaces":           c.Palaces,
		"method_note":       c.MethodNote,
		"limits":            c.Limits,
		"category":          c.Category,
	}
	if c.Palaces == nil {
		payload["palaces"] = []ProfessionalPalace{}
	}
	return json.Marshal(payload)
}
