package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`
host             = "https://192.168.105.1"
password         = "secret"
poll_interval_s  = 7
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Host != "https://192.168.105.1" || cfg.Password != "secret" {
		t.Fatalf("config: %+v", cfg)
	}
	if cfg.PollInterval != 7*time.Second {
		t.Fatalf("poll: %v", cfg.PollInterval)
	}
}

func TestLoad_DefaultPoll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("host = \"x\"\npassword = \"y\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("default poll = %v", cfg.PollInterval)
	}
}

func TestLoad_MissingFileUsesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.toml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	assertDefaults(t, cfg)
}

func TestLoad_EmptyFileUsesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(``), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	assertDefaults(t, cfg)
}

func assertDefaults(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Host != "https://192.168.105.1" {
		t.Fatalf("host = %q, want the built-in default", cfg.Host)
	}
	if cfg.Password != "" {
		t.Fatalf("password = %q, want empty", cfg.Password)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("poll = %v, want 5s", cfg.PollInterval)
	}
}
