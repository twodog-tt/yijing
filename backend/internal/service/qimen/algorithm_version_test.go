package qimen_test

import (
	"errors"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/model"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

func TestResolveAlgorithmVersionDefaultsToV1(t *testing.T) {
	version, err := qimen.ResolveAlgorithmVersion("")
	if err != nil {
		t.Fatalf("ResolveAlgorithmVersion: %v", err)
	}
	if version != model.AlgorithmVersionQimenSimpleV1 {
		t.Fatalf("version=%q", version)
	}
}

func TestResolveAlgorithmVersionAcceptsExplicitVersions(t *testing.T) {
	for _, raw := range []string{
		model.AlgorithmVersionQimenSimpleV1,
		qimen.AlgorithmVersionQimenV2POC,
		"  " + qimen.AlgorithmVersionQimenV2POC + "  ",
	} {
		version, err := qimen.ResolveAlgorithmVersion(raw)
		if err != nil {
			t.Fatalf("ResolveAlgorithmVersion(%q): %v", raw, err)
		}
		if version != model.AlgorithmVersionQimenSimpleV1 && version != qimen.AlgorithmVersionQimenV2POC {
			t.Fatalf("unexpected version %q", version)
		}
	}
}

func TestResolveAlgorithmVersionRejectsUnknown(t *testing.T) {
	_, err := qimen.ResolveAlgorithmVersion("qimen-v3")
	if !errors.Is(err, qimen.ErrInvalidAlgorithmVersion) {
		t.Fatalf("expected ErrInvalidAlgorithmVersion, got %v", err)
	}
}
