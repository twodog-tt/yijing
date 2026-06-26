package qimen

const (
	ProfessionalLayoutVersionV1 = "professional_layout_v1_center_tianqin"

	LayoutRoleCenter = "center"
	LayoutRoleChief  = "chief"
	LayoutRolePalace = "palace"

	professionalLayoutBasisNote = "ALG2.5B 落盘口径版本化；默认天禽留中五宫；坤二/艮八寄宫仅结构预留"
)

// TianQinPlacementMode selects how 天禽 is placed on the nine-palace grid.
type TianQinPlacementMode string

const (
	TianQinPlacementCenter      TianQinPlacementMode = "center"
	TianQinPlacementKun2Pending TianQinPlacementMode = "kun2_pending"
	TianQinPlacementGen8Pending TianQinPlacementMode = "gen8_pending"
)

// ProfessionalLayoutConfig versions professional plate layout rules (ALG2.5B).
type ProfessionalLayoutConfig struct {
	Version                    string
	Basis                      string
	TianQinMode                TianQinPlacementMode
	ZhiShiCenterFallbackPalace int
}

// DefaultProfessionalLayoutConfig returns the active first-version layout (天禽留中五宫).
func DefaultProfessionalLayoutConfig() ProfessionalLayoutConfig {
	return ProfessionalLayoutConfig{
		Version:                    ProfessionalLayoutVersionV1,
		Basis:                      ProfessionalLayoutVersionV1,
		TianQinMode:                TianQinPlacementCenter,
		ZhiShiCenterFallbackPalace: 2, // 坤二宫
	}
}

func layoutConfigOrDefault(cfg *ProfessionalLayoutConfig) ProfessionalLayoutConfig {
	if cfg == nil {
		return DefaultProfessionalLayoutConfig()
	}
	c := *cfg
	if c.Version == "" {
		c.Version = ProfessionalLayoutVersionV1
	}
	if c.Basis == "" {
		c.Basis = c.Version
	}
	if c.TianQinMode == "" {
		c.TianQinMode = TianQinPlacementCenter
	}
	if c.ZhiShiCenterFallbackPalace < 1 || c.ZhiShiCenterFallbackPalace > 9 || c.ZhiShiCenterFallbackPalace == 5 {
		c.ZhiShiCenterFallbackPalace = 2
	}
	return c
}

func applyTianQinPlacement(stars map[int]string, mode TianQinPlacementMode) {
	switch mode {
	case TianQinPlacementCenter, TianQinPlacementKun2Pending, TianQinPlacementGen8Pending:
		// kun2/gen8 are reserved only; default path keeps 天禽 at center.
		stars[5] = "天禽"
	default:
		stars[5] = "天禽"
	}
}

func layoutRoleForPalace(index int, chiefPalaceName string) string {
	if index == 5 {
		return LayoutRoleCenter
	}
	if professionalPalaceName(index) == chiefPalaceName {
		return LayoutRoleChief
	}
	return LayoutRolePalace
}

func zhiShiForChiefPalace(palaceIdx int, doors map[int]string, cfg ProfessionalLayoutConfig) (door, palaceName string) {
	if palaceIdx == 5 {
		fallback := cfg.ZhiShiCenterFallbackPalace
		return doors[fallback], professionalPalaceName(fallback)
	}
	door = doors[palaceIdx]
	palaceName = professionalPalaceName(palaceIdx)
	if door == "" || door == "—" {
		return doors[1], professionalPalaceName(1)
	}
	return door, palaceName
}

func zhiFuForChiefPalace(palaceIdx int, stars map[int]string, cfg ProfessionalLayoutConfig) string {
	if palaceIdx == 5 || cfg.TianQinMode == TianQinPlacementCenter {
		if palaceIdx == 5 {
			return "天禽"
		}
	}
	zhiFu := stars[palaceIdx]
	if zhiFu == "" {
		return "天禽"
	}
	return zhiFu
}

// EnsureChiefPalaceConsistency aligns chief fields with assembled palaces.
func EnsureChiefPalaceConsistency(chief *ProfessionalChief, palaces []ProfessionalPalace) {
	if chief == nil {
		return
	}
	for _, p := range palaces {
		if p.Name == chief.ZhiFuPalace {
			chief.ZhiFu = p.Star
		}
		if p.Name == chief.ZhiShiPalace {
			chief.ZhiShi = p.Door
		}
	}
}

// ValidateProfessionalPalaceIntegrity checks nine-palace structural completeness.
func ValidateProfessionalPalaceIntegrity(palaces []ProfessionalPalace) bool {
	if len(palaces) != 9 {
		return false
	}
	seenIndex := map[int]bool{}
	seenName := map[string]bool{}
	for _, p := range palaces {
		if p.Index < 1 || p.Index > 9 || seenIndex[p.Index] {
			return false
		}
		seenIndex[p.Index] = true
		if p.Name == "" || seenName[p.Name] {
			return false
		}
		seenName[p.Name] = true
		if p.EarthPlateStem == "" || p.HeavenPlateStem == "" || p.Star == "" || p.Summary == "" {
			return false
		}
		if p.Index == 5 {
			if p.Star != "天禽" || p.Door != "—" || p.LayoutRole != LayoutRoleCenter {
				return false
			}
			if p.Deity != "—" && p.Deity != "值符" {
				return false
			}
			continue
		}
		if p.Door == "" || p.Deity == "" {
			return false
		}
	}
	return len(seenIndex) == 9 && len(seenName) == 9
}

// ValidateChiefPalaceConsistency checks chief against palace board.
func ValidateChiefPalaceConsistency(chief ProfessionalChief, palaces []ProfessionalPalace) bool {
	var fuPalace, shiPalace *ProfessionalPalace
	for i := range palaces {
		if palaces[i].Name == chief.ZhiFuPalace {
			fuPalace = &palaces[i]
		}
		if palaces[i].Name == chief.ZhiShiPalace {
			shiPalace = &palaces[i]
		}
	}
	if fuPalace == nil || shiPalace == nil {
		return false
	}
	if chief.ZhiFu != fuPalace.Star {
		return false
	}
	if chief.ZhiShi != shiPalace.Door {
		return false
	}
	if chief.ZhiFuPalace == professionalPalaceName(5) && chief.ZhiFu != "天禽" {
		return false
	}
	if chief.ZhiShiPalace == professionalPalaceName(5) {
		return false // 值使不应落中五
	}
	return chief.ZhiFu != "" && chief.ZhiShi != "" && chief.ZhiFuPalace != "" && chief.ZhiShiPalace != ""
}

// BuildProfessionalLayout assembles chief and palaces under a versioned layout config.
func BuildProfessionalLayout(ju ProfessionalJuResult, dun ProfessionalDun, xun Xun, cfg *ProfessionalLayoutConfig) (ProfessionalChief, []ProfessionalPalace) {
	c := layoutConfigOrDefault(cfg)
	chief := ResolveProfessionalChief(ju, xun, dun, &c)
	palaces := BuildProfessionalPalaces(ju, dun, xun, chief, &c)
	EnsureChiefPalaceConsistency(&chief, palaces)
	return chief, palaces
}
