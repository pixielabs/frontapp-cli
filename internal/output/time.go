package output

import (
	"sync"
	"time"

	"github.com/dedene/frontapp-cli/internal/config"
)

var (
	timezoneOnce sync.Once
	timezoneLoc  *time.Location
)

func loadLocation() *time.Location {
	timezoneOnce.Do(func() {
		cfg, err := config.ReadConfig()
		if err != nil || cfg.Timezone == "" {
			timezoneLoc = time.Local

			return
		}

		loc, err := time.LoadLocation(cfg.Timezone)
		if err != nil {
			timezoneLoc = time.Local

			return
		}

		timezoneLoc = loc
	})

	if timezoneLoc == nil {
		return time.Local
	}

	return timezoneLoc
}

func FormatTimestamp(ts float64) string {
	return FormatTimestampLayout(ts, "2006-01-02 15:04")
}

func FormatTimestampRFC3339(ts float64) string {
	return FormatTimestampLayout(ts, time.RFC3339)
}

func FormatTimestampLayout(ts float64, layout string) string {
	if ts == 0 {
		return ""
	}

	loc := loadLocation()

	return time.Unix(int64(ts), 0).In(loc).Format(layout)
}
