package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchell-wallace/plaqq/internal/border"
	"github.com/mitchell-wallace/plaqq/internal/config"
	"github.com/mitchell-wallace/plaqq/internal/font"
	"github.com/mitchell-wallace/plaqq/internal/session"
	"github.com/spf13/pflag"
)

func TestSeedFormStateTolerance(t *testing.T) {
	// 1. Config with unknown/removed font and color should fall back to defaults
	badFont := "slant"
	badColor := "coral"
	cfg1 := &config.Config{
		Font:  &badFont,
		Color: &badColor,
	}

	state1 := seedFormState(cfg1)
	fontChoice, colorChoice, customColor, bold, showHint, hintText, frameChoice := state1.fontChoice, state1.colorChoice, state1.customColor, state1.bold, state1.showHint, state1.hintText, state1.frameChoice

	if fontChoice != font.DefaultName {
		t.Errorf("expected fontChoice to fall back to %q, got %q", font.DefaultName, fontChoice)
	}
	if colorChoice != colorDefaultChoice {
		t.Errorf("expected colorChoice to fall back to %q, got %q", colorDefaultChoice, colorChoice)
	}
	if customColor != "" {
		t.Errorf("expected customColor to be empty, got %q", customColor)
	}
	if !bold {
		t.Error("expected bold to default to true")
	}
	if !showHint {
		t.Error("expected showHint to default to true")
	}
	if hintText != defaultHint {
		t.Errorf("expected hintText to default to %q, got %q", defaultHint, hintText)
	}
	if frameChoice != "" {
		t.Errorf("expected frameChoice to be empty when config Frame is nil, got %q", frameChoice)
	}

	// 2. Config with valid font and color preset should be seeded correctly
	goodFont := "heavy"
	goodColorPreset := "focus"
	cfg2 := &config.Config{
		Font:  &goodFont,
		Color: &goodColorPreset,
	}

	state2 := seedFormState(cfg2)
	fontChoice2, colorChoice2, customColor2 := state2.fontChoice, state2.colorChoice, state2.customColor
	if fontChoice2 != goodFont {
		t.Errorf("expected fontChoice to be %q, got %q", goodFont, fontChoice2)
	}
	if colorChoice2 != goodColorPreset {
		t.Errorf("expected colorChoice to be %q, got %q", goodColorPreset, colorChoice2)
	}
	if customColor2 != "" {
		t.Errorf("expected customColor to be empty, got %q", customColor2)
	}

	// 3. Config with valid custom hex / ANSI color should seedcolorChoice as custom
	goodCustomColor := "#123456"
	cfg3 := &config.Config{
		Color: &goodCustomColor,
	}

	state3 := seedFormState(cfg3)
	colorChoice3, customColor3 := state3.colorChoice, state3.customColor
	if colorChoice3 != colorCustomChoice {
		t.Errorf("expected colorChoice to be %q, got %q", colorCustomChoice, colorChoice3)
	}
	if customColor3 != goodCustomColor {
		t.Errorf("expected customColor to be %q, got %q", goodCustomColor, customColor3)
	}
}

func TestSeedFormStateFrame(t *testing.T) {
	// Unset Frame -> frameChoice remains empty
	cfg1 := &config.Config{}
	if got := seedFormState(cfg1).frameChoice; got != "" {
		t.Errorf("expected frameChoice to be empty when cfg.Frame is nil, got %q", got)
	}

	// Valid frame -> seeded
	good := "block"
	cfg2 := &config.Config{Frame: &good}
	if got := seedFormState(cfg2).frameChoice; got != "block" {
		t.Errorf("expected frameChoice to be %q, got %q", "block", got)
	}

	// Invalid frame -> tolerance, frameChoice stays empty
	bad := "nope"
	cfg3 := &config.Config{Frame: &bad}
	if got := seedFormState(cfg3).frameChoice; got != "" {
		t.Errorf("expected frameChoice to be empty for unknown frame, got %q", got)
	}
}

func TestBuildSavedConfig(t *testing.T) {
	// 1. If choices are defaults, the saved Config should have nil fields to remain minimal
	cfg := buildSavedConfig(font.DefaultName, colorDefaultChoice, "", true, true, defaultHint, border.DefaultName)

	if cfg.Font != nil {
		t.Errorf("expected default font to be omitted (nil), got %q", *cfg.Font)
	}
	if cfg.Color != nil {
		t.Errorf("expected default color to be omitted (nil), got %q", *cfg.Color)
	}
	if cfg.Frame != nil {
		t.Errorf("expected default frame to be omitted (nil), got %q", *cfg.Frame)
	}
	if cfg.Bold == nil || !*cfg.Bold {
		t.Errorf("expected bold to be saved as true, got %v", cfg.Bold)
	}
	if cfg.NoHint == nil || *cfg.NoHint {
		t.Errorf("expected noHint to be saved as false, got %v", cfg.NoHint)
	}
	if cfg.Hint != nil {
		t.Errorf("expected default hint to be omitted (nil), got %q", *cfg.Hint)
	}

	// 2. If choices are custom/presets, they should be written
	cfg2 := buildSavedConfig("heavy", "warn", "", false, false, "custom hint", "single")

	if cfg2.Font == nil || *cfg2.Font != "heavy" {
		t.Errorf("expected font to be saved as 'heavy', got %v", cfg2.Font)
	}
	if cfg2.Color == nil || *cfg2.Color != "warn" {
		t.Errorf("expected color to be saved as 'warn', got %v", cfg2.Color)
	}
	if cfg2.Frame == nil || *cfg2.Frame != "single" {
		t.Errorf("expected frame to be saved as 'single', got %v", cfg2.Frame)
	}
	if cfg2.Bold == nil || *cfg2.Bold != false {
		t.Errorf("expected bold to be saved as false, got %v", cfg2.Bold)
	}
	if cfg2.NoHint == nil || *cfg2.NoHint != true {
		t.Errorf("expected no_hint to be saved as true, got %v", cfg2.NoHint)
	}
	// Note: hint text is not saved if showHint is false, because it's hidden. Let's test showHint = true with custom hint.
	cfg3 := buildSavedConfig("compact", colorCustomChoice, "#aabbcc", true, true, "custom hint", "block")
	if cfg3.Color == nil || *cfg3.Color != "#aabbcc" {
		t.Errorf("expected custom color to be saved as '#aabbcc', got %v", cfg3.Color)
	}
	if cfg3.Frame == nil || *cfg3.Frame != "block" {
		t.Errorf("expected frame to be saved as 'block', got %v", cfg3.Frame)
	}
	if cfg3.Hint == nil || *cfg3.Hint != "custom hint" {
		t.Errorf("expected custom hint to be saved, got %v", cfg3.Hint)
	}
}

func resetConfigFlags() {
	flagSession = false
	flagConfigClear = false
	flagConfigColor = ""
	flagConfigFont = ""
	flagConfigFrame = ""
	jsonOutput = false
	rootCmd.SetArgs(nil)
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	configCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

func TestConfigSessionSet(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()
	if err := session.Save(session.State{Text: "last deploy"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	rootCmd.SetArgs([]string{"config", "--session", "--color", "alert", "--font", "heavy"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config --session: %v", err)
	}

	state, err := session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	if state.Color != "alert" {
		t.Errorf("expected session color 'alert', got %q", state.Color)
	}
	if state.Font != "heavy" {
		t.Errorf("expected session font 'heavy', got %q", state.Font)
	}
	if state.Text != "last deploy" {
		t.Errorf("expected session text to be preserved, got %q", state.Text)
	}
}

func TestConfigSessionClear(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()

	// First, save some session state
	if err := session.Save(session.State{Color: "focus", Font: "compact"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	// Make sure it exists
	state, err := session.Load(nil)
	if err != nil || state.Empty() {
		t.Fatalf("expected non-empty session before clear")
	}

	// Now clear it
	rootCmd.SetArgs([]string{"config", "--session", "--clear"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config --session --clear: %v", err)
	}

	// Verify it was cleared
	state, err = session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	if !state.Empty() {
		t.Errorf("expected empty session state after clear, got %+v", state)
	}
}

func TestConfigSessionClearIgnoresBrokenPersistentConfig(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("PLAQQ_CONFIG", filepath.Join(configDir, "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()

	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte("[[broken\n"), 0o644); err != nil {
		t.Fatalf("write broken config: %v", err)
	}
	if err := session.Save(session.State{Color: "focus", Font: "compact"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	rootCmd.SetArgs([]string{"config", "--session", "--clear"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config --session --clear with broken config: %v", err)
	}

	state, err := session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	if !state.Empty() {
		t.Errorf("expected empty session state after clear, got %+v", state)
	}
}

func TestConfigSessionSetIgnoresBrokenPersistentConfig(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("PLAQQ_CONFIG", filepath.Join(configDir, "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()

	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte("[[broken\n"), 0o644); err != nil {
		t.Fatalf("write broken config: %v", err)
	}

	rootCmd.SetArgs([]string{"config", "--session", "--color", "alert", "--font", "heavy"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config --session with broken config: %v", err)
	}

	state, err := session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	if state.Color != "alert" || state.Font != "heavy" {
		t.Errorf("expected alert/heavy session state, got %+v", state)
	}
}

func TestConfigValidationErrors(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "clear without session",
			args:    []string{"--clear"},
			wantErr: "flag --clear requires --session",
		},
		{
			name:    "color without session",
			args:    []string{"--color", "alert"},
			wantErr: "flags --color and --font require --session",
		},
		{
			name:    "font without session",
			args:    []string{"--font", "heavy"},
			wantErr: "flags --color and --font require --session",
		},
		{
			name:    "invalid session color",
			args:    []string{"--session", "--color", "invalid_color"},
			wantErr: "unknown color",
		},
		{
			name:    "invalid session font",
			args:    []string{"--session", "--font", "invalid_font"},
			wantErr: "unknown font",
		},
		{
			name:    "frame without session",
			args:    []string{"--frame", "single"},
			wantErr: "flag --frame requires --session",
		},
		{
			name:    "invalid session frame",
			args:    []string{"--session", "--frame", "nope_frame"},
			wantErr: "unknown frame",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigFlags()
			rootCmd.SetArgs(append([]string{"config"}, tt.args...))
			err := rootCmd.Execute()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestConfigSessionJSON(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set session state directly
	if err := session.Save(session.State{Color: "alert", Font: "heavy"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	rootCmd.SetArgs([]string{"config", "--session", "--json-output"})
	err := rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("config --session --json-output: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, `"color":"alert"`) || !strings.Contains(out, `"font":"heavy"`) {
		t.Errorf("expected output to contain json representations of color and font, got %q", out)
	}
}

func TestConfigSessionSetFrame(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()
	if err := session.Save(session.State{Text: "last deploy", Color: "alert"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	rootCmd.SetArgs([]string{"config", "--session", "--frame", "single"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config --session --frame: %v", err)
	}

	state, err := session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	if state.Frame != "single" {
		t.Errorf("expected session frame 'single', got %q", state.Frame)
	}
	if state.Text != "last deploy" {
		t.Errorf("expected session text to be preserved, got %q", state.Text)
	}
	if state.Color != "alert" {
		t.Errorf("expected session color to be preserved, got %q", state.Color)
	}
}

func TestConfigSessionClearAfterFrame(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()

	if err := session.Save(session.State{Color: "focus", Font: "compact", Frame: "single"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	state, err := session.Load(nil)
	if err != nil || state.Empty() {
		t.Fatalf("expected non-empty session before clear")
	}

	rootCmd.SetArgs([]string{"config", "--session", "--clear"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config --session --clear: %v", err)
	}

	state, err = session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	if !state.Empty() {
		t.Errorf("expected empty session state after clear, got %+v", state)
	}
}

func TestConfigSessionJSONWithFrame(t *testing.T) {
	t.Setenv("PLAQQ_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resetConfigFlags()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := session.Save(session.State{Color: "alert", Font: "heavy", Frame: "block"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	rootCmd.SetArgs([]string{"config", "--session", "--json-output"})
	err := rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("config --session --json-output: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, `"frame":"block"`) {
		t.Errorf("expected output to contain \"frame\":\"block\", got %q", out)
	}
}

func TestFrameOptions(t *testing.T) {
	opts := frameOptions()
	if len(opts) != 4 {
		t.Fatalf("expected 4 frame options, got %d", len(opts))
	}
	if opts[0].Value != border.DefaultName {
		t.Errorf("expected first option to be the default %q, got %q", border.DefaultName, opts[0].Value)
	}
	if !strings.Contains(opts[0].Key, "(default)") {
		t.Errorf("expected first option label to contain '(default)', got %q", opts[0].Key)
	}
	want := border.Names()
	for i, o := range opts {
		if o.Value != want[i] {
			t.Errorf("option %d: expected value %q, got %q", i, want[i], o.Value)
		}
	}
}

func TestConfigSummaryIncludesFrame(t *testing.T) {
	frame := "single"
	cfg := &config.Config{Frame: &frame}
	out := configSummary(cfg)
	if !strings.Contains(out, "frame") {
		t.Errorf("expected summary to mention frame, got %q", out)
	}
	if !strings.Contains(out, "single") {
		t.Errorf("expected summary to contain 'single', got %q", out)
	}
}
