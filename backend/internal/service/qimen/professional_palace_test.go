package qimen

import (
	"testing"
)

func TestBuildProfessionalEarthPlateStemsRange(t *testing.T) {
	yang := BuildProfessionalEarthPlateStems(1, "yang")
	yin := BuildProfessionalEarthPlateStems(1, "yin")
	if len(yang) != 9 || len(yin) != 9 {
		t.Fatalf("expected 9 stems, yang=%d yin=%d", len(yang), len(yin))
	}
	if yang[1] != "戊" {
		t.Fatalf("yang ju1 palace1=%q want 戊", yang[1])
	}
	if yin[1] != "戊" {
		t.Fatalf("yin ju1 palace1=%q want 戊", yin[1])
	}
	if yang[2] == yin[2] {
		t.Fatalf("yang and yin should differ at palace2: %q", yang[2])
	}
}

func TestBuildProfessionalStarsCenterTianQin(t *testing.T) {
	stars := BuildProfessionalStars(3, "yang", Xun{XunShou: "甲子"}, nil)
	if stars[5] != "天禽" {
		t.Fatalf("center star=%q want 天禽", stars[5])
	}
}

func TestBuildProfessionalDoorsCenterDash(t *testing.T) {
	doors := BuildProfessionalDoors(5, "yin", Xun{XunShou: "甲子"})
	if doors[5] != "—" {
		t.Fatalf("center door=%q want —", doors[5])
	}
	for p := 1; p <= 9; p++ {
		if p == 5 {
			continue
		}
		if doors[p] == "" {
			t.Fatalf("palace %d door empty", p)
		}
	}
}

func TestBuildProfessionalPalacesCountAndFields(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 4, Yuan: DunYuanUpper, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yang", Ju: 4}
	xun := Xun{XunShou: "甲子", EmptyBranches: []string{"戌", "亥"}}
	chief, palaces := BuildProfessionalLayout(ju, dun, xun, nil)
	if !ValidateProfessionalPalaceIntegrity(palaces) {
		t.Fatal("palace integrity failed")
	}
	if !ValidateChiefPalaceConsistency(chief, palaces) {
		t.Fatalf("chief consistency failed: %+v", chief)
	}
	if len(palaces) != 9 {
		t.Fatalf("palaces len=%d", len(palaces))
	}
	hasCenter := false
	for _, p := range palaces {
		if p.Index < 1 || p.Index > 9 {
			t.Fatalf("index=%d", p.Index)
		}
		if p.Name == "" || p.EarthPlateStem == "" || p.HeavenPlateStem == "" || p.Star == "" || p.Summary == "" {
			t.Fatalf("palace %d incomplete: %+v", p.Index, p)
		}
		if p.Index == 5 {
			hasCenter = true
			if p.LayoutRole != "center" {
				t.Fatalf("center layout_role=%q", p.LayoutRole)
			}
			if p.Star != "天禽" {
				t.Fatalf("center star=%q", p.Star)
			}
			if p.Door != "—" {
				t.Fatalf("center door=%q", p.Door)
			}
			if p.Deity != "—" && p.Deity != "值符" {
				t.Fatalf("center deity=%q", p.Deity)
			}
		} else if p.Door == "" {
			t.Fatalf("palace %d door empty", p.Index)
		}
	}
	if !hasCenter {
		t.Fatal("missing center palace")
	}
}

func TestBuildProfessionalPalacesStable(t *testing.T) {
	ju := ProfessionalJuResult{Ju: 2, Method: DunMethodChaiBu}
	dun := ProfessionalDun{Type: "yin", Ju: 2}
	xun := Xun{XunShou: "甲戌"}
	aChief, a := BuildProfessionalLayout(ju, dun, xun, nil)
	bChief, b := BuildProfessionalLayout(ju, dun, xun, nil)
	if aChief != bChief {
		t.Fatalf("chief not stable: %+v vs %+v", aChief, bChief)
	}
	if len(a) != len(b) {
		t.Fatal("length mismatch")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("palace %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
}
