package qimen

import (
	"testing"
)

func TestResolveProfessionalChiefNotPending(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 3, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yang", Ju: 3}
	xun := Xun{XunShou: "甲子", EmptyBranches: []string{"戌", "亥"}}
	chief := ResolveProfessionalChief(ju, xun, dun, nil)
	if chief.ZhiFu == "" || chief.ZhiFu == "professional_pending" {
		t.Fatalf("zhi_fu=%q", chief.ZhiFu)
	}
	if chief.ZhiShi == "" || chief.ZhiShi == "professional_pending" {
		t.Fatalf("zhi_shi=%q", chief.ZhiShi)
	}
	if chief.ZhiFuPalace == "" || chief.ZhiShiPalace == "" {
		t.Fatalf("palace fields empty: %+v", chief)
	}
}

func TestResolveProfessionalChiefUsesXunAndJu(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 1, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yang", Ju: 1}
	a := ResolveProfessionalChief(ju, Xun{XunShou: "甲子"}, dun, nil)
	b := ResolveProfessionalChief(ju, Xun{XunShou: "甲戌"}, dun, nil)
	if a.ZhiFuPalace == b.ZhiFuPalace && a.ZhiFu == b.ZhiFu {
		t.Fatalf("expected different chief mapping for different xun: %+v vs %+v", a, b)
	}
}

func TestChiefPalaceIndexRange(t *testing.T) {
	for _, xs := range xunShouList {
		for ju := 1; ju <= 9; ju++ {
			idx := chiefPalaceIndex(Xun{XunShou: xs}, ju)
			if idx < 1 || idx > 9 {
				t.Fatalf("xun=%s ju=%d idx=%d", xs, ju, idx)
			}
		}
	}
}

func TestResolveProfessionalChiefMatchesPalaces(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 6, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yin", Ju: 6}
	for _, xs := range xunShouList {
		xun := Xun{XunShou: xs}
		chief, palaces := BuildProfessionalLayout(ju, dun, xun, nil)
		if !ValidateChiefPalaceConsistency(chief, palaces) {
			t.Fatalf("xun=%s chief inconsistent: %+v", xs, chief)
		}
	}
}

func TestResolveProfessionalChiefCenterFallbackZhiShi(t *testing.T) {
	// 值符落中五时，值使取坤二宫门
	for ju := 1; ju <= 9; ju++ {
		for _, xs := range xunShouList {
			dun := ProfessionalDun{Type: "yang", Ju: ju}
			juResult := ProfessionalJuResult{Ju: ju, Method: DunMethodChaiBu}
			xun := Xun{XunShou: xs}
			if chiefPalaceIndex(xun, ju) != 5 {
				continue
			}
			chief, palaces := BuildProfessionalLayout(juResult, dun, xun, nil)
			if chief.ZhiFu != "天禽" {
				t.Fatalf("center zhi_fu=%q want 天禽", chief.ZhiFu)
			}
			if chief.ZhiShiPalace == professionalPalaceName(5) {
				t.Fatal("zhi_shi should not fall on center palace")
			}
			kunDoor := ""
			for _, p := range palaces {
				if p.Index == 2 {
					kunDoor = p.Door
				}
			}
			if chief.ZhiShi != kunDoor {
				t.Fatalf("zhi_shi=%q want kun2 door %q", chief.ZhiShi, kunDoor)
			}
		}
	}
}
