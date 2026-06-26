package qimen_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

func TestBuildProfessionalAPIResultPayloadShape(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	v1, err := qimen.Calculate("test question here", "career", when)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	pro, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{
		Category: "career",
		Now:      when,
	})
	if err != nil {
		t.Fatalf("CalculateProfessionalPreview: %v", err)
	}
	raw, err := qimen.BuildProfessionalAPIResultPayload(v1, pro)
	if err != nil {
		t.Fatalf("BuildProfessionalAPIResultPayload: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{
		"algorithm_version", "calendar_basis", "dun", "ganzhi", "xun", "chief", "palaces",
		"layout_version", "method_note", "limits", "category", "question_profile", "qimen_lens", "safe_question_summary",
	} {
		if _, ok := obj[key]; !ok {
			t.Fatalf("missing key %q", key)
		}
	}
	if obj["algorithm_version"] != qimen.AlgorithmVersionQimenV2Professional {
		t.Fatalf("algorithm_version=%v", obj["algorithm_version"])
	}
	palaces, ok := obj["palaces"].([]any)
	if !ok || len(palaces) != 9 {
		t.Fatalf("palaces len=%d", len(palaces))
	}
	chief := obj["chief"].(map[string]any)
	if chief["zhi_fu"] == "professional_pending" || chief["zhi_fu"] == "" {
		t.Fatalf("chief zhi_fu=%v", chief["zhi_fu"])
	}
	if !strings.Contains(string(raw), qimen.ProfessionalLayoutVersionV1) {
		t.Fatalf("layout_version missing in payload")
	}
}

func TestBuildFullContentSupportsProfessionalPayload(t *testing.T) {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	v1, err := qimen.Calculate("test question here", "career", when)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	pro, err := qimen.CalculateProfessionalPreview(qimen.CalculateInputProfessional{Category: "career", Now: when})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	raw, err := qimen.BuildProfessionalAPIResultPayload(v1, pro)
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	content, err := qimen.BuildFullContent(raw, "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	for _, want := range []string{
		qimen.AlgorithmVersionQimenV2Professional,
		"九宫",
		"第一版",
		"不构成现实决策依据",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("full_content missing %q", want)
		}
	}
}
