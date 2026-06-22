package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mitchell-wallace/plaqq/internal/config"
	"github.com/mitchell-wallace/plaqq/internal/font"
	"github.com/spf13/cobra"
)

// Sentinel Select values for the color field that are not preset names.
const (
	colorDefaultChoice = "default — adaptive teal"
	colorCustomChoice  = "custom hex / ANSI…"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit plaqq styling defaults interactively",
	Long: `plaqq reads styling defaults (color, font, bold, hint, no-hint) from a TOML
config file. CLI flags override anything set in the file.

Run 'plaqq config' with no subcommand to edit those defaults in an interactive
picker. Use the subcommands to inspect or scaffold the file directly.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		if jsonOutput {
			printJSON(configJSON(path, cfg))
			return nil
		}
		return runConfigPicker(path, cfg)
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the path to the plaqq config file",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}
		_, statErr := os.Stat(path)
		exists := statErr == nil

		if jsonOutput {
			printJSON(map[string]any{"path": path, "exists": exists})
			return nil
		}
		fmt.Println(path)
		if !exists {
			fmt.Fprintln(os.Stderr, "(does not exist yet; run 'plaqq config' to create it)")
		}
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a commented config template to the config path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}
		if err := config.WriteTemplate(path); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(map[string]any{"path": path, "created": true})
			return nil
		}
		fmt.Printf("Wrote config template to %s\n", path)
		return nil
	},
}

// runConfigPicker presents an interactive form seeded with the current config,
// then writes the chosen styling back to path.
func runConfigPicker(path string, cfg *config.Config) error {
	// Seed form state from the current config, falling back to built-in defaults.
	fontChoice := font.DefaultName
	if cfg.Font != nil && *cfg.Font != "" {
		fontChoice = *cfg.Font
	}

	colorChoice := colorDefaultChoice
	customColor := ""
	if cfg.Color != nil {
		if _, ok := presetColor(*cfg.Color); ok {
			colorChoice = strings.ToLower(strings.TrimSpace(*cfg.Color))
		} else {
			colorChoice = colorCustomChoice
			customColor = *cfg.Color
		}
	}

	bold := true
	if cfg.Bold != nil {
		bold = *cfg.Bold
	}
	showHint := true
	if cfg.NoHint != nil {
		showHint = !*cfg.NoHint
	}
	hintText := defaultHint
	if cfg.Hint != nil {
		hintText = *cfg.Hint
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Font").
				Description("block fonts render in Unicode blocks (block, heavy, compact)").
				Options(fontOptions()...).
				Value(&fontChoice),
			huh.NewSelect[string]().
				Title("Color").
				Options(colorOptions()...).
				Value(&colorChoice),
			huh.NewConfirm().
				Title("Bold text?").
				Value(&bold),
			huh.NewConfirm().
				Title("Show the dismiss hint?").
				Value(&showHint),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom color").
				Description("hex like #ff5f87, or an ANSI index 0-255").
				Value(&customColor).
				Validate(validateOptionalColor),
		).WithHideFunc(func() bool { return colorChoice != colorCustomChoice }),
		huh.NewGroup(
			huh.NewInput().
				Title("Hint text").
				Value(&hintText),
		).WithHideFunc(func() bool { return !showHint }),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprintln(os.Stderr, "config unchanged")
			return nil
		}
		return err
	}

	// Build a config from the choices. Fields left at their built-in default are
	// omitted so the saved file stays minimal.
	out := &config.Config{}
	if fontChoice != "" && fontChoice != font.DefaultName {
		f := fontChoice
		out.Font = &f
	}
	switch colorChoice {
	case colorDefaultChoice:
		// leave nil -> built-in adaptive teal
	case colorCustomChoice:
		if c := strings.TrimSpace(customColor); c != "" {
			out.Color = &c
		}
	default:
		c := colorChoice
		out.Color = &c
	}
	b := bold
	out.Bold = &b
	nh := !showHint
	out.NoHint = &nh
	if showHint {
		if h := strings.TrimSpace(hintText); h != "" && h != defaultHint {
			out.Hint = &h
		}
	}

	if err := config.Save(path, out); err != nil {
		return err
	}

	fmt.Printf("Saved styling defaults to %s\n\n", path)
	fmt.Print(configSummary(out))
	return nil
}

func fontOptions() []huh.Option[string] {
	names := font.Names()
	opts := make([]huh.Option[string], 0, len(names))
	for _, name := range names {
		label := name
		if name == font.DefaultName {
			label = name + " (default)"
		}
		opts = append(opts, huh.NewOption(label, name))
	}
	return opts
}

func colorOptions() []huh.Option[string] {
	opts := []huh.Option[string]{huh.NewOption(colorDefaultChoice, colorDefaultChoice)}
	for _, name := range presetOrder {
		opts = append(opts, huh.NewOption(name, name))
	}
	return append(opts, huh.NewOption(colorCustomChoice, colorCustomChoice))
}

func validateOptionalColor(s string) error {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	_, err := parseColor(s)
	return err
}

// configSummary renders the effective settings for the saved config, showing
// "(default)" for any option that falls through to a built-in default.
func configSummary(cfg *config.Config) string {
	str := func(p *string, dflt string) string {
		if p != nil {
			return *p
		}
		return dflt + "  (default)"
	}
	boolStr := func(p *bool, dflt bool) string {
		if p != nil {
			return fmt.Sprintf("%t", *p)
		}
		return fmt.Sprintf("%t  (default)", dflt)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "  color    %s\n", str(cfg.Color, "teal"))
	fmt.Fprintf(&b, "  font     %s\n", str(cfg.Font, font.DefaultName))
	fmt.Fprintf(&b, "  bold     %s\n", boolStr(cfg.Bold, true))
	fmt.Fprintf(&b, "  hint     %s\n", str(cfg.Hint, defaultHint))
	fmt.Fprintf(&b, "  no_hint  %s\n", boolStr(cfg.NoHint, false))
	return b.String()
}

// configJSON is the structured view of the config for --json-output.
func configJSON(path string, cfg *config.Config) map[string]any {
	m := map[string]any{"path": path}
	if cfg.Color != nil {
		m["color"] = *cfg.Color
	}
	if cfg.Font != nil {
		m["font"] = *cfg.Font
	}
	if cfg.Bold != nil {
		m["bold"] = *cfg.Bold
	}
	if cfg.Hint != nil {
		m["hint"] = *cfg.Hint
	}
	if cfg.NoHint != nil {
		m["no_hint"] = *cfg.NoHint
	}
	return m
}

func init() {
	configCmd.AddCommand(configPathCmd, configInitCmd)
	rootCmd.AddCommand(configCmd)
}
