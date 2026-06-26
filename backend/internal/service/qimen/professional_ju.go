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

	juBasisTwelveJieChaiBu = "twelve_jie_chai_bu_v1"
	juNoteChaiBuV1         = "ALG2.4B 第一版拆补局数：按十二节映射起局，三元按节内日序近似（上5/中5/下），阳遁顺行、阴遁逆行；非完整二十四节气口径"
	yuanNoteSchemeA        = "三元按当前节令交节后的日序划分：0–4 日 upper、5–9 日 middle、10+ 日 lower；第一版近似，后续可替换"
)

// ProfessionalJuResult holds ju/yuan/method metadata for professional preview.
type ProfessionalJuResult struct {
	Method string
	Yuan   string
	Ju     int
	Basis  string
	Note   string
}

// chaiBuBaseJuByJie maps the twelve 节 to starting ju (ALG2.4B twelve-jie approximation).
// Full 24-term mapping is deferred to ALG2.4C.
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

// ResolveProfessionalYuan derives upper/middle/lower using scheme A (days since current 节).
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

// ResolveChaiBuJu computes ju 1–9 from twelve-jie base mapping and yuan adjustment.
func ResolveChaiBuJu(t time.Time, dun ProfessionalDun, calendar ProfessionalCalendarBasis, ganZhi ProfessionalGanzhi) ProfessionalJuResult {
	t = normalizeProfessionalMoment(t)
	yuan := ResolveProfessionalYuan(t, calendar.SolarTerm, calendar, ganZhi)
	base, ok := chaiBuBaseJuByJie[calendar.SolarTerm]
	if !ok {
		return ProfessionalJuResult{
			Method: DunMethodChaiBu,
			Yuan:   yuan,
			Ju:     1,
			Basis:  juBasisTwelveJieChaiBu,
			Note:   juNoteChaiBuV1 + "；未知节令回退 base=1",
		}
	}
	ju := applyChaiBuYuanOffset(base, yuan, dun.Type)
	return ProfessionalJuResult{
		Method: DunMethodChaiBu,
		Yuan:   yuan,
		Ju:     ju,
		Basis:  juBasisTwelveJieChaiBu,
		Note:   juNoteChaiBuV1 + "；" + yuanNoteSchemeA,
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
	yinYangNote := dun.Note
	dun.Note = ju.Note
	if strings.TrimSpace(yinYangNote) != "" {
		dun.Note = yinYangNote + "；" + ju.Note
	}
}
