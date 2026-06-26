package qimen

import "time"

const palaceSummaryPOC = "用于结构化观察，不作强预测"

var (
	palaceNames = []string{
		"坎一宫", "坤二宫", "震三宫", "巽四宫", "中五宫", "乾六宫", "兑七宫", "艮八宫", "离九宫",
	}
	palaceStars = []string{
		"天蓬", "天芮", "天冲", "天辅", "天禽", "天心", "天柱", "天任", "天英",
	}
	palaceDoors = []string{
		"休门", "生门", "伤门", "杜门", "—", "死门", "惊门", "开门", "景门",
	}
	palaceDeities = []string{
		"值符", "螣蛇", "太阴", "六合", "白虎", "玄武", "九地", "九天", "太阴",
	}
	heavenStems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	earthStems  = []string{"戊", "己", "庚", "辛", "壬", "癸", "丁", "丙", "乙", "甲"}
)

// Palace is one cell of the nine-palace grid for qimen-v2-poc.
type Palace struct {
	Index           int    `json:"index"`
	Name            string `json:"name"`
	EarthPlateStem  string `json:"earth_plate_stem"`
	HeavenPlateStem string `json:"heaven_plate_stem"`
	Star            string `json:"star"`
	Door            string `json:"door"`
	Deity           string `json:"deity"`
	Summary         string `json:"summary"`
}

func buildPalaces(dun Dun, category string, t time.Time) []Palace {
	t = normalizeMoment(t)
	rotate := (dun.Ju - 1 + int(hashSeed(NormalizeCategory(category), t.Format("2006-01-02"))%3)) % 9
	palaces := make([]Palace, 0, 9)
	for i := 0; i < 9; i++ {
		idx := (i + rotate) % 9
		palaces = append(palaces, Palace{
			Index:           i + 1,
			Name:            palaceNames[i],
			EarthPlateStem:  earthStems[(idx+dun.Ju)%len(earthStems)],
			HeavenPlateStem: heavenStems[(idx+rotate+dun.Ju)%len(heavenStems)],
			Star:            palaceStars[idx],
			Door:            palaceDoors[i],
			Deity:           palaceDeities[(idx+rotate)%len(palaceDeities)],
			Summary:         palaceSummaryPOC,
		})
	}
	return palaces
}
