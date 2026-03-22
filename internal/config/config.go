package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	General  GeneralConfig  `toml:"general"`
	LeetCode LeetCodeConfig `toml:"leetcode"`
	Editor   EditorConfig   `toml:"editor"`
	SRS      SRSConfig      `toml:"srs"`
	Theme    string         `toml:"theme"`
}

type GeneralConfig struct {
	Language    string `toml:"language"`     // preferred coding language: go, python, cpp, java...
	DataDir     string `toml:"data_dir"`     // database location
	NewPerDay   int    `toml:"new_per_day"`  // new problems per day
}

type LeetCodeConfig struct {
	Site      string `toml:"site"`        // "us" or "cn"
	Cookie    string `toml:"cookie"`      // LEETCODE_SESSION cookie (set by `leet-tui login`)
	CSRFToken string `toml:"csrf_token"`  // csrftoken cookie (set by `leet-tui login`)
	Browser   string `toml:"browser"`     // browser used for last login, informational
}

type EditorConfig struct {
	Command string `toml:"command"` // e.g. "nvim", "code", "vim"
	Args    string `toml:"args"`    // extra args
}

type SRSConfig struct {
	RetentionRate float64   `toml:"retention_rate"` // target retention, default 0.9
	Weights       []float64 `toml:"weights"`        // FSRS-5 weights (19 params)
}

func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			Language:  "go",
			DataDir:   filepath.Join(configDir(), "data"),
			NewPerDay: 3,
		},
		LeetCode: LeetCodeConfig{
			Site: "us",
		},
		Editor: EditorConfig{
			Command: defaultEditor(),
		},
		SRS: SRSConfig{
			RetentionRate: 0.9,
			Weights:       DefaultFSRSWeights(),
		},
		Theme: "default",
	}
}

func DefaultFSRSWeights() []float64 {
	return []float64{
		0.4072, 1.1829, 3.1262, 15.4722, // initial stability for Again/Hard/Good/Easy
		7.2102,                             // initial difficulty
		0.5316, 1.0651, 0.0589,            // difficulty
		1.5330, 0.1544, 1.0347,            // stability after success
		1.9395, 0.1100, 0.2939,            // stability after failure (w11-w14)
		2.0091, 0.2640, 2.9898,            // short-term stability
		0.5100, 0.6100,                    // additional params
	}
}

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "leet-tui")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "leet-tui")
}

func ConfigDir() string {
	return configDir()
}

// GetSite returns the configured LeetCode site, defaulting to "us".
func (c *Config) GetSite() string {
	if c.LeetCode.Site == "" {
		return "us"
	}
	return c.LeetCode.Site
}

func defaultEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vim"
}

func Load() (*Config, error) {
	cfg := DefaultConfig()
	path := filepath.Join(configDir(), "config.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			if err := os.MkdirAll(configDir(), 0o755); err != nil {
				return nil, fmt.Errorf("create config dir: %w", err)
			}
			if err := Save(cfg); err != nil {
				return nil, fmt.Errorf("save default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	path := filepath.Join(configDir(), "config.toml")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
