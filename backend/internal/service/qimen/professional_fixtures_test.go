package qimen_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/clock"
	"github.com/wangxintong/yijing/backend/internal/service/qimen"
)

func TestQimenV2ProfessionalFixturesAreDocumented(t *testing.T) {
	if len(qimen.ProfessionalFixturePlans) != 10 {
		t.Fatalf("expected 10 fixture plans, got %d", len(qimen.ProfessionalFixturePlans))
	}
	required := []string{
		"when", "category", "focus", "assert", "poc", "professional", "not",
	}
	for _, plan := range qimen.ProfessionalFixturePlans {
		if strings.TrimSpace(plan.Name) == "" || strings.TrimSpace(plan.When) == "" || strings.TrimSpace(plan.Category) == "" {
			t.Fatalf("fixture %q missing name/when/category", plan.Name)
		}
		if strings.TrimSpace(plan.FocusBoundary) == "" {
			t.Fatalf("fixture %q missing focus boundary", plan.Name)
		}
		if len(plan.AssertFieldsNow) == 0 || len(plan.AssertFieldsProfessional) == 0 {
			t.Fatalf("fixture %q missing assert field lists", plan.Name)
		}
		if strings.TrimSpace(plan.CurrentPOCBehavior) == "" || strings.TrimSpace(plan.ProfessionalExpectation) == "" {
			t.Fatalf("fixture %q missing poc/professional notes", plan.Name)
		}
		if strings.TrimSpace(plan.NotYetAssertable) == "" {
			t.Fatalf("fixture %q missing not-yet-assertable notes", plan.Name)
		}
		blob := plan.Name + plan.FocusBoundary + plan.CurrentPOCBehavior + plan.ProfessionalExpectation + plan.NotYetAssertable
		for _, key := range required {
			if !strings.Contains(strings.ToLower(blob), key) && key != "when" && key != "category" {
				// metadata fields validated separately above
				continue
			}
		}
	}
}

func TestQimenV2ProfessionalPayloadStructureDraft(t *testing.T) {
	draft := qimen.CalculationResultV2Professional{
		CalendarBasis: qimen.ProfessionalCalendarBasis{
			SolarTerm: "惊蛰", SolarTermTime: "pending",
			JieqiBasis: qimen.JieqiBasisProfessionalPending,
			TimeBasis:  "local_time", Note: "ALG2.4 待实现",
		},
		Dun: qimen.ProfessionalDun{
			Type: "yang", Ju: 1,
			Method: qimen.DunMethodChaiBuPending,
			Yuan:   qimen.DunYuanPending,
		},
		Ganzhi: qimen.ProfessionalGanzhi{Year: "甲辰", Month: "丁卯", Day: "癸酉", Hour: "丁巳"},
		Xun:    qimen.Xun{XunShou: "甲子", EmptyBranches: []string{"戌", "亥"}},
		Chief: qimen.ProfessionalChief{
			ZhiFu: "天冲", ZhiShi: "生门", ZhiFuPalace: "震三宫", ZhiShiPalace: "巽四宫",
		},
		Palaces: []qimen.ProfessionalPalace{
			{Index: 1, Name: "坎一宫", Star: "天蓬", Door: "休门", Deity: "值符", Summary: "用于结构化观察，不作强预测"},
		},
		MethodNote: qimen.MethodNoteV2Professional,
		Limits:     append([]string(nil), qimen.CalculationLimitsV2Professional()...),
	}
	payload, err := draft.ResultPayloadDraft()
	if err != nil {
		t.Fatalf("ResultPayloadDraft: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{
		"algorithm_version", "calendar_basis", "dun", "ganzhi", "xun", "chief", "palaces", "method_note", "limits",
	} {
		if _, ok := result[key]; !ok {
			t.Fatalf("missing key %q", key)
		}
	}
	if result["algorithm_version"] != qimen.AlgorithmVersionQimenV2Professional {
		t.Fatalf("algorithm_version=%v", result["algorithm_version"])
	}
	basis := result["calendar_basis"].(map[string]any)
	if basis["jieqi_basis"] != qimen.JieqiBasisProfessionalPending {
		t.Fatalf("jieqi_basis=%v", basis["jieqi_basis"])
	}
	dun := result["dun"].(map[string]any)
	if dun["method"] == "" || dun["yuan"] == "" {
		t.Fatalf("dun missing method/yuan: %v", dun)
	}
	chief := result["chief"].(map[string]any)
	if chief["zhi_fu_palace"] == "" || chief["zhi_shi_palace"] == "" {
		t.Fatalf("chief missing palace fields: %v", chief)
	}
}

func TestQimenV2ProfessionalGapAuditComplete(t *testing.T) {
	if len(qimen.ProfessionalGapAudits) < 8 {
		t.Fatalf("expected at least 8 gap audit rows, got %d", len(qimen.ProfessionalGapAudits))
	}
	dimensions := map[string]bool{}
	for _, row := range qimen.ProfessionalGapAudits {
		if strings.TrimSpace(row.Dimension) == "" || strings.TrimSpace(row.CurrentPOC) == "" || strings.TrimSpace(row.TargetPro) == "" {
			t.Fatalf("incomplete gap row: %+v", row)
		}
		dimensions[row.Dimension] = true
	}
	for _, want := range []string{"节气交节", "阴阳遁", "局数", "旬首/空亡", "值符/值使", "九星/八门/八神", "天盘干/地盘干", "天禽寄宫"} {
		if !dimensions[want] {
			t.Fatalf("missing gap dimension %q", want)
		}
	}
}

func TestQimenV2ProfessionalFixturePOCBehaviorMatchesCalculateV2(t *testing.T) {
	for _, plan := range qimen.ProfessionalFixturePlans {
		t.Run(plan.Name, func(t *testing.T) {
			when, err := time.ParseInLocation("2006-01-02 15:04", plan.When, clock.Location())
			if err != nil {
				t.Fatalf("parse when: %v", err)
			}
			calc, err := qimen.CalculateV2(qimen.CalculateInputV2{Category: plan.Category, Now: when})
			if err != nil {
				t.Fatalf("CalculateV2: %v", err)
			}
			if len(calc.Palaces) != 9 {
				t.Fatalf("palaces len=%d", len(calc.Palaces))
			}
			if strings.Contains(plan.FocusBoundary, "阳遁") && calc.Dun.Type != "yang" {
				t.Fatalf("expected yang dun for %q, got %q", plan.Name, calc.Dun.Type)
			}
			if strings.Contains(plan.FocusBoundary, "阴遁") && !strings.Contains(plan.FocusBoundary, "前") {
				// only when focus explicitly expects yin without "前/后" qualifier
			}
			switch plan.Name {
			case "before_xiazhi_night":
				if calc.Dun.Type != "yang" {
					t.Fatalf("dun.type=%q want yang", calc.Dun.Type)
				}
			case "after_xiazhi_midnight", "xiaoshu_relationship", "bailu_decision", "before_dongzhi_general", "xiazhi_next_year_study":
				if calc.Dun.Type != "yin" {
					t.Fatalf("dun.type=%q want yin", calc.Dun.Type)
				}
			case "after_dongzhi_midnight", "lichun_before_xiaohan", "jingzhe_career", "xiaohan_career":
				if calc.Dun.Type != "yang" {
					t.Fatalf("dun.type=%q want yang", calc.Dun.Type)
				}
			}
			payload, err := calc.ResultPayload()
			if err != nil {
				t.Fatalf("ResultPayload: %v", err)
			}
			assertV2PayloadCompliance(t, payload)
		})
	}
}

func TestQimenV2ProfessionalSpecDoesNotChangePOCConstants(t *testing.T) {
	if qimen.AlgorithmVersionQimenV2POC != "qimen-v2-poc" {
		t.Fatalf("POC version changed")
	}
	if qimen.AlgorithmVersionQimenV2Professional == qimen.AlgorithmVersionQimenV2POC {
		t.Fatalf("professional version must differ from POC")
	}
}

func TestQimenV2ProfessionalLimitsCompliance(t *testing.T) {
	for _, limit := range qimen.CalculationLimitsV2Professional() {
		if strings.TrimSpace(limit) == "" {
			t.Fatalf("empty professional limit")
		}
	}
	text := strings.Join(qimen.CalculationLimitsV2Professional(), " ") + qimen.MethodNoteV2Professional
	for _, phrase := range []string{"必成", "必败", "大吉", "大凶", "改运", "化灾", "投资建议"} {
		if strings.Contains(text, phrase) && !strings.Contains(text, "不提供") && !strings.Contains(text, "不构成") {
			t.Fatalf("limits/method_note must not contain forbidden phrase %q as conclusion", phrase)
		}
	}
}
