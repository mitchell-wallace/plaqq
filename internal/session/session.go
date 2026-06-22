// Package session persists per-terminal style choices for the current shell.
package session

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// State is the small per-terminal styling record. Empty fields are ignored by
// callers so unset values fall through to lower-precedence style layers.
type State struct {
	Color string `toml:"color"`
	Font  string `toml:"font"`
}

// Empty reports whether the record has no styling values set.
func (s State) Empty() bool {
	return strings.TrimSpace(s.Color) == "" && strings.TrimSpace(s.Font) == ""
}

// Store reads and writes the session-state record for one terminal key.
type Store struct {
	dir string
	key int
}

// Current returns the store for the shell that launched this plaqq process.
func Current() Store {
	// PPID is a zero-dependency approximation of "this terminal session": every
	// plaqq run from one shell prompt shares the parent shell PID. PID reuse after
	// that shell exits can make a stale file appear current; users can clear it,
	// and a future TTY/start-time key could refine this without changing callers.
	return NewStore(defaultDir(), os.Getppid())
}

// NewStore returns a store rooted at dir and keyed by key. It is exported so
// tests and future command plumbing can isolate session records deliberately.
func NewStore(dir string, key int) Store {
	if dir == "" {
		dir = defaultDir()
	}
	return Store{dir: dir, key: key}
}

// Load reads the current session-state record. A missing, malformed, or
// unreadable record returns an empty State with no error; when warn is non-nil,
// Load writes a warning explaining why the record was ignored.
func Load(warn io.Writer) (State, error) {
	return Current().Load(warn)
}

// Save atomically writes the current session-state record.
func Save(state State) error {
	return Current().Save(state)
}

// Clear removes the current session-state record.
func Clear() error {
	return Current().Clear()
}

// Path returns the session-state file path for this store.
func (s Store) Path() string {
	return filepath.Join(s.dir, fmt.Sprintf("session-%d.toml", s.key))
}

// Load reads this store's session-state record.
func (s Store) Load(warn io.Writer) (State, error) {
	path := s.Path()
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			warnf(warn, "ignoring session state %s: %v", path, err)
		}
		return State{}, nil
	}

	var state State
	if err := toml.Unmarshal(data, &state); err != nil {
		warnf(warn, "ignoring session state %s: %v", path, err)
		return State{}, nil
	}

	state.Color = strings.TrimSpace(state.Color)
	state.Font = strings.TrimSpace(state.Font)
	return state, nil
}

// Save writes this store's session-state record using a temp file in the target
// directory followed by rename. The temp file is mode 0600 so replacement also
// repairs any overly broad permissions on an existing state file.
func (s Store) Save(state State) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(s.dir, ".session-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.WriteString(formatState(state)); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, s.Path()); err != nil {
		return err
	}
	removeTmp = false
	return nil
}

// Clear removes this store's session-state record. Removing an already-missing
// record is successful so callers can use it as an idempotent reset.
func (s Store) Clear() error {
	if err := os.Remove(s.Path()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func defaultDir() string {
	if runtimeDir := strings.TrimSpace(os.Getenv("XDG_RUNTIME_DIR")); runtimeDir != "" {
		return filepath.Join(runtimeDir, "plaqq")
	}
	return os.TempDir()
}

func formatState(state State) string {
	var b strings.Builder
	b.WriteString("# plaqq session state -- managed by plaqq.\n")
	if color := strings.TrimSpace(state.Color); color != "" {
		fmt.Fprintf(&b, "color = %s\n", strconv.Quote(color))
	}
	if font := strings.TrimSpace(state.Font); font != "" {
		fmt.Fprintf(&b, "font = %s\n", strconv.Quote(font))
	}
	return b.String()
}

func warnf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, "plaqq: warning: "+format+"\n", args...)
}
