package session

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSaveThenLoadRoundTrips(t *testing.T) {
	store := NewStore(t.TempDir(), 1001)
	in := State{Color: "alert", Font: "heavy"}

	if err := store.Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var warnings bytes.Buffer
	got, err := store.Load(&warnings)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != in {
		t.Fatalf("Load = %+v; want %+v", got, in)
	}
	if warnings.Len() != 0 {
		t.Fatalf("unexpected warnings: %s", warnings.String())
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(store.Path())
		if err != nil {
			t.Fatalf("stat state file: %v", err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Fatalf("state file mode = %o; want 0600", mode)
		}
	}
}

func TestSaveReplacesExistingFileAtomicallyWithPrivateMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file mode expectations do not apply on Windows")
	}

	store := NewStore(t.TempDir(), 1002)
	if err := os.MkdirAll(filepath.Dir(store.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Path(), []byte("font = \"block\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(State{Color: "ok", Font: "compact"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := string(got)
	if strings.Contains(text, "block") || !strings.Contains(text, `font = "compact"`) {
		t.Fatalf("state file content = %q; want replacement with compact only", text)
	}

	info, err := os.Stat(store.Path())
	if err != nil {
		t.Fatalf("stat state file: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("state file mode = %o; want 0600", mode)
	}
}

func TestClearRemovesState(t *testing.T) {
	store := NewStore(t.TempDir(), 1003)
	if err := store.Save(State{Color: "warn", Font: "compact"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if _, err := os.Stat(store.Path()); !os.IsNotExist(err) {
		t.Fatalf("state file exists after Clear: %v", err)
	}

	got, err := store.Load(nil)
	if err != nil {
		t.Fatalf("Load after Clear: %v", err)
	}
	if !got.Empty() {
		t.Fatalf("Load after Clear = %+v; want empty", got)
	}
}

func TestDistinctKeysIsolateState(t *testing.T) {
	dir := t.TempDir()
	first := NewStore(dir, 2001)
	second := NewStore(dir, 2002)

	if err := first.Save(State{Color: "alert", Font: "heavy"}); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	gotSecond, err := second.Load(nil)
	if err != nil {
		t.Fatalf("second Load: %v", err)
	}
	if !gotSecond.Empty() {
		t.Fatalf("fresh second key = %+v; want empty", gotSecond)
	}

	if err := second.Save(State{Color: "ok", Font: "compact"}); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	gotFirst, err := first.Load(nil)
	if err != nil {
		t.Fatalf("first Load: %v", err)
	}
	if gotFirst != (State{Color: "alert", Font: "heavy"}) {
		t.Fatalf("first key changed to %+v; want original", gotFirst)
	}
}

func TestMalformedFileWarnsAndReturnsEmpty(t *testing.T) {
	store := NewStore(t.TempDir(), 3001)
	if err := os.MkdirAll(filepath.Dir(store.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Path(), []byte("font = "), 0o600); err != nil {
		t.Fatal(err)
	}

	var warnings bytes.Buffer
	got, err := store.Load(&warnings)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !got.Empty() {
		t.Fatalf("Load malformed = %+v; want empty", got)
	}
	warn := warnings.String()
	if !strings.Contains(warn, "plaqq: warning: ignoring session state") || !strings.Contains(warn, store.Path()) {
		t.Fatalf("warning = %q; want warning naming state file", warn)
	}
}

func TestInvalidRecordWarnsAndReturnsEmpty(t *testing.T) {
	store := NewStore(t.TempDir(), 3002)
	if err := os.MkdirAll(filepath.Dir(store.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Path(), []byte("font = 42\ncolor = \"alert\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var warnings bytes.Buffer
	got, err := store.Load(&warnings)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !got.Empty() {
		t.Fatalf("Load invalid = %+v; want empty", got)
	}
	if warn := warnings.String(); !strings.Contains(warn, "plaqq: warning: ignoring session state") {
		t.Fatalf("warning = %q; want warning", warn)
	}
}

func TestDefaultDirUsesXDGRuntimePlaqqDirThenTempFallback(t *testing.T) {
	runtimeDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	if got, want := defaultDir(), filepath.Join(runtimeDir, "plaqq"); got != want {
		t.Fatalf("defaultDir with XDG_RUNTIME_DIR = %q; want %q", got, want)
	}

	t.Setenv("XDG_RUNTIME_DIR", "")
	if got, want := defaultDir(), os.TempDir(); got != want {
		t.Fatalf("defaultDir fallback = %q; want %q", got, want)
	}
}
