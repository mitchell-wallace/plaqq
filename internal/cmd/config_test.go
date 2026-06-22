package cmd

import (
	"testing"

	"github.com/mitchell-wallace/plaqq/internal/config"
	"github.com/mitchell-wallace/plaqq/internal/font"
)

func TestSeedFormStateTolerance(t *testing.T) {
	// 1. Config with unknown/removed font and color should fall back to defaults
	badFont := "slant"
	badColor := "coral"
	cfg1 := &config.Config{
		Font:  &badFont,
		Color: &badColor,
	}

	fontChoice, colorChoice, customColor, bold, showHint, hintText := seedFormState(cfg1)

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

	// 2. Config with valid font and color preset should be seeded correctly
	goodFont := "heavy"
	goodColorPreset := "focus"
	cfg2 := &config.Config{
		Font:  &goodFont,
		Color: &goodColorPreset,
	}

	fontChoice2, colorChoice2, customColor2, _, _, _ := seedFormState(cfg2)
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

	_, colorChoice3, customColor3, _, _, _ := seedFormState(cfg3)
	if colorChoice3 != colorCustomChoice {
		t.Errorf("expected colorChoice to be %q, got %q", colorCustomChoice, colorChoice3)
	}
	if customColor3 != goodCustomColor {
		t.Errorf("expected customColor to be %q, got %q", goodCustomColor, customColor3)
	}
}

func TestBuildSavedConfig(t *testing.T) {
	// 1. If choices are defaults, the saved Config should have nil fields to remain minimal
	cfg := buildSavedConfig(font.DefaultName, colorDefaultChoice, "", true, true, defaultHint)

	if cfg.Font != nil {
		t.Errorf("expected default font to be omitted (nil), got %q", *cfg.Font)
	}
	if cfg.Color != nil {
		t.Errorf("expected default color to be omitted (nil), got %q", *cfg.Color)
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
	cfg2 := buildSavedConfig("heavy", "warn", "", false, false, "custom hint")

	if cfg2.Font == nil || *cfg2.Font != "heavy" {
		t.Errorf("expected font to be saved as 'heavy', got %v", cfg2.Font)
	}
	if cfg2.Color == nil || *cfg2.Color != "warn" {
		t.Errorf("expected color to be saved as 'warn', got %v", cfg2.Color)
	}
	if cfg2.Bold == nil || *cfg2.Bold != false {
		t.Errorf("expected bold to be saved as false, got %v", cfg2.Bold)
	}
	if cfg2.NoHint == nil || *cfg2.NoHint != true {
		t.Errorf("expected no_hint to be saved as true, got %v", cfg2.NoHint)
	}
	// Note: hint text is not saved if showHint is false, because it's hidden. Let's test showHint = true with custom hint.
	cfg3 := buildSavedConfig("compact", colorCustomChoice, "#aabbcc", true, true, "custom hint")
	if cfg3.Color == nil || *cfg3.Color != "#aabbcc" {
		t.Errorf("expected custom color to be saved as '#aabbcc', got %v", cfg3.Color)
	}
	if cfg3.Hint == nil || *cfg3.Hint != "custom hint" {
		t.Errorf("expected custom hint to be saved, got %v", cfg3.Hint)
	}
}
