package bazi_test

import (
	"errors"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/service/bazi"
)

func TestResolveAlgorithmVersionDefaultsToV1(t *testing.T) {
	version, err := bazi.ResolveAlgorithmVersion("")
	if err != nil {
		t.Fatalf("ResolveAlgorithmVersion: %v", err)
	}
	if version != model.AlgorithmVersionBaziSimpleV1 {
		t.Fatalf("got %q", version)
	}
}

func TestResolveAlgorithmVersionAcceptsExplicitVersions(t *testing.T) {
	for _, raw := range []string{
		model.AlgorithmVersionBaziSimpleV1,
		bazi.AlgorithmVersionBaziV2POC,
		"  " + bazi.AlgorithmVersionBaziV2POC + "  ",
	} {
		version, err := bazi.ResolveAlgorithmVersion(raw)
		if err != nil {
			t.Fatalf("ResolveAlgorithmVersion(%q): %v", raw, err)
		}
		if version != model.AlgorithmVersionBaziSimpleV1 && version != bazi.AlgorithmVersionBaziV2POC {
			t.Fatalf("unexpected version %q", version)
		}
	}
}

func TestResolveAlgorithmVersionRejectsUnknown(t *testing.T) {
	_, err := bazi.ResolveAlgorithmVersion("bazi-v3")
	if !errors.Is(err, bazi.ErrInvalidAlgorithmVersion) {
		t.Fatalf("expected ErrInvalidAlgorithmVersion, got %v", err)
	}
}
