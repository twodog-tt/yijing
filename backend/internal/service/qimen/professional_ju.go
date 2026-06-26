package qimen

import (
	"strings"
	"time"
)

const (
	DunMethodChaiBu        = "chai_bu"
	DunMethodZhiRunPending = "zhi_run_pending"

	DunYuanUpper  = "upper"
	DunYuanMiddle = "middle"
	DunYuanLower  = "lower"

	juBasisTwelveJieChaiBu        = "twelve_jie_chai_bu_v1"
	juBasisTwentyFourTermsChaiBu  = "twenty_four_terms_chai_bu_v1"
	juNoteChaiBuV1                = "ALG2.4B 第一版拆补局数：按十二节映射起局，三元按节内日序近似（上5/中5/下），阳遁顺行、阴遁逆行；非完整二十四节气口径"
	juNoteTwentyFourTermsChaiBuV1 = "ALG2.4C 第一版拆补局数：按二十四节气映射起局（pending_verification），三元仍用节内日序方案 A，阳遁顺行、阴遁逆行"
	yuanNoteSchemeA               = "三元按当前节气交节后的日序划分：0–4 日 upper、5–9 日 middle、10+ 日 lower；第一版近似，后续可替换"
)

// ProfessionalJuResult holds ju/yuan/method metadata for professional preview.
type ProfessionalJuResult struct {
	Method string
	Yuan   string
	Ju     int
	Basis  string
	Note   string
}

// chaiBuBaseJuByTwentyFourTerm maps all 24 solar terms to starting ju (first_version / pending_verification).
// Derived from ALG2.4B twelve-jie pairs extended to qi terms; not final professional calibration.
var chaiBuBaseJuByTwentyFourTerm = map[string]int{
	"小寒": 2, "大寒": 2,
	"立春": 8, "雨水": 8,
	"惊蛰": 1, "春分": 1,
	"清明": 3, "谷雨": 3,
	"立夏": 4, "小满": 4,
	"芒种": 5,
	"夏至": 9,
	"小暑": 8, "大暑": 8,
	"立秋": 2, "处暑": 2,
	"白露": 9,
	"秋分": 7, "寒露": 7,
	"霜降": 6, "立冬": 6,
	"小雪": 5, "大雪": 5,
	"冬至": 1,
}

// chaiBuBaseJuByJie maps the twelve 节 to starting ju (ALG2.4B twelve-jie approximation).
var chaiBuBaseJuByJie = map[string]int{
	"小寒": 2, // 阳二（对应小寒/大寒口径，仅十二节）
	"立春": 8, // 阳八
	"惊蛰": 1, // 阳一
	"清明": 3, // 阳三
	"立夏": 4, // 阳四
	"芒种": 5, // 阳五
	"小暑": 8, // 阴八
	"立秋": 2, // 阴二
	"白露": 9, // 阴九
	"寒露": 7, // 阴七
	"立冬": 6, // 阴六
	"大雪": 5, // 阴五
}

func JuBasisTwentyFourTermsChaiBu() string {
	return juBasisTwentyFourTermsChaiBu
}

// ResolveProfessionalYuan derives upper/middle/lower using scheme A (days since current节气).
// Scheme A is chosen because it aligns with solar-term-based 拆补 and fixture stability.
func ResolveProfessionalYuan(t time.Time, solarTerm string, calendar ProfessionalCalendarBasis, _ ProfessionalGanzhi) string {
	_ = solarTerm
	t = normalizeProfessionalMoment(t)
	days := daysSinceSolarTermStart(calendar.SolarTermTime, t)
	switch {
	case days < 5:
		return DunYuanUpper
	case days < 10:
		return DunYuanMiddle
	default:
		return DunYuanLower
	}
}

func daysSinceSolarTermStart(solarTermTimeRFC3339 string, t time.Time) int {
	termStart, err := time.Parse(time.RFC3339, strings.TrimSpace(solarTermTimeRFC3339))
	if err != nil || termStart.IsZero() {
		return 0
	}
	termDate := dateOnly(termStart.In(t.Location()))
	current := dateOnly(t)
	if current.Before(termDate) {
		return 0
	}
	return int(current.Sub(termDate).Hours() / 24)
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// BaseJuForProfessionalTerm returns the first-version base ju for a 24-term name.
func BaseJuForProfessionalTerm(termName string, _ string) (int, bool) {
	ju, ok := chaiBuBaseJuByTwentyFourTerm[termName]
	if !ok {
		return 0, false
	}
	return ju, true
}

// ResolveChaiBuJu computes ju 1–9 from twenty-four-term base mapping and yuan adjustment.
func ResolveChaiBuJu(t time.Time, dun ProfessionalDun, calendar ProfessionalCalendarBasis, ganZhi ProfessionalGanzhi) ProfessionalJuResult {
	t = normalizeProfessionalMoment(t)
	yuan := ResolveProfessionalYuan(t, calendar.SolarTerm, calendar, ganZhi)
	base, ok := BaseJuForProfessionalTerm(calendar.SolarTerm, dun.Type)
	if !ok {
		return ProfessionalJuResult{
			Method: DunMethodChaiBu,
			Yuan:   yuan,
			Ju:     1,
			Basis:  juBasisTwentyFourTermsChaiBu,
			Note:   juNoteTwentyFourTermsChaiBuV1 + "；未知节气回退 base=1",
		}
	}
	ju := applyChaiBuYuanOffset(base, yuan, dun.Type)
	return ProfessionalJuResult{
		Method: DunMethodChaiBu,
		Yuan:   yuan,
		Ju:     ju,
		Basis:  juBasisTwentyFourTermsChaiBu,
		Note:   juNoteTwentyFourTermsChaiBuV1 + "；" + yuanNoteSchemeA,
	}
}

func applyChaiBuYuanOffset(base int, yuan, dunType string) int {
	offset := yuanOffset(yuan)
	if dunType == "yang" {
		return normalizeJu(base + offset)
	}
	return normalizeJu(base - offset)
}

func yuanOffset(yuan string) int {
	switch yuan {
	case DunYuanMiddle:
		return 1
	case DunYuanLower:
		return 2
	default:
		return 0
	}
}

func normalizeJu(ju int) int {
	ju = ((ju - 1) % 9 + 9) % 9
	return ju + 1
}

// ResolveZhiRunJuPending reserves 置闰法 for future calibration; not used by preview default path.
func ResolveZhiRunJuPending() ProfessionalJuResult {
	return ProfessionalJuResult{
		Method: DunMethodZhiRunPending,
		Yuan:   DunYuanPending,
		Ju:     0,
		Basis:  "zhi_run_not_implemented",
		Note:   "置闰法需另行校准，当前未实现；preview 默认使用拆补法 chai_bu",
	}
}

func applyJuToDun(dun *ProfessionalDun, ju ProfessionalJuResult) {
	if dun == nil {
		return
	}
	dun.Ju = ju.Ju
	dun.Method = ju.Method
	dun.Yuan = ju.Yuan
	dun.Source = ju.Basis
	yinYangNote := dun.Note
	dun.Note = ju.Note
	if strings.TrimSpace(yinYangNote) != "" {
		dun.Note = yinYangNote + "；" + ju.Note
	}
}
