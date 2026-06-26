package qimen

const (
	professionalChiefNote = "ALG2.5 第一版值符值使：旬首索引与局数映射落宫，pending_verification"
)

// ResolveProfessionalChief derives 值符/值使 from ju, xun, and dun (first-version stable rules).
func ResolveProfessionalChief(ju ProfessionalJuResult, xun Xun, dun ProfessionalDun) ProfessionalChief {
	palaceIdx := chiefPalaceIndex(xun, ju.Ju)
	palaceName := professionalPalaceName(palaceIdx)
	stars := BuildProfessionalStars(ju.Ju, dun.Type, xun)
	doors := BuildProfessionalDoors(ju.Ju, dun.Type, xun)
	zhiFu := stars[palaceIdx]
	if palaceIdx == 5 {
		zhiFu = "天禽"
	}
	zhiShi := doors[palaceIdx]
	zhiShiPalace := palaceName
	if palaceIdx == 5 {
		zhiShi = doors[2]
		zhiShiPalace = professionalPalaceName(2)
	}
	if zhiFu == "" {
		zhiFu = "天禽"
	}
	if zhiShi == "" || zhiShi == "—" {
		zhiShi = doors[1]
		zhiShiPalace = professionalPalaceName(1)
	}
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
