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
	flagColor, flagFont, flagBold, flagHint, flagNoHint = "", "", true, defaultHint, false
	c := &cobra.Command{Use: "test"}
	c.Flags().StringVar(&flagColor, "color", "", "")
	c.Flags().StringVar(&flagFont, "font", "", "")
	c.Flags().BoolVar(&flagBold, "bold", true, "")
	c.Flags().StringVar(&flagHint, "hint", defaultHint, "")
	c.Flags().BoolVar(&flagNoHint, "no-hint", false, "")
	return c
}

func TestResolveStyleDefaults(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "none.toml"))
	cmd := newStyleFlagCmd()

	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := styleSettings{color: "", font: "", bold: true, hint: defaultHint, noHint: false}
	if s != want {
		t.Errorf("defaults = %+v; want %+v", s, want)
	}
}

func TestResolveStyleConfigThenFlagOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"#aabbcc\"\nfont = \"heavy\"\nbold = false\nhint = \"cfg hint\"\nno_hint = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PLAQQ_CONFIG", path)

	// Config only: values come from the file.
	cmd := newStyleFlagCmd()
	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := styleSettings{color: "#aabbcc", font: "heavy", bold: false, hint: "cfg hint", noHint: true}
	if s != want {
		t.Errorf("config-only = %+v; want %+v", s, want)
	}

	// Flags override the config file for the flags that were set.
	cmd = newStyleFlagCmd()
	for flag, val := range map[string]string{"color": "#123456", "font": "compact", "bold": "true"} {
		if err := cmd.Flags().Set(flag, val); err != nil {
			t.Fatal(err)
		}
	}
	s, err = resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if s.color != "#123456" || s.font != "compact" || s.bold != true {
		t.Errorf("flag override = %+v; want color=#123456 font=compact bold=true", s)
	}
	// Unset flags still fall back to the config file.
	if s.hint != "cfg hint" || s.noHint != true {
		t.Errorf("unset flags fell back wrong = %+v; want hint='cfg hint' noHint=true", s)
	}
}

func TestParseColor(t *testing.T) {
	valid := []string{"#fff", "#00f5d4", "#ABCDEF", "0", "255", "213", " #fff ", "info", "Alert", "FOCUS"}
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
