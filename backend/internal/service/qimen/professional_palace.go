package qimen

const (
	professionalPalaceSummary = "用于结构化观察，不作强预测"
)

var (
	professionalEarthStemOrder = []string{"戊", "己", "庚", "辛", "壬", "癸", "丁", "丙", "乙"}
	professionalStars          = []string{"天蓬", "天芮", "天冲", "天辅", "天禽", "天心", "天柱", "天任", "天英"}
	professionalDoors          = []string{"休门", "生门", "伤门", "杜门", "景门", "死门", "惊门", "开门"}
	professionalDeities        = []string{"值符", "螣蛇", "太阴", "六合", "白虎", "玄武", "九地", "九天"}
	// doorPalaceOrder is the eight-palace door track (excluding center) in first-version layout.
	professionalDoorPalaceOrder = []int{1, 8, 3, 4, 9, 2, 7, 6}
)

func professionalPalaceName(index int) string {
	if index < 1 || index > 9 {
		return palaceNames[0]
	}
	return palaceNames[index-1]
}

// BuildProfessionalEarthPlateStems lays earth-plate stems by ju and dun direction.
func BuildProfessionalEarthPlateStems(ju int, dunType string) map[int]string {
	if ju < 1 || ju > 9 {
		ju = 1
	}
	out := make(map[int]string, 9)
	for i, stem := range professionalEarthStemOrder {
		palace := rotatePalaceIndex(ju, i, dunType)
		out[palace] = stem
	}
	return out
}

func rotatePalaceIndex(startJu, step int, dunType string) int {
	if dunType == "yin" {
		return (startJu-1-step+9*9)%9 + 1
	}
	return (startJu-1+step)%9 + 1
}

// BuildProfessionalStars lays nine stars by ju and dun direction; center uses layout config.
func BuildProfessionalStars(ju int, dunType string, _ Xun, cfg *ProfessionalLayoutConfig) map[int]string {
	c := layoutConfigOrDefault(cfg)
	if ju < 1 || ju > 9 {
		ju = 1
	}
	out := make(map[int]string, 9)
	for i, star := range professionalStars {
		palace := rotatePalaceIndex(ju, i, dunType)
		out[palace] = star
	}
	applyTianQinPlacement(out, c.TianQinMode)
	return out
}

// BuildProfessionalDoors lays eight doors by ju and dun direction; center door is —.
func BuildProfessionalDoors(ju int, dunType string, _ Xun) map[int]string {
	if ju < 1 || ju > 9 {
		ju = 1
	}
	out := make(map[int]string, 9)
	out[5] = "—"
	offset := ju - 1
	for i, door := range professionalDoors {
		var slot int
		if dunType == "yin" {
			slot = professionalDoorPalaceOrder[(i-offset+8)%8]
		} else {
			slot = professionalDoorPalaceOrder[(i+offset)%8]
		}
		out[slot] = door
	}
	return out
}

// BuildProfessionalDeities lays eight deities starting at the chief palace.
func BuildProfessionalDeities(dunType string, chief ProfessionalChief) map[int]string {
	chiefPalace := chiefPalaceIndexFromName(chief.ZhiFuPalace)
	out := make(map[int]string, 9)
	out[5] = "—"
	track := append([]int(nil), professionalDoorPalaceOrder...)
	start := 0
	for i, p := range track {
		if p == chiefPalace {
			start = i
			break
		}
	}
	for i, deity := range professionalDeities {
		var slot int
		if dunType == "yin" {
			slot = track[(start-i+8)%8]
		} else {
			slot = track[(start+i)%8]
		}
		out[slot] = deity
	}
	if chiefPalace == 5 {
		out[5] = "值符"
	}
	return out
}

// BuildProfessionalHeavenPlateStems rotates earth stems relative to chief palace.
func BuildProfessionalHeavenPlateStems(earth map[int]string, chief ProfessionalChief, dun ProfessionalDun) map[int]string {
	chiefPalace := chiefPalaceIndexFromName(chief.ZhiFuPalace)
	ordered := make([]string, 9)
	for p := 1; p <= 9; p++ {
		ordered[p-1] = earth[p]
	}
	out := make(map[int]string, 9)
	offset := chiefPalace - 1
	for p := 1; p <= 9; p++ {
		var src int
		if dun.Type == "yin" {
			src = (p - 1 + offset) % 9
		} else {
			src = (p - 1 - offset + 9) % 9
		}
		out[p] = ordered[src]
	}
	return out
}

// BuildProfessionalPalaces assembles the nine-palace professional layout (first version).
func BuildProfessionalPalaces(ju ProfessionalJuResult, dun ProfessionalDun, xun Xun, chief ProfessionalChief, cfg *ProfessionalLayoutConfig) []ProfessionalPalace {
	c := layoutConfigOrDefault(cfg)
	earth := BuildProfessionalEarthPlateStems(ju.Ju, dun.Type)
	heaven := BuildProfessionalHeavenPlateStems(earth, chief, dun)
	stars := BuildProfessionalStars(ju.Ju, dun.Type, xun, &c)
	doors := BuildProfessionalDoors(ju.Ju, dun.Type, xun)
	deities := BuildProfessionalDeities(dun.Type, chief)
	palaces := make([]ProfessionalPalace, 0, 9)
	for i := 1; i <= 9; i++ {
		palaces = append(palaces, ProfessionalPalace{
			Index:           i,
			Name:            professionalPalaceName(i),
			EarthPlateStem:  earth[i],
			HeavenPlateStem: heaven[i],
			Star:            stars[i],
			Door:            doors[i],
			Deity:           deities[i],
			Summary:         professionalPalaceSummary,
			LayoutRole:      layoutRoleForPalace(i, chief.ZhiFuPalace),
		})
	}
	return palaces
}
