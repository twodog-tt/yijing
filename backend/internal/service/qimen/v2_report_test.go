package qimen

import (
	"strings"
	"testing"
)

func TestPickQimenV2FocusPalacesStable(t *testing.T) {
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
	focus := pickQimenV2FocusPalaces(palaces, chief, "career")
	if len(focus) == 0 || len(focus) > 3 {
		t.Fatalf("focus len=%d", len(focus))
	}
	names := map[string]bool{}
	for _, p := range focus {
		names[p.Name] = true
	}
	if !names["震三宫"] {
		t.Fatalf("expected zhi_fu palace 震三宫, got %v", focus)
	}
}

func TestBuildV2FullContentStructure(t *testing.T) {
	content, err := BuildFullContent(sampleV2ResultPayload("career"), "免费解读")
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
		t.Fatalf("expected palace observation wording")
	}
	if !strings.Contains(content, "POC") {
		t.Fatalf("expected POC note")
	}
	if !strings.Contains(content, "乾六宫") {
		t.Fatalf("expected career focus palace name")
	}
	if strings.Contains(content, "我最近适合推进") {
		t.Fatalf("must not contain raw question")
	}
	for _, forbidden := range []string{"session_key", "input_payload", "result_payload", "prompt"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("must not contain %q", forbidden)
		}
	}
	if containsForbiddenReportPhrase(content) {
		t.Fatalf("must not contain forbidden phrases in body")
	}
}

func TestBuildV2FullContentDifferentiatesByCategory(t *testing.T) {
	contents := map[string]string{}
	for _, category := range []string{"career", "relationship", "study", "decision", "general"} {
		content, err := BuildFullContent(sampleV2ResultPayload(category), "免费解读")
		if err != nil {
			t.Fatalf("BuildFullContent %s: %v", category, err)
		}
		contents[category] = content
	}
	if contents["career"] == contents["relationship"] {
		t.Fatalf("expected different content across categories")
	}
	if !strings.Contains(contents["career"], "推进顺序") {
		t.Fatalf("career content missing marker")
	}
	if !strings.Contains(contents["relationship"], "沟通") {
		t.Fatalf("relationship content missing marker")
	}
}

func TestBuildV2FullContentReferencesDunAndChief(t *testing.T) {
	content, err := BuildFullContent(sampleV2ResultPayload("decision"), "免费解读")
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	if !strings.Contains(content, "局数") || !strings.Contains(content, "值符") {
		t.Fatalf("expected dun/chief references: %s", content)
	}
}

func TestBuildFullReportPromptInputUsesPalacesSummary(t *testing.T) {
	input, err := buildFullReportPromptInput(sampleV2ResultPayload("career"), "免费解读")
	if err != nil {
		t.Fatalf("buildFullReportPromptInput: %v", err)
	}
	if input.PalacesSummary == "" || input.PalacesSummary == "（无）" {
		t.Fatalf("expected palaces summary")
	}
	if input.FocusPalacesSummary == "" || input.FocusPalacesSummary == "（无）" {
		t.Fatalf("expected focus palaces summary")
	}
	prompt := buildQimenUserPrompt(input)
	for _, want := range []string{"palaces_summary", "focus_palaces_summary", "qimen-v2-poc", "dun", "chief"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	if strings.Contains(prompt, `"palaces"`) || strings.Contains(prompt, "input_payload") {
		t.Fatalf("prompt must not contain raw payload json keys")
	}
}

func TestBuildFullContentV1Unchanged(t *testing.T) {
	calc, err := Calculate("我最近适合推进这个计划吗？", "career", fixedNow())
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	payload, err := calc.ResultPayload()
	if err != nil {
		t.Fatalf("ResultPayload: %v", err)
	}
	content, err := BuildFullContent(payload, BuildFreeContent(calc))
	if err != nil {
		t.Fatalf("BuildFullContent: %v", err)
	}
	for _, snippet := range []string{
		sectionSummary, sectionFocus, sectionSupport, sectionRisks,
		sectionPacing, sectionReflection, sectionBoundary,
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("v1 missing %q", snippet)
		}
	}
	if strings.Contains(content, v2SectionPalaces) {
		t.Fatalf("v1 must not use v2 sections")
	}
}
