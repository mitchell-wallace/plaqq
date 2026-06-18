// Package config loads persistent styling defaults for plaqq from a TOML file.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds persistent styling defaults. All fields are optional; a nil
// pointer means the option was not set in the config file, which lets an unset
// value be distinguished from an explicit zero value (e.g. bold = false).
type Config struct {
	Color  *string `toml:"color"`
	Bold   *bool   `toml:"bold"`
	Hint   *string `toml:"hint"`
	NoHint *bool   `toml:"no_hint"`
}

// Template is a commented config file with every option shown at its default.
const Template = `# plaqq configuration
# Styling defaults for the notice. CLI flags override anything set here.

# Notice text color: a hex code (e.g. "#00f5d4") or an ANSI index ("0"-"255").
# Defaults to an adaptive teal that suits both light and dark terminals.
# color = "#00f5d4"

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
