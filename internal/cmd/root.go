package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchell-wallace/plaqq/internal/config"
	"github.com/mitchell-wallace/plaqq/internal/font"
	"github.com/spf13/cobra"
)

const defaultHint = "[ Press Space to dismiss ]"

var (
	version    string
	jsonOutput bool

	// Styling flags for the notice display.
	flagColor  string
	flagFont   string
	flagBold   bool
	flagHint   string
	flagNoHint bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json-output", false, "emit structured JSON output")
	rootCmd.Flags().StringVar(&flagColor, "color", "", "notice color: a preset name ("+strings.Join(presetOrder, ", ")+"), a hex code (e.g. #00f5d4), or an ANSI index (0-255); defaults to info")
	rootCmd.Flags().StringVar(&flagFont, "font", "", "font to render the notice in; one of: "+strings.Join(font.Names(), ", "))
	rootCmd.Flags().BoolVar(&flagBold, "bold", true, "render the notice text in bold")
	rootCmd.Flags().StringVar(&flagHint, "hint", defaultHint, "dismiss-hint text shown beneath the notice")
	rootCmd.Flags().BoolVar(&flagNoHint, "no-hint", false, "hide the dismiss hint")
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

func flagStyleSource(name string) styleValueSource {
	return styleValueSource{
		label:         "--" + name + " flag",
		invalidAction: invalidStyleValueError,
	}
}

func (src styleValueSource) handleInvalid(err error) error {
	switch src.invalidAction {
	case invalidStyleValueWarn:
		// Future env/session layers use this branch to warn and fall through.
		return nil
	default:
		return err
	}
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
// config file, and CLI flags.
type styleSettings struct {
	color  string
	font   string
	bold   bool
	hint   string
	noHint bool
}

// resolveStyle determines the effective notice styling by layering, in order:
// built-in defaults, values from the config file, then any CLI flags the user
// explicitly set (so a flag always wins over the config file).
func resolveStyle(cmd *cobra.Command) (styleSettings, error) {
	s := styleSettings{color: "", font: "", bold: true, hint: defaultHint, noHint: false}

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
	if cfg.Bold != nil {
		s.bold = *cfg.Bold
	}
	if cfg.Hint != nil {
		s.hint = *cfg.Hint
	}
	if cfg.NoHint != nil {
		s.noHint = *cfg.NoHint
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

func applyColorLayer(s *styleSettings, value string, source styleValueSource) error {
	value = strings.TrimSpace(value)
	if value == "" {
		s.color = ""
		return nil
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
		s.font = ""
		return nil
	}
	if !font.Has(value) {
		return source.handleInvalid(unknownStyleValueError("font", value, source, font.Names(), "fonts", ""))
	}
	s.font = value
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "plaqq [message]",
	Short: "plaqq displays a notice across the terminal pane in chunky letters",
	Long: `plaqq takes a message and displays it in a large chunky font centered on the
terminal. It waits for the user to press the spacebar to dismiss the notice.

The font, color, bold weight, and dismiss hint can be customized with the
--font, --color, --bold, --hint, and --no-hint flags, or set as persistent
defaults via 'plaqq config' (an interactive picker). Flags override the config.

Colors accept a preset name (info, alert, warn, ok, focus), a hex code, or an ANSI index. Fonts include Unicode
block faces (block, heavy, compact).`,
	Example: `  plaqq "deploy starting"
  plaqq --color alert "build failed"
  plaqq --font heavy --color "#ff5f87" "build failed"
  plaqq --font compact --bold=false "heads up"
  plaqq --hint "press space to continue" "meeting in 5"
  plaqq --no-hint "stand clear"`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve styling with precedence: built-in defaults < config file < CLI flags.
		style, err := resolveStyle(cmd)
		if err != nil {
			return err
		}

		glyphFont := font.Get(style.font)

		var noticeColor lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#00d7af", Dark: "#00f5d4"}
		if style.color != "" {
			c, err := parseColor(style.color)
			if err != nil {
				return err
			}
			noticeColor = c
		}

		var noticeMsg string
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			noticeMsg = args[0]
		} else {
			err := huh.NewInput().
				Title("Enter notice message").
				Placeholder("e.g. remember to run e2e tests before pushing").
				Value(&noticeMsg).
				Run()
			if err != nil {
				// Exit cleanly on abort
				return nil
			}
			noticeMsg = strings.TrimSpace(noticeMsg)
			if noticeMsg == "" {
				return fmt.Errorf("a message is required")
			}
		}

		// Start update check in background
		updateNoticeChan := make(chan string, 1)
		if !jsonOutput && isReleaseVersion(version) {
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

		// Initialize Bubble Tea program
		m := initialModel(noticeMsg, glyphFont, noticeColor, style.bold, style.hint, !style.noHint)
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithoutCatchPanics())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("run program: %w", err)
		}

		// Echo the message to scrollback so it survives after the alt-screen
		// is torn down (otherwise an interactively entered notice leaves no trace).
		fmt.Printf("message: %q\n", noticeMsg)

		// Print update notice if one was detected
		select {
		case notice := <-updateNoticeChan:
			fmt.Fprintln(os.Stderr, notice)
		default:
		}

		return nil
	},
}

type model struct {
	text     string
	hint     string
	font     font.Font
	color    lipgloss.TerminalColor
	width    int
	height   int
	bold     bool
	showHint bool
}

func initialModel(text string, glyphFont font.Font, color lipgloss.TerminalColor, bold bool, hint string, showHint bool) model {
	return model{
		text:     text,
		font:     glyphFont,
		color:    color,
		bold:     bold,
		hint:     hint,
		showHint: showHint,
	}
}

func (m model) Init() tea.Cmd {
	return tea.HideCursor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ", "enter", "esc", "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Dynamic word wrapping with a margin
	maxWidth := m.width - 8
	if maxWidth < 10 {
		maxWidth = 10
	}

	lines := font.Wrap(m.font, m.text, maxWidth)

	noticeStyle := lipgloss.NewStyle().Foreground(m.color).Bold(m.bold)

	// Render each wrapped line as a block and center it. Each line is padded by
	// one uniform amount so that the rendered block keeps its columns aligned
	// rather than shearing row by row.
	var centeredRows []string
	for idx, line := range lines {
		if idx > 0 {
			// Blank spacing rows between wrapped lines.
			centeredRows = append(centeredRows, "", "")
		}

		rows := m.font.Render(line)
		lineWidth := 0
		for _, row := range rows {
			if w := len([]rune(row)); w > lineWidth {
				lineWidth = w
			}
		}
		padding := (m.width - lineWidth) / 2
		if padding < 0 {
			padding = 0
		}
		leftPad := strings.Repeat(" ", padding)

		for _, row := range rows {
			if strings.TrimSpace(row) == "" {
				centeredRows = append(centeredRows, "")
				continue
			}
			centeredRows = append(centeredRows, leftPad+noticeStyle.Render(row))
		}
	}

	chunkyBlock := strings.Join(centeredRows, "\n")

	showHint := m.showHint && strings.TrimSpace(m.hint) != ""

	// Stylize and center dismissal hint
	var centeredHint string
	contentHeight := len(centeredRows) // notice height
	if showHint {
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "242"}).
			Italic(true)

		hintPadding := (m.width - len([]rune(m.hint))) / 2
		if hintPadding < 0 {
			hintPadding = 0
		}
		centeredHint = strings.Repeat(" ", hintPadding) + hintStyle.Render(m.hint)
		contentHeight += 3 // spacing + hint height
	}

	// Center everything vertically
	topPadding := (m.height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	var sb strings.Builder
	sb.WriteString(strings.Repeat("\n", topPadding))
	sb.WriteString(chunkyBlock)
	if showHint {
		sb.WriteString("\n\n\n") // Gap before hint
		sb.WriteString(centeredHint)
	}

	bottomPadding := m.height - topPadding - contentHeight
	if bottomPadding > 0 {
		sb.WriteString(strings.Repeat("\n", bottomPadding))
	}

	return sb.String()
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
