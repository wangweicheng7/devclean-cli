package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wangweicheng7/devclean-cli/internal/config"
)

func loadConfig(configPath string) (config.FileConfig, string, error) {
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	if configPath != "" {
		cfg, err := config.Load(configPath)
		return cfg, configPath, err
	}

	if p, ok := config.FindDefault(cwd); ok {
		cfg, err := config.Load(p)
		return cfg, p, err
	}

	// fallback: global config in home directory
	if home != "" {
		p := filepath.Join(home, config.DefaultConfigFilename)
		if _, err := os.Stat(p); err == nil {
			cfg, err := config.Load(p)
			return cfg, p, err
		}
	}

	return config.FileConfig{}, "", nil
}

func applyStringFromConfig(dst *string, flagSet bool, cfgVal *string) {
	if flagSet {
		return
	}
	if cfgVal != nil {
		*dst = *cfgVal
	}
}

func applyBoolFromConfig(dst *bool, flagSet bool, cfgVal *bool) {
	if flagSet {
		return
	}
	if cfgVal != nil {
		*dst = *cfgVal
	}
}

func validateConfigPath(err error, usedPath string) error {
	if err == nil {
		return nil
	}
	if usedPath == "" {
		return err
	}
	return fmt.Errorf("config %s: %w", usedPath, err)
}

