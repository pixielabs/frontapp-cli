package output

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/dedene/frontapp-cli/internal/config"
)

func TestFormatTimestampUsesConfigTimezone(t *testing.T) {
	timezoneOnce = sync.Once{}
	timezoneLoc = nil

	temp := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", temp)

	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	cfg := config.File{Timezone: "UTC"}
	if err := config.WriteConfig(cfg); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Ensure config file exists where we expect (path varies by OS, so we ignore errors).
	_, _ = os.Stat(filepath.Join(temp, "Library", "Application Support", config.AppName, "config.yaml"))

	got := FormatTimestamp(1704067200) // 2024-01-01 00:00:00 UTC
	if got != "2024-01-01 00:00" {
		t.Fatalf("unexpected timestamp: %s", got)
	}
}
