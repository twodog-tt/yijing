package clock

import "time"

var loc *time.Location

func init() {
	var err error
	loc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
}

func Location() *time.Location {
	return loc
}

func Now() time.Time {
	return time.Now().In(loc)
}

func FormatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(loc).Format(time.RFC3339)
}
