package qimen

import (
	"testing"
)

func TestResolveProfessionalChiefNotPending(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 3, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yang", Ju: 3}
	xun := Xun{XunShou: "甲子", EmptyBranches: []string{"戌", "亥"}}
	chief := ResolveProfessionalChief(ju, xun, dun)
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
	a := ResolveProfessionalChief(ju, Xun{XunShou: "甲子"}, dun)
	b := ResolveProfessionalChief(ju, Xun{XunShou: "甲戌"}, dun)
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
