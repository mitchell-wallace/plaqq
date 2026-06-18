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
	flagBold   bool
	flagHint   string
	flagNoHint bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json-output", false, "emit structured JSON output")
	rootCmd.Flags().StringVar(&flagColor, "color", "", "notice text color as a hex code (e.g. #00f5d4) or ANSI index (0-255); defaults to adaptive teal")
	rootCmd.Flags().BoolVar(&flagBold, "bold", true, "render the notice text in bold")
	rootCmd.Flags().StringVar(&flagHint, "hint", defaultHint, "dismiss-hint text shown beneath the notice")
	rootCmd.Flags().BoolVar(&flagNoHint, "no-hint", false, "hide the dismiss hint")
}

// parseColor validates a user-supplied color string and returns the matching
// lipgloss color. It accepts a hex code (#rgb or #rrggbb) or an ANSI index (0-255).
func parseColor(s string) (lipgloss.TerminalColor, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("color is empty")
	}

	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		if len(hex) != 3 && len(hex) != 6 {
			return nil, fmt.Errorf("invalid hex color %q: expected #rgb or #rrggbb", s)
		}
		for _, r := range hex {
			if !isHexDigit(r) {
				return nil, fmt.Errorf("invalid hex color %q", s)
			}
		}
		return lipgloss.Color(s), nil
	}

	if n, err := strconv.Atoi(s); err == nil {
		if n < 0 || n > 255 {
			return nil, fmt.Errorf("invalid ANSI color %d: expected an index 0-255", n)
		}
		return lipgloss.Color(s), nil
	}

	return nil, fmt.Errorf("invalid color %q: use a hex code like #00f5d4 or an ANSI index 0-255", s)
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// resolveStyle determines the effective notice styling by layering, in order:
// built-in defaults, values from the config file, then any CLI flags the user
// explicitly set (so a flag always wins over the config file).
func resolveStyle(cmd *cobra.Command) (color string, bold bool, hint string, noHint bool, err error) {
	color, bold, hint, noHint = "", true, defaultHint, false

	path, err := config.Path()
	if err != nil {
		return "", false, "", false, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return "", false, "", false, err
	}
	if cfg.Color != nil {
		color = *cfg.Color
	}
	if cfg.Bold != nil {
		bold = *cfg.Bold
	}
	if cfg.Hint != nil {
		hint = *cfg.Hint
	}
	if cfg.NoHint != nil {
		noHint = *cfg.NoHint
	}

	if cmd.Flags().Changed("color") {
		color = flagColor
	}
	if cmd.Flags().Changed("bold") {
		bold = flagBold
	}
	if cmd.Flags().Changed("hint") {
		hint = flagHint
	}
	if cmd.Flags().Changed("no-hint") {
		noHint = flagNoHint
	}

	return color, bold, hint, noHint, nil
}

var rootCmd = &cobra.Command{
	Use:   "plaqq [message]",
	Short: "plaqq displays a notice across the terminal pane in chunky letters",
	Long: `plaqq takes a message and displays it in a large chunky ASCII font centered
on the terminal. It waits for the user to press the spacebar to dismiss the notice.

The notice color, bold weight, and dismiss hint can be customized with the
--color, --bold, --hint, and --no-hint flags, or set as persistent defaults
in the config file (see 'plaqq config'). Flags override the config file.`,
	Example: `  plaqq "deploy starting"
  plaqq --color "#ff5f87" "build failed"
  plaqq --color 213 --bold=false "heads up"
  plaqq --hint "press space to continue" "meeting in 5"
  plaqq --no-hint "stand clear"`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve styling with precedence: built-in defaults < config file < CLI flags.
		colorStr, bold, hint, noHint, err := resolveStyle(cmd)
		if err != nil {
			return err
		}

		var noticeColor lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#00d7af", Dark: "#00f5d4"}
		if colorStr != "" {
			c, err := parseColor(colorStr)
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
		m := initialModel(noticeMsg, noticeColor, bold, hint, !noHint)
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
	color    lipgloss.TerminalColor
	width    int
	height   int
	bold     bool
	showHint bool
}

func initialModel(text string, color lipgloss.TerminalColor, bold bool, hint string, showHint bool) model {
	return model{
		text:     text,
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

	lines := font.WrapText(m.text, maxWidth)

	// Build the chunky ASCII text rows
	var chunkyRows []string
	for idx, line := range lines {
		if idx > 0 {
			// Blank spacing row between chunky wrap lines
			chunkyRows = append(chunkyRows, "", "")
		}

		renderedLine := font.RenderString(line)
		for r := 0; r < font.Height; r++ {
			chunkyRows = append(chunkyRows, renderedLine[r])
		}
	}

	// Stylize and center each chunky row horizontally
	noticeStyle := lipgloss.NewStyle().Foreground(m.color).Bold(m.bold)

	var centeredRows []string
	for _, row := range chunkyRows {
		if row == "" {
			centeredRows = append(centeredRows, "")
			continue
		}

		rowRunesCount := len([]rune(row))
		padding := (m.width - rowRunesCount) / 2
		if padding < 0 {
			padding = 0
		}

		centeredRow := strings.Repeat(" ", padding) + noticeStyle.Render(row)
		centeredRows = append(centeredRows, centeredRow)
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
