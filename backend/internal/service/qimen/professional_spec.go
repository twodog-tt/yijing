package qimen

import "encoding/json"

// ALG2.3-SPEC: professional calculation roadmap — structure and constants only.
// Do NOT wire to Create API, handler, or replace qimen-v2-poc in this phase.

const (
	AlgorithmVersionQimenV2Professional = "qimen-v2-professional"

	JieqiBasisProfessionalPending = "professional_pending"
	DunMethodChaiBuPending        = "chai_bu_or_zhi_run_pending"
	DunYuanPending                = "upper_middle_lower_pending"
	DunMethodSolarTermBoundary    = "solar_term_boundary"

	MethodNoteV2Professional = "奇门 v2 专业口径（ALG2.4A–C / ALG2.5），含二十四节气拆补与九宫落盘第一版；仅供传统文化学习与结构化观察，不等同于线下大师排盘，不构成现实决策依据。"
)

// CalculationLimitsV2Professional returns limits for the professional target payload (design only).
func CalculationLimitsV2Professional() []string {
	return append([]string(nil), calculationLimitsV2Professional...)
}

var calculationLimitsV2Professional = []string{
	"当前不提供精准预测",
	"当前不提供应期断言",
	"当前不构成现实决策依据",
	"九宫落盘为 ALG2.5 第一版转盘排布（pending_verification），天禽暂保留中五宫",
	"置闰法尚未实现",
}

// ProfessionalGapAudit documents qimen-v2-poc vs qimen-v2-professional target gaps.
type ProfessionalGapAudit struct {
	Dimension       string `json:"dimension"`
	CurrentPOC      string `json:"current_poc"`
	TargetPro       string `json:"target_professional"`
	Status          string `json:"status"` // gap | partial | pending
	Implementation  string `json:"implementation_phase"`
}

// ProfessionalGapAudits is the canonical gap table for ALG2.3-SPEC documentation.
var ProfessionalGapAudits = []ProfessionalGapAudit{
	{
		Dimension: "节气交节", CurrentPOC: "复用 bazi/calendar 十二节公式近似时刻",
		TargetPro: "精确交节时刻（秒级），可与真太阳时扩展",
		Status: "partial", Implementation: "ALG2.4A",
	},
	{
		Dimension: "阴阳遁", CurrentPOC: "公历 6/21、12/22 简化切换",
		TargetPro: "与冬至/夏至节气交节时刻绑定",
		Status: "partial", Implementation: "ALG2.4A",
	},
	{
		Dimension: "局数", CurrentPOC: "hash(RFC3339+category+阴阳遁) % 9 + 1",
		TargetPro: "拆补法 / 置闰法 / 三元（上中下元）明确口径",
		Status: "partial", Implementation: "ALG2.4C",
	},
	{
		Dimension: "旬首/空亡", CurrentPOC: "六旬首固定表 + 日期/category hash",
		TargetPro: "由日时干支推导符头、旬首与空亡",
		Status: "partial", Implementation: "ALG2.4A",
	},
	{
		Dimension: "值符/值使", CurrentPOC: "按局数宫位取星/门占位",
		TargetPro: "旬首、转盘后星门落宫规则",
		Status: "partial", Implementation: "ALG2.5",
	},
	{
		Dimension: "九星/八门/八神", CurrentPOC: "固定名表 + hash 轮转",
		TargetPro: "转盘 / 飞布规则，随遁局变化",
		Status: "partial", Implementation: "ALG2.5",
	},
	{
		Dimension: "天盘干/地盘干", CurrentPOC: "十天干表取模轮转",
		TargetPro: "按局数、遁法、旬首排布",
		Status: "partial", Implementation: "ALG2.5",
	},
	{
		Dimension: "天禽寄宫", CurrentPOC: "中五宫门为 —，未专业寄宫",
		TargetPro: "明确寄坤二 / 寄艮八等流派口径并文档化",
		Status: "partial", Implementation: "ALG2.5",
	},
}

// ProfessionalCalendarBasis extends calendar metadata for professional target payload.
type ProfessionalCalendarBasis struct {
	SolarTerm     string `json:"solar_term"`
	SolarTermTime string `json:"solar_term_time"`
	JieqiBasis    string `json:"jieqi_basis"`
	TimeBasis     string `json:"time_basis"`
	Note          string `json:"note"`
}

// ProfessionalDun describes yin/yang dun, ju, and pending method metadata.
type ProfessionalDun struct {
	Type      string `json:"type"`
	Ju        int    `json:"ju"`
	Method    string `json:"method"`
	Yuan      string `json:"yuan"`
	Source    string `json:"source,omitempty"`
	BasisTerm string `json:"basis_term,omitempty"`
	BasisTime string `json:"basis_time,omitempty"`
	Note      string `json:"note,omitempty"`
}

// ProfessionalGanzhi holds four-pillar stems/branches for professional timing.
type ProfessionalGanzhi struct {
	Year  string `json:"year"`
	Month string `json:"month"`
	Day   string `json:"day"`
	Hour  string `json:"hour"`
	Basis string `json:"basis,omitempty"`
	Note  string `json:"note,omitempty"`
}

// ProfessionalChief extends chief with palace mapping.
type ProfessionalChief struct {
	ZhiFu        string `json:"zhi_fu"`
	ZhiShi       string `json:"zhi_shi"`
	ZhiFuPalace  string `json:"zhi_fu_palace"`
	ZhiShiPalace string `json:"zhi_shi_palace"`
}

// ProfessionalPalace extends POC palace with optional layout metadata.
type ProfessionalPalace struct {
	Index           int    `json:"index"`
	Name            string `json:"name"`
	EarthPlateStem  string `json:"earth_plate_stem"`
	HeavenPlateStem string `json:"heaven_plate_stem"`
	Star            string `json:"star"`
	Door            string `json:"door"`
	Deity           string `json:"deity"`
	Summary         string `json:"summary"`
	LayoutRole      string `json:"layout_role,omitempty"`
}

// CalculationResultV2Professional is the target result shape for ALG2.4+.
// Not used by CalculateV2 or Create API in ALG2.3-SPEC.
type CalculationResultV2Professional struct {
	Category       string
	CalendarBasis  ProfessionalCalendarBasis
	Dun            ProfessionalDun
	Ganzhi         ProfessionalGanzhi
	Xun            Xun
	Chief          ProfessionalChief
	Palaces        []ProfessionalPalace
	MethodNote     string
	Limits         []string
}

// ResultPayloadDraft returns a JSON-shaped draft for documentation and fixture planning.
func (c CalculationResultV2Professional) ResultPayloadDraft() (json.RawMessage, error) {
	payload := map[string]any{
		"algorithm_version": AlgorithmVersionQimenV2Professional,
		"calendar_basis":    c.CalendarBasis,
		"dun":               c.Dun,
		"ganzhi":            c.Ganzhi,
		"xun":               c.Xun,
		"chief":             c.Chief,
		"palaces":           c.Palaces,
		"method_note":       c.MethodNote,
		"limits":            c.Limits,
	}
	return json.Marshal(payload)
}

// ProfessionalFixturePlan documents a golden fixture for future professional calibration.
type ProfessionalFixturePlan struct {
	Name                    string
	When                    string
	Category                string
	FocusBoundary           string
	AssertFieldsNow         []string
	AssertFieldsProfessional []string
	CurrentPOCBehavior      string
	ProfessionalExpectation string
	NotYetAssertable        string
}

// ProfessionalFixturePlans is the ALG2.3-SPEC fixtures table (metadata only).
var ProfessionalFixturePlans = []ProfessionalFixturePlan{
	{
		Name: "lichun_before_xiaohan", When: "2024-02-04 10:30", Category: "general",
		FocusBoundary: "节令边界（小寒区间内）",
		AssertFieldsNow:         []string{"calendar_basis.solar_term", "dun.type", "dun.ju", "palaces.len=9"},
		AssertFieldsProfessional: []string{"calendar_basis.solar_term_time", "dun.method", "dun.yuan", "ganzhi.hour"},
		CurrentPOCBehavior:      "节令=小寒；阳遁；局数 hash 稳定；九宫 POC 轮转",
		ProfessionalExpectation: "交节时刻精确；拆补/置闰局数；干支旬首推导",
		NotYetAssertable:        "专业局数、转盘落宫、天禽寄宫",
	},
	{
		Name: "jingzhe_career", When: "2024-03-20 09:00", Category: "career",
		FocusBoundary: "节令切换（惊蛰）+ category 影响局数",
		AssertFieldsNow:         []string{"calendar_basis.solar_term", "dun.ju", "chief.zhi_fu", "palaces.len=9"},
		AssertFieldsProfessional: []string{"calendar_basis.solar_term_time", "dun.method", "chief.zhi_fu_palace", "ganzhi.day/hour"},
		CurrentPOCBehavior:      "节令=惊蛰；category 影响 ju/xun/宫位轮转",
		ProfessionalExpectation: "career 类重点宫与专业转盘一致",
		NotYetAssertable:        "拆补局数、值符落宫精确映射",
	},
	{
		Name: "before_xiazhi_night", When: "2024-06-20 23:30", Category: "study",
		FocusBoundary: "夏至 POC 边界前（6/20 仍阳遁）",
		AssertFieldsNow:         []string{"dun.type=yang", "calendar_basis.solar_term"},
		AssertFieldsProfessional: []string{"dun.type bound to 夏至交节", "calendar_basis.solar_term_time"},
		CurrentPOCBehavior:      "公历 6/20 → 阳遁；节令仍为芒种区间",
		ProfessionalExpectation: "夏至交节前仍为阳遁（按节气时刻）",
		NotYetAssertable:        "交节秒级时刻、置闰局",
	},
	{
		Name: "after_xiazhi_midnight", When: "2024-06-21 00:30", Category: "study",
		FocusBoundary: "夏至 POC 边界后（6/21 起阴遁）",
		AssertFieldsNow:         []string{"dun.type=yin"},
		AssertFieldsProfessional: []string{"dun flip at 夏至 solar_term_time", "dun.method"},
		CurrentPOCBehavior:      "公历 6/21 00:30 → 阴遁（POC 公历近似）",
		ProfessionalExpectation: "仅在夏至交节后切换阴遁",
		NotYetAssertable:        "若交节在白天，POC 与 professional 可能不同 — 需 ALG2.4 断言",
	},
	{
		Name: "xiaoshu_relationship", When: "2024-08-07 15:00", Category: "relationship",
		FocusBoundary: "小暑节令 + category 差异化",
		AssertFieldsNow:         []string{"calendar_basis.solar_term", "dun.type=yin", "palaces.len=9"},
		AssertFieldsProfessional: []string{"ganzhi.hour", "xun from 干支", "chief.zhi_shi_palace"},
		CurrentPOCBehavior:      "阴遁；relationship 影响 ju 与宫位",
		ProfessionalExpectation: "communication 类重点宫：兑七/巽四",
		NotYetAssertable:        "专业飞布与 POC 轮转差异",
	},
	{
		Name: "bailu_decision", When: "2024-09-22 18:30", Category: "decision",
		FocusBoundary: "白露节令 + 决策类 category",
		AssertFieldsNow:         []string{"calendar_basis.solar_term", "chief.zhi_fu", "chief.zhi_shi"},
		AssertFieldsProfessional: []string{"dun.yuan", "dun.method", "chief palace mapping"},
		CurrentPOCBehavior:      "白露区间；值符值使占位",
		ProfessionalExpectation: "决策类信息补齐 / 小步试探叙事与宫位一致",
		NotYetAssertable:        "三元、拆补局数",
	},
	{
		Name: "before_dongzhi_general", When: "2024-12-21 23:10", Category: "general",
		FocusBoundary: "冬至 POC 边界前（12/21 阴遁）",
		AssertFieldsNow:         []string{"dun.type=yin"},
		AssertFieldsProfessional: []string{"dun.type before 冬至 solar_term_time"},
		CurrentPOCBehavior:      "12/21 23:10 仍为阴遁",
		ProfessionalExpectation: "冬至交节前阴遁",
		NotYetAssertable:        "交节秒级与 POC 公历 12/22 差异",
	},
	{
		Name: "after_dongzhi_midnight", When: "2024-12-22 00:30", Category: "general",
		FocusBoundary: "冬至 POC 边界后（12/22 阳遁）",
		AssertFieldsNow:         []string{"dun.type=yang"},
		AssertFieldsProfessional: []string{"dun flip at 冬至 solar_term_time", "dun.yuan"},
		CurrentPOCBehavior:      "12/22 00:30 → 阳遁（POC 公历近似）",
		ProfessionalExpectation: "仅在冬至交节后阳遁",
		NotYetAssertable:        "若交节在白天，时刻边界与 POC 可能不同",
	},
	{
		Name: "xiaohan_career", When: "2025-02-03 11:30", Category: "career",
		FocusBoundary: "小寒节令 + 跨年",
		AssertFieldsNow:         []string{"calendar_basis.solar_term", "dun.type=yang", "palaces.len=9"},
		AssertFieldsProfessional: []string{"calendar_basis.solar_term_time", "ganzhi.year/month", "dun.method"},
		CurrentPOCBehavior:      "小寒区间；阳遁；career category hash",
		ProfessionalExpectation: "立春前后年柱/节令专业口径",
		NotYetAssertable:        "拆补与置闰流派差异",
	},
	{
		Name: "xiazhi_next_year_study", When: "2025-06-21 09:00", Category: "study",
		FocusBoundary: "夏至日 + 次年重复样例",
		AssertFieldsNow:         []string{"dun.type=yin", "calendar_basis.solar_term"},
		AssertFieldsProfessional: []string{"stable professional ju at same solar term", "ganzhi.hour"},
		CurrentPOCBehavior:      "与 2024-06-21 同类 POC 阴遁规则",
		ProfessionalExpectation: "同年夏至后专业局数可复现",
		NotYetAssertable:        "专业局数绝对值",
	},
}

// ProfessionalModuleRoadmap lists planned implementation modules for ALG2.4+.
var ProfessionalModuleRoadmap = []struct {
	Module      string
	Description string
	Phase       string
}{
	{Module: "solar_term_precise", Description: "精确节令交节时刻与查询", Phase: "ALG2.4"},
	{Module: "dun_by_solar_term", Description: "阴阳遁与节气交节绑定", Phase: "ALG2.4"},
	{Module: "ju_chaibu", Description: "拆补法局数", Phase: "ALG2.4"},
	{Module: "ju_zhirun", Description: "置闰法局数（可选流派）", Phase: "ALG2.5"},
	{Module: "sanyuan", Description: "上中下元判定", Phase: "ALG2.5"},
	{Module: "xun_from_ganzhi", Description: "旬首/空亡由日时干支推导", Phase: "ALG2.4"},
	{Module: "plate_rotation", Description: "九星八门八神转盘飞布", Phase: "ALG2.5"},
	{Module: "stem_layout", Description: "天盘/地盘干排布", Phase: "ALG2.5"},
	{Module: "tianqin_jigong", Description: "天禽寄宫规则", Phase: "ALG2.5"},
	{Module: "chief_mapping", Description: "值符值使落宫映射", Phase: "ALG2.5"},
}
