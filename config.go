package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// Config is the on-disk configuration, stored as YAML or TOML.
type Config struct {
	DefaultPrefix string   `yaml:"default_prefix" toml:"default_prefix"`
	Directories   []string `yaml:"directories" toml:"directories"`
}

var errNoConfig = errors.New("no config file")

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

// configPath returns the config file to use: .yml wins over .toml. If neither
// exists it returns the .yml path with exists=false.
func configPath() (path string, exists bool, err error) {
	dir, err := configDir()
	if err != nil {
		return "", false, err
	}
	yml := filepath.Join(dir, "cc-launch.yml")
	tml := filepath.Join(dir, "cc-launch.toml")
	for _, p := range []string{yml, tml} {
		if _, err := os.Stat(p); err == nil {
			return p, true, nil
		}
	}
	return yml, false, nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	switch filepath.Ext(path) {
	case ".toml":
		err = toml.Unmarshal(data, &cfg)
	default:
		err = yaml.Unmarshal(data, &cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

func saveConfig(path string, cfg *Config) error {
	var (
		data []byte
		err  error
	)
	switch filepath.Ext(path) {
	case ".toml":
		data, err = toml.Marshal(cfg)
	default:
		data, err = yaml.Marshal(cfg)
	}
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func expandHome(p string) string {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~"))
}

// tildify is the inverse of expandHome, for display.
func tildify(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == home {
		return "~"
	}
	if strings.HasPrefix(p, home+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
