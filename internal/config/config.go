// Package config parses the umr module's per-rover config file
// (/run/waypoint/modules/umr/config.toml on the rover, written by the
// agent from modules.toml).
package config

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds the runtime configuration for the UMR module.
type Config struct {
	Host         string
	Password     string
	PollInterval time.Duration
}

// Load reads the TOML config file at path. The file is optional; an unreadable
// or absent path yields the defaults. A malformed body is still an error.
func Load(path string) (*Config, error) {
	cfg := &Config{
		Host:         "https://192.168.105.1",
		Password:     "",
		PollInterval: 5 * time.Second,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil // config is optional; defaults are fine
	}
	var raw struct {
		Host          string `toml:"host"`
		Password      string `toml:"password"`
		PollIntervalS *int   `toml:"poll_interval_s"`
	}
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return nil, err
	}
	if raw.Host != "" {
		cfg.Host = raw.Host
	}
	if raw.Password != "" {
		cfg.Password = raw.Password
	}
	if raw.PollIntervalS != nil && *raw.PollIntervalS != 0 {
		cfg.PollInterval = time.Duration(*raw.PollIntervalS) * time.Second
	}
	return cfg, nil
}
