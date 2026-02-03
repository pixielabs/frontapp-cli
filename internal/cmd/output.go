package cmd

import (
	"fmt"

	"github.com/dedene/frontapp-cli/internal/config"
	"github.com/dedene/frontapp-cli/internal/output"
)

func resolveOutputMode(flags *RootFlags) (output.Mode, error) {
	cfg, err := config.ReadConfig()
	if err != nil {
		return output.Mode{}, err
	}

	mode := output.Mode{}

	if cfg.DefaultOutput != "" {
		switch cfg.DefaultOutput {
		case "json":
			mode.JSON = true
			mode.Plain = false
		case "plain":
			mode.Plain = true
			mode.JSON = false
		default:
		}
	}

	envMode := output.FromEnv()
	if envMode.JSON {
		mode.JSON = true
		mode.Plain = false
	}
	if envMode.Plain {
		mode.Plain = true
		mode.JSON = false
	}

	if flags.JSON {
		mode.JSON = true
		mode.Plain = false
	}

	if flags.Plain {
		mode.Plain = true
		mode.JSON = false
	}

	if mode.JSON && mode.Plain {
		return output.Mode{}, fmt.Errorf("cannot use both JSON and plain output")
	}

	return mode, nil
}
