package qimen_test

import (
	"encoding/json"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

func TestBuildV2APIResultPayloadShape(t *testing.T) {
	when := parseQimenV2Time(t, "2024-03-20 09:00")
	v1, err := qimen.Calculate("我最近适合推进这个计划吗？", "career", when)
	if err != nil {
		t.Fatalf("Calculate v1: %v", err)
	}
	v2, err := qimen.CalculateV2(qimen.CalculateInputV2{
		Category: "career",
		Now:      when,
	})
	if err != nil {
		t.Fatalf("CalculateV2: %v", err)
	}

	payload, err := qimen.BuildV2APIResultPayload(v1, v2)
	if err != nil {
		t.Fatalf("BuildV2APIResultPayload: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["algorithm_version"] != qimen.AlgorithmVersionQimenV2POC {
		t.Fatalf("algorithm_version=%v", result["algorithm_version"])
	}
	for _, key := range []string{
		"situation_overview", "question_profile", "qimen_lens",
		"calendar_basis", "dun", "xun", "chief", "palaces",
	} {
		if _, ok := result[key]; !ok {
			t.Fatalf("missing %q", key)
		}
	}
	palaces, ok := result["palaces"].([]any)
	if !ok || len(palaces) != 9 {
		t.Fatalf("palaces len=%d", len(palaces))
	}
}
