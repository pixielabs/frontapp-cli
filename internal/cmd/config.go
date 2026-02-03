package cmd

import (
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/config"
)

type ConfigCmd struct {
	Path ConfigPathCmd `cmd:"" help:"Show configuration paths"`
}

type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Run() error {
	dir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("resolve config dir: %w", err)
	}

	configPath, err := config.ConfigPath()
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	clientsDir, err := config.ClientsDir()
	if err != nil {
		return fmt.Errorf("resolve clients dir: %w", err)
	}

	keyringDir, err := config.KeyringDir()
	if err != nil {
		return fmt.Errorf("resolve keyring dir: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Config dir:  %s\n", dir)
	fmt.Fprintf(os.Stdout, "Config file: %s\n", configPath)
	fmt.Fprintf(os.Stdout, "Clients dir: %s\n", clientsDir)
	fmt.Fprintf(os.Stdout, "Keyring dir: %s\n", keyringDir)

	return nil
}
