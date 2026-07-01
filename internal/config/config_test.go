package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileIsEmpty(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err != nil {
		t.Fatalf("Load missing file: %v", err)
	}
	if cfg.Color != nil || cfg.Font != nil || cfg.Frame != nil || cfg.Bold != nil || cfg.Hint != nil || cfg.NoHint != nil {
		t.Errorf("expected all-nil config, got %+v", cfg)
	}
}

func TestLoadParsesValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"#ff5f87\"\nframe = \"single\"\nbold = false\nno_hint = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Color == nil || *cfg.Color != "#ff5f87" {
		t.Errorf("color = %v; want #ff5f87", cfg.Color)
	}
	if cfg.Frame == nil || *cfg.Frame != "single" {
		t.Errorf("frame = %v; want single", cfg.Frame)
	}
	if cfg.Bold == nil || *cfg.Bold != false {
		t.Errorf("bold = %v; want false", cfg.Bold)
	}
	if cfg.NoHint == nil || *cfg.NoHint != true {
		t.Errorf("no_hint = %v; want true", cfg.NoHint)
	}
	if cfg.Hint != nil {
		t.Errorf("hint = %v; want nil (unset)", cfg.Hint)
	}
}

func TestSaveRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.toml")
	color, font, frame, hint := "coral", "heavy", "double", "press space"
	bold, noHint := false, true
	in := &Config{Color: &color, Font: &font, Frame: &frame, Bold: &bold, Hint: &hint, NoHint: &noHint}

	if err := Save(path, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load saved config: %v", err)
	}
	if got.Color == nil || *got.Color != color || got.Font == nil || *got.Font != font {
		t.Errorf("color/font round-trip = %v/%v; want %q/%q", got.Color, got.Font, color, font)
	}
	if got.Frame == nil || *got.Frame != frame {
		t.Errorf("frame round-trip = %v; want %q", got.Frame, frame)
	}
	if got.Bold == nil || *got.Bold != bold || got.NoHint == nil || *got.NoHint != noHint {
		t.Errorf("bold/no_hint round-trip = %v/%v; want %v/%v", got.Bold, got.NoHint, bold, noHint)
	}
	if got.Hint == nil || *got.Hint != hint {
		t.Errorf("hint round-trip = %v; want %q", got.Hint, hint)
	}
}

func TestSaveOmitsUnsetFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	font := "compact"
	if err := Save(path, &Config{Font: &font}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Font == nil || *got.Font != font {
		t.Errorf("font = %v; want %q", got.Font, font)
	}
	if got.Color != nil || got.Frame != nil || got.Bold != nil || got.Hint != nil || got.NoHint != nil {
		t.Errorf("unset fields should stay nil, got %+v", got)
	}
}

func TestSaveOmitsFrameWhenNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	frame := "block"
	if err := Save(path, &Config{Frame: &frame}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Frame == nil || *got.Frame != frame {
		t.Errorf("frame = %v; want %q", got.Frame, frame)
	}
	if got.Color != nil || got.Font != nil || got.Bold != nil || got.Hint != nil || got.NoHint != nil {
		t.Errorf("only frame should be set, got %+v", got)
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = "), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error for invalid TOML, got nil")
	}
}

func TestPathHonorsEnv(t *testing.T) {
	want := "/custom/plaqq.toml"
	t.Setenv("PLAQQ_CONFIG", want)
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Path() = %q; want %q", got, want)
	}
}

func TestTemplateContainsFrameComment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := WriteTemplate(path); err != nil {
		t.Fatalf("WriteTemplate: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Contains(data, []byte("# frame = \"single\"")) {
		t.Errorf("template missing frame comment block; got:\n%s", data)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("template is not valid TOML: %v", err)
	}
	if cfg.Color != nil || cfg.Font != nil || cfg.Frame != nil || cfg.Bold != nil || cfg.Hint != nil || cfg.NoHint != nil {
		t.Errorf("template should parse to all-nil config, got %+v", cfg)
	}
}

func TestWriteTemplateRefusesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.toml")

	if err := WriteTemplate(path); err != nil {
		t.Fatalf("WriteTemplate: %v", err)
	}
	// The template must round-trip through the parser.
	if _, err := Load(path); err != nil {
		t.Fatalf("template is not valid TOML: %v", err)
	}
	// A second write must refuse rather than clobber.
	if err := WriteTemplate(path); err == nil {
		t.Error("expected WriteTemplate to refuse overwriting, got nil")
	}
}
