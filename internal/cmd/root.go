package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchell-wallace/plaqq/internal/border"
	"github.com/mitchell-wallace/plaqq/internal/config"
	"github.com/mitchell-wallace/plaqq/internal/font"
	"github.com/mitchell-wallace/plaqq/internal/session"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const defaultHint = "[ Space: dismiss | Enter: edit ]"
const replayHint = "Run `plaqq -c` to show again."

const (
	envColor  = "PLAQQ_COLOR"
	envFont   = "PLAQQ_FONT"
	envFrame  = "PLAQQ_FRAME"
	envBold   = "PLAQQ_BOLD"
	envHint   = "PLAQQ_HINT"
	envNoHint = "PLAQQ_NO_HINT"
	envText   = "PLAQQ_TEXT"
)

var (
	version            string
	jsonOutput         bool
	styleWarningOutput io.Writer = os.Stderr

	// Styling flags for the notice display.
	flagColor    string
	flagFont     string
	flagFrame    string
	flagBold     bool
	flagHint     string
	flagNoHint   bool
	flagContinue bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json-output", false, "emit structured JSON output")
	rootCmd.Flags().StringVar(&flagColor, "color", "", "notice color: a preset name ("+strings.Join(presetOrder, ", ")+"), a hex code (e.g. #00f5d4), or an ANSI index (0-255); defaults to info")
	rootCmd.Flags().StringVar(&flagFont, "font", "", "font to render the notice in; one of: "+strings.Join(font.Names(), ", "))
	rootCmd.Flags().StringVar(&flagFrame, "frame", "", "frame around the notice: one of: "+strings.Join(border.Names(), ", "))
	rootCmd.Flags().BoolVar(&flagBold, "bold", true, "render the notice text in bold")
	rootCmd.Flags().StringVar(&flagHint, "hint", defaultHint, "action-hint text shown beneath the notice")
	rootCmd.Flags().BoolVar(&flagNoHint, "no-hint", false, "hide the dismiss hint")
	rootCmd.Flags().BoolVarP(&flagContinue, "continue", "c", false, "open the last notice from this terminal session")
}

// parseColor validates a user-supplied color string and returns the matching
// lipgloss color. It accepts a preset name, a hex code (#rgb or #rrggbb), or an
// ANSI index (0-255).
func parseColor(s string) (lipgloss.TerminalColor, error) {
	return parseColorFromSource(s, styleValueSource{})
}

func parseColorFromSource(s string, source styleValueSource) (lipgloss.TerminalColor, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("color is empty")
	}

	if c, ok := presetColor(s); ok {
		return c, nil
	}

	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		if len(hex) != 3 && len(hex) != 6 {
			return nil, styleValueError(source, "invalid hex color %q: expected #rgb or #rrggbb", s)
		}
		for _, r := range hex {
			if !isHexDigit(r) {
				return nil, styleValueError(source, "invalid hex color %q", s)
			}
		}
		return lipgloss.Color(s), nil
	}

	if n, err := strconv.Atoi(s); err == nil {
		if n < 0 || n > 255 {
			return nil, styleValueError(source, "invalid ANSI color %d: expected an index 0-255", n)
		}
		return lipgloss.Color(s), nil
	}

	return nil, unknownStyleValueError("color", s, source, presetOrder, "preset names", "; hex codes like #00f5d4 and ANSI indexes 0-255 are also valid")
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

type invalidStyleValueAction int

const (
	invalidStyleValueError invalidStyleValueAction = iota
	invalidStyleValueWarn
)

type styleValueSource struct {
	label         string
	invalidAction invalidStyleValueAction
}

var configStyleSource = styleValueSource{
	label:         "config file",
	invalidAction: invalidStyleValueError,
}

func envStyleSource(name string) styleValueSource {
	return styleValueSource{
		label:         name + " environment variable",
		invalidAction: invalidStyleValueWarn,
	}
}

func flagStyleSource(name string) styleValueSource {
	return styleValueSource{
		label:         "--" + name + " flag",
		invalidAction: invalidStyleValueError,
	}
}

func sessionStyleSource(field string) styleValueSource {
	return styleValueSource{
		label:         "session state " + field,
		invalidAction: invalidStyleValueWarn,
	}
}

func (src styleValueSource) handleInvalid(err error) error {
	switch src.invalidAction {
	case invalidStyleValueWarn:
		warnStyleValue(err)
		return nil
	default:
		return err
	}
}

func warnStyleValue(err error) {
	if styleWarningOutput == nil {
		return
	}
	_, _ = fmt.Fprintf(styleWarningOutput, "plaqq: warning: %v; ignoring value\n", err)
}

func styleValueError(source styleValueSource, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	if source.label == "" {
		return fmt.Errorf("%s", msg)
	}
	return fmt.Errorf("%s from %s", msg, source.label)
}

func unknownStyleValueError(kind, value string, source styleValueSource, options []string, optionLabel, suffix string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "unknown %s %q", kind, value)
	if source.label != "" {
		fmt.Fprintf(&b, " from %s", source.label)
	}
	fmt.Fprintf(&b, ": valid %s are %s", optionLabel, strings.Join(options, ", "))
	b.WriteString(suffix)
	if suggestion, ok := suggestName(value, options); ok {
		fmt.Fprintf(&b, "; did you mean %q?", suggestion)
	}
	return fmt.Errorf("%s", b.String())
}

func suggestName(input string, options []string) (string, bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return "", false
	}

	best := ""
	bestDistance := 0
	tied := false
	for _, option := range options {
		d := editDistance(input, strings.ToLower(option))
		if best == "" || d < bestDistance {
			best = option
			bestDistance = d
			tied = false
		} else if d == bestDistance {
			tied = true
		}
	}
	if best == "" || tied {
		return "", false
	}

	limit := 2
	if len([]rune(input)) > 6 || len([]rune(best)) > 6 {
		limit = 3
	}
	if bestDistance > limit {
		return "", false
	}
	return best, true
}

func editDistance(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}

	for i, ca := range ar {
		curr := make([]int, len(br)+1)
		curr[0] = i + 1
		for j, cb := range br {
			cost := 0
			if ca != cb {
				cost = 1
			}
			curr[j+1] = minInt(
				curr[j]+1,
				prev[j+1]+1,
				prev[j]+cost,
			)
		}
		prev = curr
	}
	return prev[len(br)]
}

func minInt(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

// styleSettings is the effective notice styling after layering defaults, the
// config file, environment variables, session state, and CLI flags.
type styleSettings struct {
	color  string
	font   string
	frame  string
	bold   bool
	hint   string
	noHint bool
}

// resolveStyle determines the effective notice styling by layering, in order:
// built-in defaults, values from the config file, environment variables,
// session state, then any CLI flags the user explicitly set.
func resolveStyle(cmd *cobra.Command) (styleSettings, error) {
	s := styleSettings{color: "", font: "", frame: "", bold: true, hint: defaultHint, noHint: false}

	path, err := config.Path()
	if err != nil {
		return styleSettings{}, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return styleSettings{}, err
	}
	if cfg.Color != nil {
		if err := applyColorLayer(&s, *cfg.Color, configStyleSource); err != nil {
			return styleSettings{}, err
		}
	}
	if cfg.Font != nil {
		if err := applyFontLayer(&s, *cfg.Font, configStyleSource); err != nil {
			return styleSettings{}, err
		}
	}
	if cfg.Frame != nil {
		if err := applyFrameLayer(&s, *cfg.Frame, configStyleSource); err != nil {
			return styleSettings{}, err
		}
	}
	if cfg.Bold != nil {
		s.bold = *cfg.Bold
	}
	if cfg.Hint != nil {
		s.hint = *cfg.Hint
	}
	if cfg.NoHint != nil {
		s.noHint = *cfg.NoHint
	}

	if err := applyEnvLayer(&s); err != nil {
		return styleSettings{}, err
	}

	state, err := session.Load(styleWarningOutput)
	if err != nil {
		warnStyleValue(styleValueError(sessionStyleSource("record"), "%v", err))
	} else {
		if state.Color != "" {
			if err := applyColorLayer(&s, state.Color, sessionStyleSource("color")); err != nil {
				return styleSettings{}, err
			}
		}
		if state.Font != "" {
			if err := applyFontLayer(&s, state.Font, sessionStyleSource("font")); err != nil {
				return styleSettings{}, err
			}
		}
		if state.Frame != "" {
			if err := applyFrameLayer(&s, state.Frame, sessionStyleSource("frame")); err != nil {
				return styleSettings{}, err
			}
		}
	}

	if cmd.Flags().Changed("color") {
		if err := applyColorLayer(&s, flagColor, flagStyleSource("color")); err != nil {
			return styleSettings{}, err
		}
	}
	if cmd.Flags().Changed("font") {
		if err := applyFontLayer(&s, flagFont, flagStyleSource("font")); err != nil {
			return styleSettings{}, err
		}
	}
	if cmd.Flags().Changed("frame") {
		if err := applyFrameLayer(&s, flagFrame, flagStyleSource("frame")); err != nil {
			return styleSettings{}, err
		}
	}
	if cmd.Flags().Changed("bold") {
		s.bold = flagBold
	}
	if cmd.Flags().Changed("hint") {
		s.hint = flagHint
	}
	if cmd.Flags().Changed("no-hint") {
		s.noHint = flagNoHint
	}

	return s, nil
}

func applyEnvLayer(s *styleSettings) error {
	if value, ok := os.LookupEnv(envColor); ok && value != "" {
		if err := applyColorLayer(s, value, envStyleSource(envColor)); err != nil {
			return err
		}
	}
	if value, ok := os.LookupEnv(envFont); ok && value != "" {
		if err := applyFontLayer(s, value, envStyleSource(envFont)); err != nil {
			return err
		}
	}
	if value, ok := os.LookupEnv(envFrame); ok && value != "" {
		if err := applyFrameLayer(s, value, envStyleSource(envFrame)); err != nil {
			return err
		}
	}
	if err := applyBoolEnvLayer(envBold, func(parsed bool) {
		s.bold = parsed
	}); err != nil {
		return err
	}
	if value, ok := os.LookupEnv(envHint); ok && value != "" {
		s.hint = value
	}
	if err := applyBoolEnvLayer(envNoHint, func(parsed bool) {
		s.noHint = parsed
	}); err != nil {
		return err
	}
	return nil
}

func applyBoolEnvLayer(name string, set func(bool)) error {
	if value, ok := os.LookupEnv(name); ok && value != "" {
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err == nil {
			set(parsed)
			return nil
		}
		source := envStyleSource(name)
		if err := source.handleInvalid(styleValueError(source, "invalid boolean %q", value)); err != nil {
			return err
		}
	}
	return nil
}

func applyColorLayer(s *styleSettings, value string, source styleValueSource) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return source.handleInvalid(styleValueError(source, "color is empty"))
	}
	if _, err := parseColorFromSource(value, source); err != nil {
		return source.handleInvalid(err)
	}
	s.color = value
	return nil
}

func applyFontLayer(s *styleSettings, value string, source styleValueSource) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return source.handleInvalid(styleValueError(source, "font is empty"))
	}
	if !font.Has(value) {
		return source.handleInvalid(unknownStyleValueError("font", value, source, font.Names(), "fonts", ""))
	}
	s.font = value
	return nil
}

func applyFrameLayer(s *styleSettings, value string, source styleValueSource) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return source.handleInvalid(styleValueError(source, "frame is empty"))
	}
	if _, err := border.Parse(value); err != nil {
		return source.handleInvalid(unknownStyleValueError("frame", value, source, border.Names(), "frames", ""))
	}
	s.frame = value
	return nil
}

// interactiveFormAllowed reports whether the bare-invocation confirm/customise
// form can run: it needs an interactive stdin TTY and must not be in
// --json-output mode (the form, like the old huh.NewInput prompt, cannot render
// without a TTY).
func interactiveFormAllowed(isTTY, jsonOut bool) bool {
	return isTTY && !jsonOut
}

// seedCustomiseFont returns the font select choice seeded from the resolved
// style, defaulting to the block font when the style carries no valid font. The
// result is lower-cased so it matches the canonical font option values.
func seedCustomiseFont(resolved string) string {
	if v := strings.ToLower(strings.TrimSpace(resolved)); v != "" && font.Has(v) {
		return v
	}
	return font.DefaultName
}

func seedCustomiseFrame(resolved string) string {
	if v := strings.ToLower(strings.TrimSpace(resolved)); v != "" && border.Has(v) {
		return v
	}
	return border.DefaultName
}

// customiseColorOptions builds the colour options for the customise select,
// seeded from the resolved style colour. The semantic presets are always
// offered; a resolved custom (hex/ANSI) colour is included as a leading option
// so the current value stays represented and selectable. It returns the options
// and the seeded choice (which always matches one of the option values).
func customiseColorOptions(resolved string) ([]huh.Option[string], string) {
	resolved = strings.TrimSpace(resolved)
	choice := "info"
	if resolved != "" {
		if _, ok := presetColor(resolved); ok {
			choice = strings.ToLower(resolved)
		} else if _, err := parseColor(resolved); err == nil {
			choice = resolved
		}
	}

	opts := make([]huh.Option[string], 0, len(presetOrder)+1)
	if resolved != "" {
		if _, ok := presetColor(resolved); !ok {
			if _, err := parseColor(resolved); err == nil {
				opts = append(opts, huh.NewOption(resolved, resolved))
			}
		}
	}
	for _, name := range presetOrder {
		opts = append(opts, huh.NewOption(name, name))
	}
	return opts, choice
}

// interactiveFormResult holds the outcome of the confirm/customise form.
type interactiveFormResult struct {
	message     string
	fontChoice  string
	colorChoice string
	frameChoice string
	customised  bool
}

// runInteractiveForm presents the confirm/customise form for a bare invocation
// (no message argument) and returns the chosen message and style.
//
// Group 1 holds the message text area plus Confirm/Customise buttons (Confirm
// selected, so the fast path is enter, enter). Group 2 holds the font and colour
// selects and is shown via WithHideFunc only when Customise is chosen, so the
// user can shift+tab back to edit the message. Esc and Ctrl-C abort at any step
// (returned as huh.ErrUserAborted). Each select is seeded from the fully-resolved
// style.
func runInteractiveForm(style styleSettings, initialMessage string) (interactiveFormResult, error) {
	message := strings.TrimSpace(initialMessage)
	confirm := true
	fontChoice := seedCustomiseFont(style.font)
	colorOpts, colorChoice := customiseColorOptions(style.color)
	frameChoice := seedCustomiseFrame(style.frame)

	// Bind both Esc and Ctrl-C to quit so the abort affordance is consistent at
	// every step (huh only binds Ctrl-C by default).
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("esc/ctrl+c", "quit"))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Notice message").
				Placeholder("e.g. remember to run e2e tests before pushing").
				Lines(3).
				ExternalEditor(false).
				Value(&message).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("a message is required")
					}
					return nil
				}),
			huh.NewConfirm().
				TitleFunc(func() string {
					return confirmTitle(fontChoice, colorChoice, frameChoice)
				}, []any{&fontChoice, &colorChoice, &frameChoice}).
				Affirmative("Confirm").
				Negative("Customise").
				Value(&confirm),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Font").
				Options(fontOptions()...).
				Value(&fontChoice),
			huh.NewSelect[string]().
				Title("Color").
				Options(colorOpts...).
				Value(&colorChoice),
			huh.NewSelect[string]().
				Title("Frame").
				Options(frameOptions()...).
				Value(&frameChoice),
		).WithHideFunc(func() bool { return confirm }),
	).WithKeyMap(keymap)

	if err := form.Run(); err != nil {
		return interactiveFormResult{}, err
	}

	return interactiveFormResult{
		message:     strings.TrimSpace(message),
		fontChoice:  fontChoice,
		colorChoice: colorChoice,
		frameChoice: frameChoice,
		customised:  !confirm,
	}, nil
}

func confirmTitle(fontName, colorName, frameName string) string {
	return fmt.Sprintf("Font: %s | Colour: %s | Frame: %s",
		displayStyleName(fontName, font.DefaultName),
		displayStyleName(colorName, "info"),
		displayStyleName(frameName, border.DefaultName),
	)
}

func displayStyleName(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

var rootCmd = &cobra.Command{
	Use:   "plaqq [message]",
	Short: "plaqq displays a notice across the terminal pane in chunky letters",
	Long: `plaqq takes a message and displays it in a large chunky font centered on the
terminal. Space dismisses the notice; Enter returns to the prompt so the message,
font, or color can be edited and shown again.

The font, color, bold weight, and dismiss hint can be customized with the
--font, --color, --bold, --hint, and --no-hint flags, or set as persistent
defaults via 'plaqq config' (an interactive picker). Flags override the config.

Colors accept a preset name (info, alert, warn, ok, focus), a hex code, or an ANSI index. Fonts include Unicode
block faces (block, heavy, compact, wide).`,
	Example: `  plaqq "deploy starting"
  plaqq --color alert "build failed"
  plaqq --font heavy --color "#ff5f87" "build failed"
  plaqq --font compact --bold=false "heads up"
  plaqq --hint "press space to continue" "meeting in 5"
  plaqq --no-hint "stand clear"
  plaqq --continue`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve styling with precedence: built-in defaults < config file <
		// environment variables < session state < CLI flags.
		style, err := resolveStyle(cmd)
		if err != nil {
			return err
		}

		state, err := session.Load(styleWarningOutput)
		if err != nil {
			warnStyleValue(styleValueError(sessionStyleSource("record"), "%v", err))
		}

		// Render font/colour start from the resolved style and may be overridden
		// by the customise step below.
		renderFont := style.font
		renderColor := style.color
		renderFrame := style.frame
		persistStyle := false

		var noticeMsg string
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			// Message-argument path: render immediately, no confirm/customise step.
			noticeMsg = args[0]
		} else if flagContinue {
			noticeMsg = strings.TrimSpace(os.Getenv(envText))
			if noticeMsg == "" {
				noticeMsg = state.Text
			}
			if noticeMsg == "" {
				exit(1, "no previous message for this terminal session")
			}
		} else {
			// The confirm/customise form needs an interactive TTY. When stdin is
			// piped/CI or --json-output is set, refuse to launch it and ask for a
			// message argument instead. This also fixes a latent huh.NewInput hang
			// under --json-output.
			if !interactiveFormAllowed(term.IsTerminal(int(os.Stdin.Fd())), jsonOutput) {
				exit(1, "a message is required")
			}
			res, err := runInteractiveForm(style, "")
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					// Esc/Ctrl-C at any step: clean exit, no notice.
					return nil
				}
				return err
			}
			noticeMsg = res.message
			if res.customised {
				renderFont = res.fontChoice
				renderColor = res.colorChoice
				renderFrame = res.frameChoice
				persistStyle = true
			}
		}

		for {
			nextState := session.State{Text: noticeMsg}
			if persistStyle {
				nextState.Font = renderFont
				nextState.Color = renderColor
				nextState.Frame = renderFrame
			}
			if err := saveSessionState(nextState); err != nil {
				warnStyleValue(fmt.Errorf("could not save session state: %w", err))
			}

			glyphFont := font.Get(renderFont)

			var noticeColor lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#00d7af", Dark: "#00f5d4"}
			if renderColor != "" {
				c, err := parseColor(renderColor)
				if err != nil {
					return err
				}
				noticeColor = c
			}

			frameKind := border.Kind(strings.ToLower(strings.TrimSpace(renderFrame)))
			if frameKind == "" {
				frameKind = border.None
			}

			updateNoticeChan := make(chan string, 1)
			startUpdateCheck(updateNoticeChan)

			m := initialModel(noticeMsg, glyphFont, noticeColor, frameKind, style.bold, style.hint, !style.noHint)
			p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithoutCatchPanics())
			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("run program: %w", err)
			}

			select {
			case notice := <-updateNoticeChan:
				fmt.Fprintln(os.Stderr, notice)
			default:
			}

			if final, ok := finalModel.(*model); !ok || final.action != modelActionEdit {
				fmt.Print(dismissedNoticeOutput(noticeMsg))
				return nil
			}

			if !interactiveFormAllowed(term.IsTerminal(int(os.Stdin.Fd())), jsonOutput) {
				exit(1, "cannot edit without an interactive terminal")
			}
			res, err := runInteractiveForm(styleSettings{color: renderColor, font: renderFont, frame: renderFrame, bold: style.bold, hint: style.hint, noHint: style.noHint}, noticeMsg)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					return nil
				}
				return err
			}
			noticeMsg = res.message
			if res.customised {
				renderFont = res.fontChoice
				renderColor = res.colorChoice
				renderFrame = res.frameChoice
				persistStyle = true
			}
		}
	},
}

func startUpdateCheck(updateNoticeChan chan<- string) {
	if jsonOutput || !isReleaseVersion(version) {
		return
	}
	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		req, err := http.NewRequest("GET", "https://api.github.com/repos/mitchell-wallace/plaqq/releases/latest", nil)
		if err != nil {
			return
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		} else if token := os.Getenv("GH_TOKEN"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return
		}
		var payload struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			latest := strings.TrimPrefix(payload.TagName, "v")
			cmp, err := compareVersions(version, latest)
			if err == nil && cmp < 0 {
				updateNoticeChan <- fmt.Sprintf("\n✨ A new version of plaqq is available: v%s (current: v%s). Run 'plaqq update' to update.", latest, version)
			}
		}
	}()
}

func dismissedNoticeOutput(message string) string {
	return fmt.Sprintf("message: %q\n%s\n", message, replayHint)
}

func saveSessionState(next session.State) error {
	current, err := session.Load(styleWarningOutput)
	if err != nil {
		return err
	}
	if strings.TrimSpace(next.Color) == "" {
		if _, err := parseColor(current.Color); err == nil {
			next.Color = current.Color
		}
	}
	if strings.TrimSpace(next.Font) == "" {
		if font.Has(current.Font) {
			next.Font = current.Font
		}
	}
	if strings.TrimSpace(next.Frame) == "" {
		if border.Has(current.Frame) {
			next.Frame = current.Frame
		}
	}
	if strings.TrimSpace(next.Text) == "" {
		next.Text = current.Text
	}
	return session.Save(next)
}

type modelAction int

const (
	modelActionDismiss modelAction = iota
	modelActionEdit
)

type model struct {
	text         string
	hint         string
	font         font.Font
	color        lipgloss.TerminalColor
	frame        border.Kind
	width        int
	height       int
	bold         bool
	showHint     bool
	scrollOffset int
	action       modelAction
}

func initialModel(text string, glyphFont font.Font, color lipgloss.TerminalColor, frame border.Kind, bold bool, hint string, showHint bool) model {
	return model{
		text:     text,
		font:     glyphFont,
		color:    color,
		frame:    frame,
		bold:     bold,
		hint:     hint,
		showHint: showHint,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.HideCursor,
		tea.SetWindowTitle(m.text),
	)
}

func frameMarginFor(fontName string) int {
	switch fontName {
	case "heavy":
		return 3
	case "compact", "block", "wide":
		return 2
	case "terminal":
		return 1
	default:
		return 2
	}
}

func gapRowsFor(f font.Font) int {
	if _, ok := f.(font.Terminal); ok {
		return 1
	}
	return 2
}

func fontNameFor(f font.Font) string {
	if _, ok := f.(font.Terminal); ok {
		return "terminal"
	}
	for _, name := range font.Names() {
		if font.Get(name) == f {
			return name
		}
	}
	return font.DefaultName
}

func (m *model) getScreenLines() []string {
	if m.width == 0 || m.height == 0 {
		return nil
	}

	// Dynamic word wrapping with a margin
	maxWidth := m.width - 8
	if maxWidth < 10 {
		maxWidth = 10
	}

	lines := font.Wrap(m.font, m.text, maxWidth)

	var rows []string
	for idx, line := range lines {
		if idx > 0 {
			gap := gapRowsFor(m.font)
			for i := 0; i < gap; i++ {
				rows = append(rows, "")
			}
		}
		for _, row := range m.font.Render(line) {
			if strings.TrimSpace(row) == "" {
				rows = append(rows, "")
				continue
			}
			rows = append(rows, row)
		}
	}

	if m.frame != border.None {
		fontName := fontNameFor(m.font)
		overflowing := len(rows) > m.height
		buffer := 0
		if overflowing {
			buffer = 5
		}
		framed := m.frame.Wrap(rows, frameMarginFor(fontName))
		available := m.width - buffer
		fits := true
		for _, row := range framed {
			if utf8.RuneCountInString(row) > available {
				fits = false
				break
			}
		}
		inner := 0
		if len(framed) > 2 {
			inner = utf8.RuneCountInString(framed[1]) - 2
		}
		if fits && inner >= 2 {
			rows = framed
		}
	}

	maxRowWidth := 0
	for _, row := range rows {
		if w := utf8.RuneCountInString(row); w > maxRowWidth {
			maxRowWidth = w
		}
	}
	padding := (m.width - maxRowWidth) / 2
	if padding < 0 {
		padding = 0
	}
	leftPad := strings.Repeat(" ", padding)

	noticeStyle := lipgloss.NewStyle().Foreground(m.color).Bold(m.bold)
	var centeredRows []string
	for _, row := range rows {
		if strings.TrimSpace(row) == "" {
			centeredRows = append(centeredRows, "")
			continue
		}
		centeredRows = append(centeredRows, leftPad+noticeStyle.Render(row))
	}

	showHint := m.showHint && strings.TrimSpace(m.hint) != ""

	// Stylize and center dismissal hint
	var centeredHint string
	if showHint {
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "242"}).
			Italic(true)

		hintPadding := (m.width - len([]rune(m.hint))) / 2
		if hintPadding < 0 {
			hintPadding = 0
		}
		centeredHint = strings.Repeat(" ", hintPadding) + hintStyle.Render(m.hint)
	}

	var contentLines []string
	contentLines = append(contentLines, centeredRows...)
	if showHint {
		contentLines = append(contentLines, "", "", centeredHint)
	}

	if len(contentLines) < m.height {
		// Center everything vertically
		topPadding := (m.height - len(contentLines)) / 2
		bottomPadding := m.height - topPadding - len(contentLines)

		var screenLines []string
		for i := 0; i < topPadding; i++ {
			screenLines = append(screenLines, "")
		}
		screenLines = append(screenLines, contentLines...)
		for i := 0; i < bottomPadding; i++ {
			screenLines = append(screenLines, "")
		}
		return screenLines
	}

	return contentLines
}

func (m *model) maxScrollOffset() int {
	lines := m.getScreenLines()
	if len(lines) == 0 {
		return 0
	}
	maxOffset := len(lines) - m.height
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.action = modelActionEdit
			return m, tea.Quit
		case " ", "esc", "ctrl+c", "q":
			m.action = modelActionDismiss
			return m, tea.Quit
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j":
			if m.scrollOffset < m.maxScrollOffset() {
				m.scrollOffset++
			}
		case "pgup":
			m.scrollOffset -= m.height / 2
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
		case "pgdown":
			m.scrollOffset += m.height / 2
			maxOffset := m.maxScrollOffset()
			if m.scrollOffset > maxOffset {
				m.scrollOffset = maxOffset
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		maxOffset := m.maxScrollOffset()
		if m.scrollOffset > maxOffset {
			m.scrollOffset = maxOffset
		}
	}
	return m, nil
}

func (m *model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	screenLines := m.getScreenLines()
	if len(screenLines) == 0 {
		return ""
	}

	start := m.scrollOffset
	if start > len(screenLines)-m.height {
		start = len(screenLines) - m.height
	}
	if start < 0 {
		start = 0
	}

	isScrollable := len(screenLines) > m.height
	var visibleLines []string

	thumbStyle := lipgloss.NewStyle().Foreground(m.color)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})

	for i := 0; i < m.height; i++ {
		lineIdx := start + i
		if lineIdx >= len(screenLines) {
			break
		}
		line := screenLines[lineIdx]
		if isScrollable {
			thumbHeight := m.height * m.height / len(screenLines)
			if thumbHeight < 1 {
				thumbHeight = 1
			}
			scrollableThumbRange := m.height - thumbHeight
			scrollableOffsetRange := len(screenLines) - m.height

			thumbStart := 0
			if scrollableOffsetRange > 0 {
				thumbStart = start * scrollableThumbRange / scrollableOffsetRange
			}

			isThumb := i >= thumbStart && i < thumbStart+thumbHeight
			var sbChar string
			if isThumb {
				sbChar = thumbStyle.Render("█")
			} else {
				sbChar = trackStyle.Render("░")
			}

			w := lipgloss.Width(line)
			paddingNeeded := m.width - 1 - w
			if paddingNeeded < 0 {
				paddingNeeded = 0
			}
			line = line + strings.Repeat(" ", paddingNeeded) + sbChar
		}
		visibleLines = append(visibleLines, line)
	}

	return strings.Join(visibleLines, "\n")
}

type exitError struct {
	code int
}

func (e *exitError) Error() string { return fmt.Sprintf("exit %d", e.code) }

func exit(code int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if jsonOutput {
		b, _ := json.Marshal(map[string]any{
			"error":    msg,
			"exitCode": code,
		})
		fmt.Fprintln(os.Stderr, string(b))
	} else {
		fmt.Fprintf(os.Stderr, "plaqq: %s\n", msg)
	}
	panic(&exitError{code: code})
}

// Execute parses commands and executes rootCmd
func Execute(v string) error {
	version = v
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	defer func() {
		if r := recover(); r != nil {
			if ee, ok := r.(*exitError); ok {
				os.Exit(ee.code)
			}
			panic(r)
		}
	}()

	// Detect JSON output flag early
	for _, a := range os.Args[1:] {
		if a == "--json-output" {
			jsonOutput = true
		}
	}

	// Version flag interception
	for _, a := range os.Args[1:] {
		if a == "--" {
			break
		}
		if a == "--version" || a == "-v" {
			if jsonOutput {
				printJSON(map[string]any{"version": version})
			} else {
				fmt.Println(version)
			}
			return nil
		}
	}

	err := rootCmd.Execute()
	if err != nil {
		exit(1, "%v", err)
	}
	return nil
}

func printJSON(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		exit(2, "json marshal: %v", err)
	}
	fmt.Println(string(b))
}
