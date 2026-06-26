package bazi_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/service/bazi"
)

func TestBuildV2APIResultPayloadIncludesCompatibilityFields(t *testing.T) {
	v2, err := bazi.CalculateV2("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	calc := bazi.CalculationResultFromV2(v2)
	payload, err := bazi.BuildV2APIResultPayload(v2, calc)
	if err != nil {
		t.Fatalf("BuildV2APIResultPayload: %v", err)
	}

	raw := string(payload)
	for _, forbidden := range []string{"birth_date", "birth_hour", "session_key", "prompt", "input_payload"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("result_payload must not contain %q", forbidden)
		}
	}

	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["algorithm_version"] != bazi.AlgorithmVersionBaziV2POC {
		t.Fatalf("algorithm_version=%v", result["algorithm_version"])
	}
	for _, key := range []string{"pillars", "pillars_v2", "calendar_basis", "bazi_profile", "interpretation_lens"} {
		if _, ok := result[key]; !ok {
			t.Fatalf("missing %q", key)
		}
	}
}

func TestBuildFullContentSupportsV2Payload(t *testing.T) {
	v2, err := bazi.CalculateV2("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	calc := bazi.CalculationResultFromV2(v2)
	payload, err := bazi.BuildV2APIResultPayload(v2, calc)
	if err != nil {
		t.Fatalf("BuildV2APIResultPayload: %v", err)
	}

	content, err := bazi.BuildFullContent(payload, bazi.BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	if !strings.Contains(content, "bazi-v2-poc") {
		t.Fatalf("expected v2 disclaimer in full content")
	}
	if !strings.Contains(content, "【一、简要说明】") {
		t.Fatalf("missing required section")
	}
}

func TestBuildFullReportPromptInputSupportsV2Payload(t *testing.T) {
	v2, err := bazi.CalculateV2("2024-02-05", "wu", false)
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}
	calc := bazi.CalculationResultFromV2(v2)
	payload, err := bazi.BuildV2APIResultPayload(v2, calc)
	if err != nil {
		t.Fatalf("BuildV2APIResultPayload: %v", err)
	}

	input, err := bazi.BuildFullReportPromptInputForTest(payload, bazi.BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	if input.AlgorithmVersion != bazi.AlgorithmVersionBaziV2POC {
		t.Fatalf("algorithm_version=%q", input.AlgorithmVersion)
	}
	if input.CalendarBasis.YearBoundary != "lichun" {
		t.Fatalf("calendar_basis=%+v", input.CalendarBasis)
	}
}
