package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mitchell-wallace/plaqq/internal/config"
	"github.com/mitchell-wallace/plaqq/internal/font"
	"github.com/mitchell-wallace/plaqq/internal/session"
	"github.com/spf13/cobra"
)

// Sentinel Select values for the color field that are not preset names.
const (
	colorDefaultChoice = "default — info"
	colorCustomChoice  = "custom hex / ANSI…"
)

var (
	flagSession     bool
	flagConfigClear bool
	flagConfigColor string
	flagConfigFont  string
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
		if !flagSession {
			if flagConfigClear {
				return errors.New("flag --clear requires --session")
			}
			if cmd.Flags().Changed("color") || cmd.Flags().Changed("font") {
				return errors.New("flags --color and --font require --session")
			}
		}

		if flagSession {
			if flagConfigClear {
				if err := session.Clear(); err != nil {
					return err
				}
				if jsonOutput {
					printJSON(map[string]any{"path": session.Current().Path(), "cleared": true})
					return nil
				}
				fmt.Printf("Cleared session state from %s\n", session.Current().Path())
				return nil
			}

			// If flags are set, save directly without picker
			if cmd.Flags().Changed("color") || cmd.Flags().Changed("font") {
				var sessState session.State
				if cmd.Flags().Changed("color") {
					colorVal := strings.TrimSpace(flagConfigColor)
					if colorVal != "" {
						if _, err := parseColor(colorVal); err != nil {
							return err
						}
					}
					sessState.Color = colorVal
				}
				if cmd.Flags().Changed("font") {
					fontVal := strings.TrimSpace(flagConfigFont)
					if fontVal != "" {
						if !font.Has(fontVal) {
							return fmt.Errorf("unknown font %q: valid fonts are %s", fontVal, strings.Join(font.Names(), ", "))
						}
					}
					sessState.Font = fontVal
				}

				if err := session.Save(sessState); err != nil {
					return err
				}

				if jsonOutput {
					printJSON(map[string]any{
						"path":  session.Current().Path(),
						"color": sessState.Color,
						"font":  sessState.Font,
					})
					return nil
				}

				fmt.Printf("Saved session styling to %s\n\n", session.Current().Path())
				str := func(s string, dflt string) string {
					if s != "" {
						return s
					}
					return dflt + "  (default)"
				}
				var b strings.Builder
				fmt.Fprintf(&b, "  color    %s\n", str(sessState.Color, "info"))
				fmt.Fprintf(&b, "  font     %s\n", str(sessState.Font, font.DefaultName))
				fmt.Print(b.String())
				return nil
			}

			if jsonOutput {
				sessState, _ := session.Load(nil)
				m := map[string]any{"path": session.Current().Path()}
				if sessState.Color != "" {
					m["color"] = sessState.Color
				}
				if sessState.Font != "" {
					m["font"] = sessState.Font
				}
				printJSON(m)
				return nil
			}

			// Otherwise, run the picker seeded with current session state (or fallback config)
			path, err := config.Path()
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			sessState, err := session.Load(styleWarningOutput)
			if err == nil {
				if sessState.Color != "" {
					cfg.Color = &sessState.Color
				}
				if sessState.Font != "" {
					cfg.Font = &sessState.Font
				}
			}
		}

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

type formState struct {
	fontChoice  string
	colorChoice string
	customColor string
	bold        bool
	showHint    bool
	hintText    string
}

// seedFormState determines the initial values for the interactive picker,
// defaulting any invalid or unknown values to the built-in defaults (block / info).
func seedFormState(cfg *config.Config) formState {
	state := formState{
		fontChoice:  font.DefaultName,
		colorChoice: colorDefaultChoice,
		bold:        true,
		showHint:    true,
		hintText:    defaultHint,
	}
	if cfg.Font != nil && *cfg.Font != "" {
		val := strings.TrimSpace(*cfg.Font)
		if font.Has(val) {
			state.fontChoice = val
		}
	}

	if cfg.Color != nil && *cfg.Color != "" {
		val := strings.TrimSpace(*cfg.Color)
		if _, ok := presetColor(val); ok {
			state.colorChoice = strings.ToLower(val)
		} else if _, err := parseColor(val); err == nil {
			state.colorChoice = colorCustomChoice
			state.customColor = val
		}
	}

	if cfg.Bold != nil {
		state.bold = *cfg.Bold
	}
	if cfg.NoHint != nil {
		state.showHint = !*cfg.NoHint
	}
	if cfg.Hint != nil {
		state.hintText = *cfg.Hint
	}
	return state
}

// buildSavedConfig creates a Config object from the picker choices.
// Fields left at their built-in default are omitted (left nil) so the saved file remains minimal.
func buildSavedConfig(fontChoice, colorChoice, customColor string, bold, showHint bool, hintText string) *config.Config {
	out := &config.Config{}
	if fontChoice != "" && fontChoice != font.DefaultName {
		f := fontChoice
		out.Font = &f
	}
	switch colorChoice {
	case colorDefaultChoice:
		// leave nil -> built-in info
	case colorCustomChoice:
		if c := strings.TrimSpace(customColor); c != "" {
			if _, err := parseColor(c); err == nil {
				out.Color = &c
			}
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
	return out
}

// runConfigPicker presents an interactive form seeded with the current config,
// then writes the chosen styling back to path.
func runConfigPicker(path string, cfg *config.Config) error {
	state := seedFormState(cfg)
	fontChoice := state.fontChoice
	colorChoice := state.colorChoice
	customColor := state.customColor
	bold := state.bold
	showHint := state.showHint
	hintText := state.hintText

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Font").
				Description("block fonts render in Unicode blocks (block, heavy, compact, wide)").
				Options(fontOptions()...).
				Value(&fontChoice),
			huh.NewSelect[string]().
				Title("Color").
				Options(colorOptions()...).
				Value(&colorChoice),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Bold text?").
				Value(&bold),
			huh.NewConfirm().
				Title("Show the dismiss hint?").
				Value(&showHint),
		).WithHideFunc(func() bool { return flagSession }),
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
		).WithHideFunc(func() bool { return !showHint || flagSession }),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprintln(os.Stderr, "config unchanged")
			return nil
		}
		return err
	}

	out := buildSavedConfig(fontChoice, colorChoice, customColor, bold, showHint, hintText)

	if flagSession {
		state := session.State{
			Color: "",
			Font:  "",
		}
		if out.Color != nil {
			state.Color = *out.Color
		}
		if out.Font != nil {
			state.Font = *out.Font
		}
		if err := session.Save(state); err != nil {
			return err
		}
		fmt.Printf("Saved session styling to %s\n\n", session.Current().Path())
		str := func(s string, dflt string) string {
			if s != "" {
				return s
			}
			return dflt + "  (default)"
		}
		var b strings.Builder
		fmt.Fprintf(&b, "  color    %s\n", str(state.Color, "info"))
		fmt.Fprintf(&b, "  font     %s\n", str(state.Font, font.DefaultName))
		fmt.Print(b.String())
		return nil
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
	fmt.Fprintf(&b, "  color    %s\n", str(cfg.Color, "info"))
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
	configCmd.Flags().BoolVar(&flagSession, "session", false, "edit or clear session-state instead of persistent config")
	configCmd.Flags().BoolVar(&flagConfigClear, "clear", false, "remove the session-state record (requires --session)")
	configCmd.Flags().StringVar(&flagConfigColor, "color", "", "session color: a preset name, a hex code, or an ANSI index")
	configCmd.Flags().StringVar(&flagConfigFont, "font", "", "session font: one of the registered fonts")

	configCmd.AddCommand(configPathCmd, configInitCmd)
	rootCmd.AddCommand(configCmd)
}
