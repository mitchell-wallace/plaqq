package cmd

import (
	"fmt"
	"os"

	"github.com/mitchell-wallace/plaqq/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the plaqq config file",
	Long: `plaqq reads styling defaults (color, bold, hint, no-hint) from a TOML
config file. CLI flags override anything set in the file.`,
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
			fmt.Fprintln(os.Stderr, "(does not exist yet; run 'plaqq config init' to create it)")
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

func init() {
	configCmd.AddCommand(configPathCmd, configInitCmd)
	rootCmd.AddCommand(configCmd)
}
