package cmd

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// colorPresets maps friendly names to terminal colors. "teal" is the adaptive
// default (it reads well on both light and dark backgrounds); the rest are
// fixed values picked to stay legible across terminal themes.
var colorPresets = map[string]lipgloss.TerminalColor{
	"teal":    lipgloss.AdaptiveColor{Light: "#00d7af", Dark: "#00f5d4"},
	"coral":   lipgloss.Color("#ff5f87"),
	"amber":   lipgloss.Color("#ffaf00"),
	"lime":    lipgloss.Color("#5fd75f"),
	"azure":   lipgloss.Color("#5fafff"),
	"violet":  lipgloss.Color("#af87ff"),
	"magenta": lipgloss.Color("#ff5fff"),
	"rose":    lipgloss.Color("#ff87af"),
	"crimson": lipgloss.Color("#d70000"),
	"slate":   lipgloss.Color("#9e9e9e"),
}

// presetOrder is the stable display order for presets in help text and the
// interactive picker.
var presetOrder = []string{
	"teal", "coral", "amber", "lime", "azure",
	"violet", "magenta", "rose", "crimson", "slate",
}

// presetColor looks up a color preset by name, case-insensitively.
func presetColor(name string) (lipgloss.TerminalColor, bool) {
	c, ok := colorPresets[strings.ToLower(strings.TrimSpace(name))]
	return c, ok
}
