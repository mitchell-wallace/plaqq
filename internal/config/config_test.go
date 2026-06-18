package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileIsEmpty(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err != nil {
		t.Fatalf("Load missing file: %v", err)
	}
	if cfg.Color != nil || cfg.Bold != nil || cfg.Hint != nil || cfg.NoHint != nil {
		t.Errorf("expected all-nil config, got %+v", cfg)
	}
}

func TestLoadParsesValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("color = \"#ff5f87\"\nbold = false\nno_hint = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Color == nil || *cfg.Color != "#ff5f87" {
		t.Errorf("color = %v; want #ff5f87", cfg.Color)
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
