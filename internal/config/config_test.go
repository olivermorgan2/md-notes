package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupHome points HOME and XDG_CONFIG_HOME at fresh temp dirs so the
// test never reads the real user environment. It also clears the
// notes-dir env override so each test starts from a known state.
func setupHome(t *testing.T) (home, xdg string) {
	t.Helper()
	home = t.TempDir()
	xdg = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv(envVar, "")
	return home, xdg
}

// writeConfigFile writes the given TOML content to the path that
// configPath() would resolve to under the test's HOME/XDG.
func writeConfigFile(t *testing.T, xdg, content string) string {
	t.Helper()
	dir := filepath.Join(xdg, "md-notes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestLoad_FlagOverridesEnv(t *testing.T) {
	setupHome(t)
	t.Setenv(envVar, "/tmp/from-env")
	cfg, err := Load("/tmp/from-flag")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.NotesDir != "/tmp/from-flag" {
		t.Errorf("NotesDir = %q, want /tmp/from-flag", cfg.NotesDir)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	_, xdg := setupHome(t)
	writeConfigFile(t, xdg, "[paths]\ndir = \"/tmp/from-file\"\n")
	t.Setenv(envVar, "/tmp/from-env")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.NotesDir != "/tmp/from-env" {
		t.Errorf("NotesDir = %q, want /tmp/from-env", cfg.NotesDir)
	}
}

func TestLoad_FileOverridesDefault(t *testing.T) {
	_, xdg := setupHome(t)
	writeConfigFile(t, xdg, "[paths]\ndir = \"/tmp/from-file\"\n")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.NotesDir != "/tmp/from-file" {
		t.Errorf("NotesDir = %q, want /tmp/from-file", cfg.NotesDir)
	}
}

func TestLoad_DefaultWhenNothingSet(t *testing.T) {
	home, _ := setupHome(t)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := filepath.Join(home, "notes")
	if cfg.NotesDir != want {
		t.Errorf("NotesDir = %q, want %q", cfg.NotesDir, want)
	}
}

func TestLoad_MissingFileIsNotError(t *testing.T) {
	home, _ := setupHome(t)
	// No config file written under XDG; env unset; flag empty.
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error for missing config file: %v", err)
	}
	want := filepath.Join(home, "notes")
	if cfg.NotesDir != want {
		t.Errorf("NotesDir = %q, want %q (default fallback)", cfg.NotesDir, want)
	}
}

func TestLoad_MalformedTOMLErrors(t *testing.T) {
	_, xdg := setupHome(t)
	path := writeConfigFile(t, xdg, "[paths\ndir = ")
	_, err := Load("")
	if err == nil {
		t.Fatalf("Load: expected error for malformed TOML, got nil")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not mention config path %q", err, path)
	}
}

func TestLoad_UnknownKeysIgnored(t *testing.T) {
	_, xdg := setupHome(t)
	writeConfigFile(t, xdg, "[paths]\ndir = \"/tmp/from-file\"\nmystery = \"ignored\"\n[other]\nkey = 1\n")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.NotesDir != "/tmp/from-file" {
		t.Errorf("NotesDir = %q, want /tmp/from-file", cfg.NotesDir)
	}
}
