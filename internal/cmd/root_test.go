package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// newStyleFlagCmd builds a command whose styling flags are bound to the same
// package-level vars resolveStyle reads, so flag.Set both records the value and
// marks the flag as Changed (mirroring real CLI parsing).
func newStyleFlagCmd() *cobra.Command {
	flagColor, flagBold, flagHint, flagNoHint = "", true, defaultHint, false
	c := &cobra.Command{Use: "test"}
	c.Flags().StringVar(&flagColor, "color", "", "")
	c.Flags().BoolVar(&flagBold, "bold", true, "")
	c.Flags().StringVar(&flagHint, "hint", defaultHint, "")
	c.Flags().BoolVar(&flagNoHint, "no-hint", false, "")
	return c
}

func TestResolveStyleDefaults(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "none.toml"))
	cmd := newStyleFlagCmd()

	color, bold, hint, noHint, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if color != "" || bold != true || hint != defaultHint || noHint != false {
		t.Errorf("defaults = (%q,%v,%q,%v); want (\"\",true,%q,false)", color, bold, hint, noHint, defaultHint)
	}
}

func TestResolveStyleConfigThenFlagOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"#aabbcc\"\nbold = false\nhint = \"cfg hint\"\nno_hint = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PLAQQ_CONFIG", path)

	// Config only: values come from the file.
	cmd := newStyleFlagCmd()
	color, bold, hint, noHint, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if color != "#aabbcc" || bold != false || hint != "cfg hint" || noHint != true {
		t.Errorf("config-only = (%q,%v,%q,%v); want (#aabbcc,false,cfg hint,true)", color, bold, hint, noHint)
	}

	// Flags override the config file for the flags that were set.
	cmd = newStyleFlagCmd()
	if err := cmd.Flags().Set("color", "#123456"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("bold", "true"); err != nil {
		t.Fatal(err)
	}
	color, bold, hint, noHint, err = resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if color != "#123456" || bold != true {
		t.Errorf("flag override = (%q,%v); want (#123456,true)", color, bold)
	}
	// Unset flags still fall back to the config file.
	if hint != "cfg hint" || noHint != true {
		t.Errorf("unset flags fell back wrong = (%q,%v); want (cfg hint,true)", hint, noHint)
	}
}

func TestParseColor(t *testing.T) {
	valid := []string{"#fff", "#00f5d4", "#ABCDEF", "0", "255", "213", " #fff "}
	for _, s := range valid {
		if _, err := parseColor(s); err != nil {
			t.Errorf("parseColor(%q) returned error: %v", s, err)
		}
	}

	invalid := []string{"", "#", "#zz", "#12", "#1234567", "256", "-1", "nope", "ff00ff"}
	for _, s := range invalid {
		if _, err := parseColor(s); err == nil {
			t.Errorf("parseColor(%q) expected error, got nil", s)
		}
	}
}
