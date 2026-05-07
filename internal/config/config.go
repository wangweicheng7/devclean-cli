package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultConfigFilename = ".devcleanrc.json"

// FileConfig is an optional configuration overlay. Pointer fields mean:
// - nil => not set in config file
// - non-nil => set and should override defaults (unless CLI flags are provided)
type FileConfig struct {
	Profile    *string         `json:"profile,omitempty"`   // safe|dev|aggressive
	Category   *string         `json:"category,omitempty"`  // "cache,logs,build"
	Repo       *string         `json:"repo,omitempty"`      // repo root path
	WithSize   *bool           `json:"with_size,omitempty"` // default true
	Discover   *DiscoverConfig `json:"discover,omitempty"`
	ExcludeIDs []string        `json:"exclude_ids,omitempty"`
	IncludeIDs []string        `json:"include_ids,omitempty"`
}

type DiscoverConfig struct {
	Enabled  *bool    `json:"enabled,omitempty"`
	Roots    []string `json:"roots,omitempty"`
	MaxDepth *int     `json:"max_depth,omitempty"`
}

func Load(path string) (FileConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, err
	}
	var cfg FileConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return FileConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

func FindDefault(cwd string) (string, bool) {
	p := filepath.Join(cwd, DefaultConfigFilename)
	if _, err := os.Stat(p); err == nil {
		return p, true
	}
	return "", false
}

func WriteTemplate(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists (use --force to overwrite)", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	tpl := FileConfig{
		Profile:  strPtr("dev"),
		Category: strPtr("cache,logs,build"),
		Repo:     strPtr(""),
		WithSize: boolPtr(true),
		Discover: &DiscoverConfig{
			Enabled:  boolPtr(true),
			Roots:    []string{"~/Documents"},
			MaxDepth: intPtr(8),
		},
		// Optional:
		// IncludeIDs: []string{"go-build-cache"},
		// ExcludeIDs: []string{"npm-cache"},
	}

	b, err := json.MarshalIndent(tpl, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	return nil
}

func Save(path string, cfg FileConfig) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func intPtr(i int) *int       { return &i }
