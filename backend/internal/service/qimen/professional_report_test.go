package qimen

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
)

func sampleProfessionalResultPayload(category string) json.RawMessage {
	when := time.Date(2024, 3, 20, 9, 0, 0, 0, clock.Location())
	v1, err := Calculate("test question here for professional report", category, when)
	if err != nil {
		panic(err)
	}
	pro, err := CalculateProfessionalPreview(CalculateInputProfessional{Category: category, Now: when})
	if err != nil {
		panic(err)
	}
	raw, err := BuildProfessionalAPIResultPayload(v1, pro)
	if err != nil {
		panic(err)
	}
	return raw
}

func TestPickProfessionalFocusPalacesUsesChiefPalaces(t *testing.T) {
	palaces := make([]Palace, 9)
	for i := range palaces {
		palaces[i] = Palace{
			Index: i + 1,
			Name:  palaceNames[i],
			Star:  palaceStars[i],
			Door:  palaceDoors[i],
		}
	}
	chief := Chief{
		ZhiFu:        "天任",
		ZhiShi:       "开门",
		ZhiFuPalace:  "艮八宫",
		ZhiShiPalace: "乾六宫",
	}
	focus := pickProfessionalFocusPalaces(palaces, chief, "career")
	if len(focus) < 2 || len(focus) > 3 {
		t.Fatalf("focus len=%d, want 2-3", len(focus))
	}
	names := map[string]bool{}
	for _, p := range focus {
		names[p.Name] = true
	}
	if !names["艮八宫"] || !names["乾六宫"] {
		t.Fatalf("expected chief palaces in focus, got %v", focus)
	}
}

func TestPickProfessionalFocusPalacesStableWithoutChiefPalaceFields(t *testing.T) {
	palaces := make([]Palace, 9)
	for i := range palaces {
		palaces[i] = Palace{
			Index: i + 1,
			Name:  palaceNames[i],
			Star:  palaceStars[i],
			Door:  palaceDoors[i],
		}
	}
	chief := Chief{ZhiFu: "天冲", ZhiShi: "生门"}
	focus := pickProfessionalFocusPalaces(palaces, chief, "career")
	if len(focus) < 2 {
		t.Fatalf("expected at least 2 focus palaces, got %d", len(focus))
	}
}

func TestBuildProfessionalFullContentStructure(t *testing.T) {
	content, err := BuildFullContent(sampleProfessionalResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	for _, snippet := range []string{
		v2SectionSummary, v2SectionBasis, v2SectionPalaces, v2SectionFocusPalaces,
		v2SectionSupport, v2SectionRisks, v2SectionPacing, v2SectionReflection, v2SectionBoundary,
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("missing section %q", snippet)
		}
	}
	if !strings.Contains(content, "九宫") && !strings.Contains(content, "宫位") {
		t.Fatalf("expected palace wording")
	}
	if !strings.Contains(content, ProfessionalLayoutVersionV1) && !strings.Contains(content, "第一版") {
		t.Fatalf("expected layout/first-version note")
	}
	if !strings.Contains(content, "值符") || !strings.Contains(content, "值使") {
		t.Fatalf("expected chief references")
	}
	if !strings.Contains(content, "天盘干") || !strings.Contains(content, "地盘干") {
		t.Fatalf("expected stem references in focus section")
	}
}

func TestBuildProfessionalFullContentReferencesPalaceNames(t *testing.T) {
	content, err := BuildFullContent(sampleProfessionalResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	found := false
	for _, name := range palaceNames {
		if strings.Contains(content, name) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected at least one palace name in full content")
	}
}

func TestBuildProfessionalFullContentDifferentiatesByCategory(t *testing.T) {
	contents := map[string]string{}
	for _, category := range []string{"career", "relationship", "study", "decision", "general"} {
		content, err := BuildFullContent(sampleProfessionalResultPayload(category), "免费解读")
		if err != nil {
			t.Fatalf("BuildFullContent %s: %v", category, err)
		}
		contents[category] = content
	}
	if contents["career"] == contents["relationship"] {
		t.Fatalf("expected different content across categories")
	}
	for _, marker := range []struct {
		category string
		want     string
	}{
		{"career", "推进顺序"},
		{"relationship", "沟通边界"},
		{"study", "复盘"},
		{"decision", "备用方案"},
		{"general", "问题整理"},
	} {
		if !strings.Contains(contents[marker.category], marker.want) {
			t.Fatalf("%s content missing %q", marker.category, marker.want)
		}
	}
}

func TestBuildProfessionalFullContentPrivacyAndCompliance(t *testing.T) {
	content, err := BuildFullContent(sampleProfessionalResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	for _, forbidden := range []string{
		"session_key", "input_payload", "result_payload", "prompt",
		"test question here for professional report",
		"我最近适合推进",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("must not contain %q", forbidden)
		}
	}
	if containsForbiddenReportPhrase(content) {
		t.Fatalf("must not contain forbidden phrases in body")
	}
}

func TestBuildFullReportPromptInputProfessionalFocusSummary(t *testing.T) {
	raw := sampleProfessionalResultPayload("career")
	input, err := buildFullReportPromptInput(raw, "免费解读")
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	if input.FocusPalacesSummary == "" || input.FocusPalacesSummary == "（无）" {
		t.Fatalf("expected focus palaces summary")
	}
	if !strings.Contains(input.FocusPalacesSummary, "天盘干") {
		t.Fatalf("expected detailed focus summary, got %q", input.FocusPalacesSummary)
	}
	focus := pickProfessionalFocusPalaces(input.Palaces, input.Chief, "career")
	if len(focus) < 2 || len(focus) > 3 {
		t.Fatalf("focus len=%d, want 2-3", len(focus))
	}
}

func TestBuildProfessionalUserPromptIncludesRequiredFields(t *testing.T) {
	input, err := buildFullReportPromptInput(sampleProfessionalResultPayload("decision"), "免费解读")
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	prompt := buildQimenUserPrompt(input)
	for _, want := range []string{
		"layout_version",
		"palaces_summary",
		"focus_palaces_summary",
		"ganzhi",
		"calendar_basis",
		"qimen-v2-professional",
		"九、边界声明",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	for _, forbidden := range []string{"input_payload", "result_payload", `"palaces"`, "test question here for professional report"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt must not contain %q", forbidden)
		}
	}
	if strings.Contains(prompt, "session_key=") {
		t.Fatalf("prompt must not contain session_key value")
	}
}

func TestBuildProfessionalFullContentDoesNotAffectV1OrPOC(t *testing.T) {
	v1Calc, err := Calculate("test question here", "career", fixedNow())
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	v1Payload, err := v1Calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	v1Content, err := BuildFullContent(v1Payload, BuildFreeContent(v1Calc))
	if err != nil {
		t.Fatalf("BuildFullContent v1: %v", err)
	}
	if strings.Contains(v1Content, v2SectionPalaces) {
		t.Fatalf("v1 must not use professional sections")
	}

	pocContent, err := BuildFullContent(sampleV2ResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent poc: %v", err)
	}
	if !strings.Contains(pocContent, "POC") {
		t.Fatalf("expected POC marker")
	}
	if strings.Contains(pocContent, ProfessionalLayoutVersionV1) {
		t.Fatalf("poc must not reference professional layout version")
	}
}

func TestSummarizeProfessionalFocusPalaces(t *testing.T) {
	focus := []Palace{{
		Name: "乾六宫", Star: "天心", Door: "开门", Deity: "值符",
		HeavenPlateStem: "甲", EarthPlateStem: "戊", Summary: "结构化观察",
	}}
	summary := summarizeProfessionalFocusPalaces(focus)
	for _, want := range []string{"乾六宫", "天心", "开门", "值符", "天盘干=甲", "地盘干=戊"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q: %s", want, summary)
		}
	}
}
