// Package config loads persistent styling defaults for plaqq from a TOML file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds persistent styling defaults. All fields are optional; a nil
// pointer means the option was not set in the config file, which lets an unset
// value be distinguished from an explicit zero value (e.g. bold = false).
type Config struct {
	Color  *string `toml:"color"`
	Font   *string `toml:"font"`
	Bold   *bool   `toml:"bold"`
	Hint   *string `toml:"hint"`
	NoHint *bool   `toml:"no_hint"`
}

// Template is a commented config file with every option shown at its default.
const Template = `# plaqq configuration
# Styling defaults for the notice. CLI flags override anything set here.
# Tip: run 'plaqq config' (no subcommand) for an interactive editor.

# Notice text color: a preset name ("info", "alert", "warn", "ok", "focus"),
# a hex code (e.g. "#00f5d4"), or an ANSI index ("0"-"255"). Defaults to info.
# color = "info"

# Font used to render the notice. Block fonts: "block" (default), "heavy",
# "compact".
# font = "block"

# Render the notice text in bold.
# bold = true

# Dismiss-hint text shown beneath the notice.
# hint = "[ Press Space to dismiss ]"

# Hide the dismiss hint entirely.
# no_hint = false
`

// Path returns the path to the plaqq config file. It honors the PLAQQ_CONFIG
// environment variable; otherwise it uses the OS user config directory
// (e.g. ~/.config/plaqq/config.toml on Linux).
func Path() (string, error) {
	if p := os.Getenv("PLAQQ_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "plaqq", "config.toml"), nil
}

// Load reads and parses the config file at path. A missing file is not an
// error: it returns an empty Config so callers fall back to built-in defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &cfg, nil
}

// WriteTemplate writes the commented template to path, creating parent
// directories as needed. It refuses to overwrite an existing file.
func WriteTemplate(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Template), 0o644)
}

// Save writes cfg to path as TOML, overwriting any existing file and creating
// parent directories as needed. Only fields that are set (non-nil) are written,
// so an unset option falls through to plaqq's built-in default on load. This is
// what the interactive editor ('plaqq config') uses to persist choices.
func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# plaqq configuration — managed by 'plaqq config'.\n")
	b.WriteString("# Edit by hand or re-run 'plaqq config'. CLI flags override these.\n\n")
	if cfg.Color != nil {
		fmt.Fprintf(&b, "color = %s\n", strconv.Quote(*cfg.Color))
	}
	if cfg.Font != nil {
		fmt.Fprintf(&b, "font = %s\n", strconv.Quote(*cfg.Font))
	}
	if cfg.Bold != nil {
		fmt.Fprintf(&b, "bold = %t\n", *cfg.Bold)
	}
	if cfg.Hint != nil {
		fmt.Fprintf(&b, "hint = %s\n", strconv.Quote(*cfg.Hint))
	}
	if cfg.NoHint != nil {
		fmt.Fprintf(&b, "no_hint = %t\n", *cfg.NoHint)
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}
