package qimen

// ResolveProfessionalChief derives 值符/值使 from ju, xun, and dun (first-version stable rules).
func ResolveProfessionalChief(ju ProfessionalJuResult, xun Xun, dun ProfessionalDun, cfg *ProfessionalLayoutConfig) ProfessionalChief {
	c := layoutConfigOrDefault(cfg)
	palaceIdx := chiefPalaceIndex(xun, ju.Ju)
	palaceName := professionalPalaceName(palaceIdx)
	stars := BuildProfessionalStars(ju.Ju, dun.Type, xun, &c)
	doors := BuildProfessionalDoors(ju.Ju, dun.Type, xun)
	zhiFu := zhiFuForChiefPalace(palaceIdx, stars, c)
	zhiShi, zhiShiPalace := zhiShiForChiefPalace(palaceIdx, doors, c)
	return ProfessionalChief{
		ZhiFu:        zhiFu,
		ZhiShi:       zhiShi,
		ZhiFuPalace:  palaceName,
		ZhiShiPalace: zhiShiPalace,
	}
}

func chiefPalaceIndex(xun Xun, ju int) int {
	idx := 0
	for i, xs := range xunShouList {
		if xs == xun.XunShou {
			idx = i
			break
		}
	}
	if ju < 1 {
		ju = 1
	}
	return (idx+ju-1)%9 + 1
}

func chiefPalaceIndexFromName(name string) int {
	for i := 1; i <= 9; i++ {
		if professionalPalaceName(i) == name {
			return i
		}
	}
	return 1
}
