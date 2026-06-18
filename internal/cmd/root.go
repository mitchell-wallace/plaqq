package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchell-wallace/plaqq/internal/font"
	"github.com/spf13/cobra"
)

var (
	version    string
	jsonOutput bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json-output", false, "emit structured JSON output")
}

var rootCmd = &cobra.Command{
	Use:   "plaqq [message]",
	Short: "plaqq displays a notice across the terminal pane in chunky letters",
	Long: `plaqq takes a message and displays it in a large chunky ASCII font centered 
on the terminal. It waits for the user to press the spacebar to dismiss the notice.`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default notice message if none is provided
		noticeMsg := "REMEMBER TO RUN E2E TESTS BEFORE PUSHING"
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			noticeMsg = args[0]
		}

		// Start update check in background
		updateNoticeChan := make(chan string, 1)
		if !jsonOutput && version != "" && version != "dev" {
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
		m := initialModel(noticeMsg)
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithoutCatchPanics())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("run program: %w", err)
		}

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
	text   string
	width  int
	height int
}

func initialModel(text string) model {
	return model{
		text: text,
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
	noticeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#00d7af", Dark: "#00f5d4"}).
		Bold(true)

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

	// Stylize and center dismissal hint
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "242"}).
		Italic(true)

	hint := "[ Press Space to dismiss ]"
	hintPadding := (m.width - len(hint)) / 2
	if hintPadding < 0 {
		hintPadding = 0
	}
	centeredHint := strings.Repeat(" ", hintPadding) + hintStyle.Render(hint)

	// Center everything vertically
	contentHeight := len(centeredRows) + 3 // notice height + spacing + hint height
	topPadding := (m.height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	var sb strings.Builder
	sb.WriteString(strings.Repeat("\n", topPadding))
	sb.WriteString(chunkyBlock)
	sb.WriteString("\n\n\n") // Gap before hint
	sb.WriteString(centeredHint)

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
