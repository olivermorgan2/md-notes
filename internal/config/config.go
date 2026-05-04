// Package config resolves the notes directory using the precedence chain
// locked in ADR-002: --dir flag > MD_NOTES_DIR env > config file > default.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the resolved runtime configuration that every notes
// command reads before doing anything else.
type Config struct {
	NotesDir string
}

// envVar is the environment variable that overrides the config file
// and the default. Per ADR-002.
const envVar = "MD_NOTES_DIR"

// defaultRelativeDir is the notes directory relative to $HOME when no
// other layer provides a value.
const defaultRelativeDir = "notes"

// fileConfig mirrors the on-disk TOML shape. Only [paths].dir is
// recognised in v0.1; unknown keys are ignored, not errors.
type fileConfig struct {
	Paths struct {
		Dir string `toml:"dir"`
	} `toml:"paths"`
}

// Load resolves the notes directory using the precedence chain
// flag > env > file > default. flag is the value of --dir from the
// CLI parser; pass "" when --dir is not set.
//
// Precedence:
//  1. flag (when non-empty)
//  2. MD_NOTES_DIR env var (when non-empty)
//  3. ${XDG_CONFIG_HOME:-$HOME/.config}/md-notes/config.toml [paths] dir
//  4. $HOME/notes
//
// A missing config file is not an error — the resolver falls through.
// A malformed config file produces an error that names the file path.
func Load(flag string) (*Config, error) {
	if flag != "" {
		return &Config{NotesDir: flag}, nil
	}
	if env := os.Getenv(envVar); env != "" {
		return &Config{NotesDir: env}, nil
	}
	dir, err := readConfigFile()
	if err != nil {
		return nil, err
	}
	if dir != "" {
		return &Config{NotesDir: dir}, nil
	}
	def, err := defaultDir()
	if err != nil {
		return nil, err
	}
	return &Config{NotesDir: def}, nil
}

// readConfigFile returns the notes directory from the user's config
// file, or "" if the file is absent or contains no [paths].dir entry.
// Any other read or parse failure is returned as an error that names
// the file path.
func readConfigFile() (string, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	var fc fileConfig
	if _, err := toml.Decode(string(data), &fc); err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}
	return fc.Paths.Dir, nil
}

// configPath is the XDG-respecting location of the config file:
// $XDG_CONFIG_HOME/md-notes/config.toml, falling back to
// $HOME/.config/md-notes/config.toml.
func configPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "md-notes", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// configPath is allowed to return a sentinel; the subsequent
		// os.ReadFile will surface a real error.
		return filepath.Join(".config", "md-notes", "config.toml")
	}
	return filepath.Join(home, ".config", "md-notes", "config.toml")
}

// defaultDir is $HOME/notes.
func defaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving default notes dir: %w", err)
	}
	return filepath.Join(home, defaultRelativeDir), nil
}
