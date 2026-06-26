package model

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestValidateModuleTypeAllowed(t *testing.T) {
	for _, moduleType := range []int{ModuleTypeBazi, ModuleTypeQimen} {
		if err := ValidateModuleType(moduleType); err != nil {
			t.Fatalf("expected module_type %d allowed, got %v", moduleType, err)
		}
	}
}

func TestValidateModuleTypeRejected(t *testing.T) {
	for _, moduleType := range []int{0, 3, 99, -1} {
		err := ValidateModuleType(moduleType)
		if !errors.Is(err, ErrInvalidModuleType) {
			t.Fatalf("expected module_type %d rejected, got %v", moduleType, err)
		}
	}
}

func TestValidateAlgorithmVersionAllowed(t *testing.T) {
	cases := []struct {
		moduleType int
		version    string
	}{
		{ModuleTypeBazi, AlgorithmVersionBaziSimpleV1},
		{ModuleTypeBazi, "  " + AlgorithmVersionBaziSimpleV1 + "  "},
		{ModuleTypeBazi, AlgorithmVersionBaziV2POC},
		{ModuleTypeQimen, AlgorithmVersionQimenSimpleV1},
		{ModuleTypeQimen, AlgorithmVersionQimenV2POC},
	}
	for _, tc := range cases {
		if err := ValidateAlgorithmVersion(tc.moduleType, tc.version); err != nil {
			t.Fatalf("expected %d/%q allowed, got %v", tc.moduleType, tc.version, err)
		}
	}
}

func TestValidateAlgorithmVersionRejected(t *testing.T) {
	cases := []struct {
		moduleType int
		version    string
	}{
		{ModuleTypeBazi, ""},
		{ModuleTypeBazi, "   "},
		{ModuleTypeBazi, "unknown-v1"},
		{ModuleTypeBazi, AlgorithmVersionQimenSimpleV1},
		{ModuleTypeQimen, AlgorithmVersionBaziSimpleV1},
		{ModuleTypeQimen, "qimen-v3"},
		{ModuleTypeQimen, ""},
		{99, AlgorithmVersionBaziSimpleV1},
	}
	for _, tc := range cases {
		err := ValidateAlgorithmVersion(tc.moduleType, tc.version)
		if !errors.Is(err, ErrInvalidAlgorithmVersion) {
			t.Fatalf("expected %d/%q rejected, got %v", tc.moduleType, tc.version, err)
		}
	}
}

func TestValidateJSONObjectPayload(t *testing.T) {
	validCases := []json.RawMessage{
		[]byte(`{}`),
		[]byte(`{"a":1}`),
	}
	for _, raw := range validCases {
		if err := ValidateJSONObjectPayload(raw, MaxAnalysisPayloadBytes); err != nil {
			t.Fatalf("expected valid payload, got %v for %s", err, string(raw))
		}
	}

	invalidCases := []json.RawMessage{
		nil,
		[]byte(`null`),
		[]byte(`[]`),
		[]byte(`"hello"`),
		[]byte(`1`),
		[]byte(`{invalid`),
	}
	for _, raw := range invalidCases {
		err := ValidateJSONObjectPayload(raw, MaxAnalysisPayloadBytes)
		if err == nil {
			t.Fatalf("expected invalid payload for %s", string(raw))
		}
		if !errors.Is(err, ErrInvalidJSONPayload) {
			t.Fatalf("expected ErrInvalidJSONPayload, got %v", err)
		}
	}

	large := json.RawMessage(`{"data":"` + strings.Repeat("x", MaxAnalysisPayloadBytes) + `"}`)
	if err := ValidateJSONObjectPayload(large, MaxAnalysisPayloadBytes); err == nil {
		t.Fatalf("expected oversized payload rejected")
	}
}
