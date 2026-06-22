package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchell-wallace/plaqq/internal/session"
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

func isolateStyleResolution(t *testing.T, configPath string) {
	t.Helper()
	t.Setenv("PLAQQ_CONFIG", configPath)
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	for _, name := range []string{envColor, envFont, envBold, envHint, envNoHint} {
		t.Setenv(name, "")
	}
}

func captureStyleWarnings(t *testing.T) *bytes.Buffer {
	t.Helper()
	var warnings bytes.Buffer
	prev := styleWarningOutput
	styleWarningOutput = &warnings
	t.Cleanup(func() { styleWarningOutput = prev })
	return &warnings
}

func TestResolveStyleDefaults(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
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
	isolateStyleResolution(t, path)

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

func TestRunERejectsUnknownFontFlagBeforeFallback(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	cmd := newStyleFlagCmd()
	if err := cmd.Flags().Set("font", "heavyy"); err != nil {
		t.Fatal(err)
	}

	err := rootCmd.RunE(cmd, []string{"hi"})
	requireErrorContains(t, err,
		`unknown font "heavyy"`,
		"from --font flag",
		"block",
		"heavy",
		"compact",
		`did you mean "heavy"?`,
	)
}

func TestResolveStyleRejectsUnknownColorFlag(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	cmd := newStyleFlagCmd()
	if err := cmd.Flags().Set("color", "infp"); err != nil {
		t.Fatal(err)
	}

	_, err := resolveStyle(cmd)
	requireErrorContains(t, err,
		`unknown color "infp"`,
		"from --color flag",
		"alert",
		"warn",
		"info",
		"ok",
		"focus",
		`did you mean "info"?`,
	)
}

func TestResolveStyleRejectsUnknownFontConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("font = \"heavyy\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	isolateStyleResolution(t, path)
	cmd := newStyleFlagCmd()

	_, err := resolveStyle(cmd)
	requireErrorContains(t, err,
		`unknown font "heavyy"`,
		"from config file",
		"block",
		"heavy",
		"compact",
		`did you mean "heavy"?`,
	)
}

func TestResolveStyleRejectsUnknownColorConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"alret\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	isolateStyleResolution(t, path)
	cmd := newStyleFlagCmd()

	_, err := resolveStyle(cmd)
	requireErrorContains(t, err,
		`unknown color "alret"`,
		"from config file",
		"alert",
		"warn",
		"info",
		"ok",
		"focus",
		`did you mean "alert"?`,
	)
}

func TestResolveStyleEnvVarsApply(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	t.Setenv(envColor, "alert")
	t.Setenv(envFont, "heavy")
	t.Setenv(envBold, "false")
	t.Setenv(envHint, "env hint")
	t.Setenv(envNoHint, "true")
	cmd := newStyleFlagCmd()

	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := styleSettings{color: "alert", font: "heavy", bold: false, hint: "env hint", noHint: true}
	if s != want {
		t.Errorf("env style = %+v; want %+v", s, want)
	}
}

func TestResolveStylePrecedenceAndPerSettingIndependence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"warn\"\nfont = \"block\"\nbold = true\nhint = \"cfg hint\"\nno_hint = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	isolateStyleResolution(t, path)
	t.Setenv(envColor, "alert")
	t.Setenv(envFont, "heavy")
	t.Setenv(envBold, "false")
	t.Setenv(envHint, "env hint")
	t.Setenv(envNoHint, "true")
	if err := session.Save(session.State{Color: "ok", Font: "compact"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	cmd := newStyleFlagCmd()
	if err := cmd.Flags().Set("color", "focus"); err != nil {
		t.Fatal(err)
	}

	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := styleSettings{color: "focus", font: "compact", bold: false, hint: "env hint", noHint: true}
	if s != want {
		t.Errorf("layered style = %+v; want %+v", s, want)
	}
}

func TestResolveStyleInvalidEnvWarnsAndFallsThrough(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"ok\"\nfont = \"compact\"\nbold = false\nhint = \"cfg hint\"\nno_hint = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	isolateStyleResolution(t, path)
	t.Setenv(envColor, "not-a-color")
	t.Setenv(envFont, "slant")
	t.Setenv(envBold, "maybe")
	t.Setenv(envNoHint, "perhaps")
	warnings := captureStyleWarnings(t)
	cmd := newStyleFlagCmd()

	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := styleSettings{color: "ok", font: "compact", bold: false, hint: "cfg hint", noHint: true}
	if s != want {
		t.Errorf("invalid env fallback = %+v; want %+v", s, want)
	}
	warn := warnings.String()
	for _, substr := range []string{envColor, envFont, envBold, envNoHint, "plaqq: warning"} {
		if !strings.Contains(warn, substr) {
			t.Fatalf("warnings %q do not contain %q", warn, substr)
		}
	}
}

func TestResolveStyleEmptyEnvFallsThroughWithoutWarning(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"ok\"\nfont = \"compact\"\nbold = false\nhint = \"cfg hint\"\nno_hint = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	isolateStyleResolution(t, path)
	warnings := captureStyleWarnings(t)
	cmd := newStyleFlagCmd()

	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := styleSettings{color: "ok", font: "compact", bold: false, hint: "cfg hint", noHint: true}
	if s != want {
		t.Errorf("empty env fallback = %+v; want %+v", s, want)
	}
	if warnings.Len() != 0 {
		t.Fatalf("empty env produced warnings: %s", warnings.String())
	}
}

func TestResolveStyleInvalidSessionWarnsAndFallsThroughPerSetting(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	t.Setenv(envColor, "alert")
	t.Setenv(envFont, "heavy")
	if err := session.Save(session.State{Color: "nope", Font: "compact"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}
	warnings := captureStyleWarnings(t)
	cmd := newStyleFlagCmd()

	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := styleSettings{color: "alert", font: "compact", bold: true, hint: defaultHint, noHint: false}
	if s != want {
		t.Errorf("invalid session fallback = %+v; want %+v", s, want)
	}
	warn := warnings.String()
	for _, substr := range []string{"session state color", `unknown color "nope"`, "plaqq: warning"} {
		if !strings.Contains(warn, substr) {
			t.Fatalf("warnings %q do not contain %q", warn, substr)
		}
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

func requireErrorContains(t *testing.T, err error, substrs ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	for _, substr := range substrs {
		if !strings.Contains(msg, substr) {
			t.Fatalf("error %q does not contain %q", msg, substr)
		}
	}
}
