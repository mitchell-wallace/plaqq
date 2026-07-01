package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/mitchell-wallace/plaqq/internal/font"
	"github.com/mitchell-wallace/plaqq/internal/session"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newStyleFlagCmd builds a command whose styling flags are bound to the same
// package-level vars resolveStyle reads, so flag.Set both records the value and
// marks the flag as Changed (mirroring real CLI parsing).
func newStyleFlagCmd() *cobra.Command {
	flagColor, flagFont, flagBold, flagHint, flagNoHint, flagContinue = "", "", true, defaultHint, false, false
	c := &cobra.Command{Use: "test"}
	c.Flags().StringVar(&flagColor, "color", "", "")
	c.Flags().StringVar(&flagFont, "font", "", "")
	c.Flags().BoolVar(&flagBold, "bold", true, "")
	c.Flags().StringVar(&flagHint, "hint", defaultHint, "")
	c.Flags().BoolVar(&flagNoHint, "no-hint", false, "")
	c.Flags().BoolVarP(&flagContinue, "continue", "c", false, "")
	return c
}

func isolateStyleResolution(t *testing.T, configPath string) {
	t.Helper()
	t.Setenv("PLAQQ_CONFIG", configPath)
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	for _, name := range []string{envColor, envFont, envBold, envHint, envNoHint, envText} {
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

func optionValues[T comparable](opts []huh.Option[T]) []T {
	out := make([]T, len(opts))
	for i, o := range opts {
		out[i] = o.Value
	}
	return out
}

func contains[T comparable](slice []T, v T) bool {
	for _, x := range slice {
		if x == v {
			return true
		}
	}
	return false
}

func TestInteractiveFormAllowed(t *testing.T) {
	cases := []struct {
		name    string
		isTTY   bool
		jsonOut bool
		allowed bool
	}{
		{"tty plain", true, false, true},
		{"tty json", true, true, false},
		{"notty plain", false, false, false},
		{"notty json", false, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := interactiveFormAllowed(c.isTTY, c.jsonOut); got != c.allowed {
				t.Errorf("interactiveFormAllowed(%v,%v) = %v; want %v", c.isTTY, c.jsonOut, got, c.allowed)
			}
		})
	}
}

func TestSeedCustomiseFont(t *testing.T) {
	cases := []struct {
		resolved string
		want     string
	}{
		{"heavy", "heavy"},
		{"compact", "compact"},
		{"", font.DefaultName},
		{"   ", font.DefaultName},
		{"slant", font.DefaultName}, // unknown font falls back to default
		{"  HEAVY  ", "heavy"},      // trimmed and case-insensitive
	}
	for _, c := range cases {
		if got := seedCustomiseFont(c.resolved); got != c.want {
			t.Errorf("seedCustomiseFont(%q) = %q; want %q", c.resolved, got, c.want)
		}
	}
}

func TestModelEnterEditsSpaceDismisses(t *testing.T) {
	f := font.Get("block")
	color, _ := parseColor("info")
	m := initialModel("hello", f, color, true, defaultHint, true)

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should quit the TUI")
	}
	if got := updatedModel.(*model).action; got != modelActionEdit {
		t.Fatalf("enter action = %v; want edit", got)
	}

	m = initialModel("hello", f, color, true, defaultHint, true)
	updatedModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if cmd == nil {
		t.Fatal("space should quit the TUI")
	}
	if got := updatedModel.(*model).action; got != modelActionDismiss {
		t.Fatalf("space action = %v; want dismiss", got)
	}
}

func TestDismissedNoticeOutputIncludesReplayHint(t *testing.T) {
	got := dismissedNoticeOutput("deploy starting")
	want := "message: \"deploy starting\"\nRun `plaqq -c` to show again.\n"
	if got != want {
		t.Fatalf("dismissedNoticeOutput = %q; want %q", got, want)
	}
}

func TestSaveSessionTextOnlyDoesNotPersistRenderStyle(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	warnings := captureStyleWarnings(t)

	if err := saveSessionState(session.State{Text: "deploy starting"}); err != nil {
		t.Fatalf("saveSessionState text only: %v", err)
	}
	state, err := session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	if state.Text != "deploy starting" {
		t.Fatalf("text = %q; want deploy starting", state.Text)
	}
	if state.Font != "" || state.Color != "" {
		t.Fatalf("text-only save persisted style: %+v", state)
	}
	if warnings.Len() != 0 {
		t.Fatalf("unexpected warnings: %s", warnings.String())
	}
}

func TestSaveSessionStatePreservesExistingStyleWhenSavingText(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	if err := session.Save(session.State{Font: "heavy", Color: "alert"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	if err := saveSessionState(session.State{Text: "new text"}); err != nil {
		t.Fatalf("saveSessionState text only: %v", err)
	}
	state, err := session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	want := session.State{Font: "heavy", Color: "alert", Text: "new text"}
	if state != want {
		t.Fatalf("state = %+v; want %+v", state, want)
	}
}

func TestSaveSessionStateDropsInvalidExistingStyleWhenSavingText(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	if err := session.Save(session.State{Font: "slant", Color: "not-a-color"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}

	if err := saveSessionState(session.State{Text: "new text"}); err != nil {
		t.Fatalf("saveSessionState text only: %v", err)
	}
	state, err := session.Load(nil)
	if err != nil {
		t.Fatalf("session Load: %v", err)
	}
	want := session.State{Text: "new text"}
	if state != want {
		t.Fatalf("state = %+v; want %+v", state, want)
	}
}

func TestCustomiseColorOptions(t *testing.T) {
	// Empty resolved -> seeds to info, presets only, no extra option.
	opts, choice := customiseColorOptions("")
	if choice != "info" {
		t.Errorf("empty choice = %q; want info", choice)
	}
	if vals := optionValues(opts); !contains(vals, "info") || len(vals) != len(presetOrder) {
		t.Errorf("empty options = %v; want exactly the presets", vals)
	}

	// Preset resolved -> seeds to that preset, no extra option.
	opts, choice = customiseColorOptions("alert")
	if choice != "alert" {
		t.Errorf("preset choice = %q; want alert", choice)
	}
	if vals := optionValues(opts); contains(vals, "alert") == false || len(vals) != len(presetOrder) {
		t.Errorf("preset options = %v; want presets only", vals)
	}

	// Case-insensitive preset.
	_, choice = customiseColorOptions("FOCUS")
	if choice != "focus" {
		t.Errorf("uppercase preset choice = %q; want focus", choice)
	}

	// Custom hex resolved -> seeds to the value, value offered as a leading option.
	opts, choice = customiseColorOptions("#ff5f87")
	if choice != "#ff5f87" {
		t.Errorf("custom choice = %q; want #ff5f87", choice)
	}
	vals := optionValues(opts)
	if !contains(vals, "#ff5f87") {
		t.Errorf("custom options %v missing the resolved value", vals)
	}
	if len(vals) != len(presetOrder)+1 {
		t.Errorf("custom options = %v; want presets plus the custom value", vals)
	}
	if opts[0].Value != "#ff5f87" {
		t.Errorf("custom value should lead the options; first = %q", opts[0].Value)
	}

	// Invalid resolved -> seeds to info, no extra option.
	opts, choice = customiseColorOptions("not-a-color")
	if choice != "info" {
		t.Errorf("invalid choice = %q; want info", choice)
	}
	if vals := optionValues(opts); len(vals) != len(presetOrder) {
		t.Errorf("invalid options = %v; want presets only", vals)
	}
}

func TestConfirmTitle(t *testing.T) {
	if got := confirmTitle("heavy", "alert"); got != "Font: heavy | Colour: alert" {
		t.Fatalf("confirmTitle = %q", got)
	}
	if got := confirmTitle("", ""); got != "Font: compact | Colour: info" {
		t.Fatalf("confirmTitle fallback = %q", got)
	}
}

// TestCustomiseSessionRoundTrip simulates the customise step writing font+colour
// to the session store, then a later plaqq resolving style (no flags) picks them
// up — the "customise sticks for the session" persistence plumbing.
func TestCustomiseSessionRoundTrip(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	if err := session.Save(session.State{Font: "heavy", Color: "alert"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}
	cmd := newStyleFlagCmd()
	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if s.font != "heavy" || s.color != "alert" {
		t.Errorf("after customise save, style = {font:%q color:%q}; want heavy/alert", s.font, s.color)
	}
}

func TestCustomiseSessionRoundTripCustomColor(t *testing.T) {
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	if err := session.Save(session.State{Font: "compact", Color: "#ff5f87"}); err != nil {
		t.Fatalf("session Save: %v", err)
	}
	cmd := newStyleFlagCmd()
	s, err := resolveStyle(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if s.font != "compact" || s.color != "#ff5f87" {
		t.Errorf("after customise save, style = {font:%q color:%q}; want compact/#ff5f87", s.font, s.color)
	}
}

// runBareExit invokes the bare (no-message) root path and recovers the exit()
// panic, returning the captured stderr, the exit code, and whether the guard
// fired. It is the harness for the non-interactive guard tests.
func runBareExit(t *testing.T, setJSON bool) (stderr string, code int, exited bool) {
	t.Helper()
	isolateStyleResolution(t, filepath.Join(t.TempDir(), "none.toml"))
	if setJSON {
		jsonOutput = true
		t.Cleanup(func() { jsonOutput = false })
	}

	r, w, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = w
	cmd := newStyleFlagCmd()

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = rootCmd.RunE(cmd, nil)
	}()
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if ee, ok := recovered.(*exitError); ok {
		code = ee.code
		exited = true
	}
	return buf.String(), code, exited
}

func TestBareInvocationNonTTYExitsForMessage(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("cannot exercise the non-TTY guard when test stdin is a terminal")
	}
	stderr, code, exited := runBareExit(t, false)
	if !exited {
		t.Fatal("expected exit() panic for bare non-TTY invocation")
	}
	if code != 1 {
		t.Errorf("exit code = %d; want 1", code)
	}
	if !strings.Contains(stderr, "a message is required") {
		t.Errorf("stderr = %q; want it to contain 'a message is required'", stderr)
	}
}

func TestBareInvocationJSONOutputExitsForMessage(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("cannot exercise the non-TTY guard when test stdin is a terminal")
	}
	stderr, code, exited := runBareExit(t, true)
	if !exited {
		t.Fatal("expected exit() panic for bare --json-output invocation")
	}
	if code != 1 {
		t.Errorf("exit code = %d; want 1", code)
	}
	if !strings.Contains(stderr, `"error":"a message is required"`) {
		t.Errorf("stderr = %q; want JSON error containing 'a message is required'", stderr)
	}
}

func TestTUIScrolling(t *testing.T) {
	// Initialize font and color presets
	f := font.Get("block")
	color, _ := parseColor("info")

	// Create a model with a message long enough to wrap and exceed a small height.
	msg := "HELLO WORLD THIS IS A VERY LONG TEST MESSAGE THAT WILL EXCEED THE HEIGHT OF THE VIEWPORT"
	m := initialModel(msg, f, color, true, defaultHint, true)

	// Simulate window size message (width 40, height 10)
	m.width = 40
	m.height = 10

	// Get lines and check size
	lines := m.getScreenLines()
	if len(lines) <= m.height {
		t.Fatalf("expected content lines (%d) to exceed height (%d)", len(lines), m.height)
	}

	maxScroll := m.maxScrollOffset()
	if maxScroll <= 0 {
		t.Fatalf("expected maxScrollOffset (%d) to be > 0", maxScroll)
	}

	// 1. Scroll Down ('j')
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd != nil {
		t.Errorf("expected no command on scroll down key")
	}
	m2 := updatedModel.(*model)
	if m2.scrollOffset != 1 {
		t.Errorf("scrollOffset after 'j' = %d; want 1", m2.scrollOffset)
	}

	// 2. Scroll Up ('k')
	updatedModel, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m3 := updatedModel.(*model)
	if m3.scrollOffset != 0 {
		t.Errorf("scrollOffset after 'k' = %d; want 0", m3.scrollOffset)
	}

	// 3. Scroll Down (Arrow)
	updatedModel, _ = m3.Update(tea.KeyMsg{Type: tea.KeyDown, Runes: []rune{}})
	m4 := updatedModel.(*model)
	if m4.scrollOffset != 1 {
		t.Errorf("scrollOffset after 'down' = %d; want 1", m4.scrollOffset)
	}

	// 4. Scroll Up (Arrow)
	updatedModel, _ = m4.Update(tea.KeyMsg{Type: tea.KeyUp, Runes: []rune{}})
	m5 := updatedModel.(*model)
	if m5.scrollOffset != 0 {
		t.Errorf("scrollOffset after 'up' = %d; want 0", m5.scrollOffset)
	}

	// 5. Page Down
	updatedModel, _ = m5.Update(tea.KeyMsg{Type: tea.KeyPgDown, Runes: []rune{}})
	m6 := updatedModel.(*model)
	if m6.scrollOffset != m.height/2 {
		t.Errorf("scrollOffset after 'pgdown' = %d; want %d", m6.scrollOffset, m.height/2)
	}

	// 6. Page Up
	updatedModel, _ = m6.Update(tea.KeyMsg{Type: tea.KeyPgUp, Runes: []rune{}})
	m7 := updatedModel.(*model)
	if m7.scrollOffset != 0 {
		t.Errorf("scrollOffset after 'pgup' = %d; want 0", m7.scrollOffset)
	}

	// 7. Verify scrollbar elements in View
	view := m.View()
	// Since scrollbar is enabled, every line should have a scrollbar character at the end.
	if !strings.Contains(view, "░") && !strings.Contains(view, "█") {
		t.Errorf("expected view to render scrollbar characters ('░' or '█') in scrollable state")
	}
}
