package qimen

import (
	"testing"
)

func TestDefaultProfessionalLayoutConfig(t *testing.T) {
	cfg := DefaultProfessionalLayoutConfig()
	if cfg.Version != ProfessionalLayoutVersionV1 {
		t.Fatalf("version=%q", cfg.Version)
	}
	if cfg.TianQinMode != TianQinPlacementCenter {
		t.Fatalf("mode=%q", cfg.TianQinMode)
	}
	if cfg.ZhiShiCenterFallbackPalace != 2 {
		t.Fatalf("fallback=%d", cfg.ZhiShiCenterFallbackPalace)
	}
}

func TestTianQinPendingModesStillUseCenterDefault(t *testing.T) {
	stars := map[int]string{1: "天蓬", 5: "天辅"}
	for _, mode := range []TianQinPlacementMode{TianQinPlacementKun2Pending, TianQinPlacementGen8Pending} {
		applyTianQinPlacement(stars, mode)
		if stars[5] != "天禽" {
			t.Fatalf("mode %q center=%q want 天禽", mode, stars[5])
		}
	}
}

func TestLayoutRoleCenterNotOverwrittenByChief(t *testing.T) {
	role := layoutRoleForPalace(5, professionalPalaceName(5))
	if role != LayoutRoleCenter {
		t.Fatalf("role=%q want center", role)
	}
	role = layoutRoleForPalace(3, professionalPalaceName(3))
	if role != LayoutRoleChief {
		t.Fatalf("role=%q want chief", role)
	}
}

func TestValidateProfessionalPalaceIntegrity(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 1, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yang", Ju: 1}
	xun := Xun{XunShou: "甲子"}
	chief, palaces := BuildProfessionalLayout(ju, dun, xun, nil)
	if !ValidateProfessionalPalaceIntegrity(palaces) {
		t.Fatal("palace integrity failed")
	}
	if !ValidateChiefPalaceConsistency(chief, palaces) {
		t.Fatalf("chief consistency failed: %+v", chief)
	}
}

func TestBuildProfessionalLayoutStable(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 5, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yin", Ju: 5}
	xun := Xun{XunShou: "甲申"}
	aChief, aPalaces := BuildProfessionalLayout(ju, dun, xun, nil)
	bChief, bPalaces := BuildProfessionalLayout(ju, dun, xun, nil)
	if aChief != bChief {
		t.Fatalf("chief not stable: %+v vs %+v", aChief, bChief)
	}
	for i := range aPalaces {
		if aPalaces[i] != bPalaces[i] {
			t.Fatalf("palace %d differs", i)
		}
	}
}

func TestDifferentJuChangesEarthStems(t *testing.T) {
	cfg := DefaultProfessionalLayoutConfig()
	a := BuildProfessionalEarthPlateStems(1, "yang")
	b := BuildProfessionalEarthPlateStems(5, "yang")
	if a[1] == b[1] && a[2] == b[2] {
		t.Fatal("expected different earth layout for different ju")
	}
	_ = cfg
}
