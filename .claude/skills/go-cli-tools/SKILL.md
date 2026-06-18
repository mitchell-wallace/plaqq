---
name: go-cli-tools
description: >-
  Build ergonomic, self-distributing Go CLI tools. Covers command/flag design
  with cobra, layered config (defaults < file < flags), interactive prompts
  with huh, structured --json-output, full-screen TUI notices with
  bubbletea/lipgloss, testing patterns, and — in depth — the distribution and
  self-update lifecycle: version embedding, VERSION-file auto-tag + goreleaser
  releases, install.sh/install.ps1, atomically replacing a running binary
  (the "text file busy" trap), and a self-update command that degrades
  gracefully on dev builds. Use when scaffolding a new Go CLI, adding
  flags/config/JSON output, wiring up releases or install scripts, or
  writing/fixing a self-update ("upgrade") command.
---

# Ergonomic Go CLI tools

Patterns for small, single-binary Go CLIs that feel good to use and can ship and
update themselves. The stack is opinionated: **cobra** (commands/flags),
**huh** (interactive prompts), **lipgloss/bubbletea** (styling and TUI),
**goreleaser** + **GitHub Releases** (distribution), and POSIX/PowerShell
install scripts. `plaqq` (the repo this skill lives in) is the reference
implementation — read its `internal/cmd/`, `justfile`, `install.sh`,
`install.ps1`, and `.github/workflows/` alongside this guide.

The deep distribution + self-update material lives in
[references/self-update.md](references/self-update.md); read it before touching
install scripts or an update command.

## Project layout

```
cmd/<tool>/main.go        # thin entrypoint: var version="dev"; call internal/cmd.Execute(version)
internal/cmd/             # cobra commands (root.go, version.go, update.go, ...)
internal/config/          # config file load/save + path resolution
internal/<feature>/       # the actual work, kept out of the cmd layer
VERSION                   # single source of truth for the release version
justfile                  # build/test/lint/install recipes
.goreleaser.yaml          # release build matrix
install.sh / install.ps1  # one-line installers
```

Keep `main` tiny so the version string is injected in exactly one place and the
command layer is testable without a real `main`:

```go
// cmd/<tool>/main.go
package main

import (
	"fmt"
	"os"

	"github.com/you/mytool/internal/cmd"
)

var version = "dev" // overridden at build time via -ldflags -X

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

## 1. Commands & flags (cobra)

- Give the root command `SilenceUsage: true` and `SilenceErrors: true` so a
  runtime error doesn't dump the full usage text; you control error printing.
- Validate positional args with `cobra.MaximumNArgs(1)` / `ExactArgs` rather
  than hand-rolled length checks.
- Put flags that every subcommand needs (e.g. `--json-output`) on
  `rootCmd.PersistentFlags()`; put display flags on the command that uses them.
- Prefer `RunE` (return an error) over `Run` for anything that can fail — it
  composes with `SilenceErrors` and your central error handler.

```go
rootCmd := &cobra.Command{
	Use:           "mytool [message]",
	Short:         "one-line description",
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error { /* ... */ },
}
rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json-output", false, "emit structured JSON")
```

## 2. Layered config (defaults < file < flags)

The ergonomic rule users expect: **a flag always wins over the config file,
which wins over built-in defaults.** Detect *explicitly set* flags with
`cmd.Flags().Changed(name)` — never compare against the zero value, or you can't
tell `--bold=false` from "unset".

```go
func resolve(cmd *cobra.Command) (Settings, error) {
	s := defaults()                       // 1. built-in defaults

	cfg, err := config.Load(path)         // 2. config file overrides defaults
	if err != nil { return s, err }
	cfg.ApplyTo(&s)                       // only fields present in the file

	if cmd.Flags().Changed("color") {     // 3. explicit flags win
		s.Color = flagColor
	}
	return s, nil
}
```

Resolve the config path with an env override so tests can point at a temp file
and the XDG location is honored in production:

```go
func Path() (string, error) {
	if p := os.Getenv("MYTOOL_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir() // ~/.config on Linux, App Support on macOS
	if err != nil { return "", err }
	return filepath.Join(dir, "mytool", "config.toml"), nil
}
```

Use pointer fields (`*string`, `*bool`) in the config struct so "absent" is
distinguishable from "set to the zero value" when layering. Ship a
`config init` subcommand that writes a commented template, and `config path`
that prints the resolved location.

## 3. Interactive fallback (huh) — and TTY awareness

When a required argument is missing, prompt instead of erroring — but only when
attached to a terminal. In a pipe or CI, a prompt hangs or silently aborts;
detect that and fail with a clear message.

```go
if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
	msg = args[0]
} else if isInteractive() { // golang.org/x/term: term.IsTerminal(int(os.Stdin.Fd()))
	if err := huh.NewInput().Title("Enter message").Value(&msg).Run(); err != nil {
		return nil // user aborted (Esc/Ctrl+C) — exit cleanly
	}
	msg = strings.TrimSpace(msg)
}
if msg == "" {
	return fmt.Errorf("a message is required")
}
```

## 4. Structured output (`--json-output`)

Machine-readable output makes a CLI scriptable. Two rules:

1. **Detect the JSON flag before cobra parses**, by scanning `os.Args`, so even
   early/parse errors can be emitted as JSON (cobra hasn't populated flags yet
   when those fire).
2. Route *all* output — success *and* errors — through helpers that respect the
   flag, so you never emit half-JSON.

```go
func printJSON(v any) {
	b, _ := json.Marshal(v)
	fmt.Println(string(b))
}

// central exit: JSON error object in JSON mode, plain text otherwise.
func exit(code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	if jsonOutput {
		b, _ := json.Marshal(map[string]any{"error": msg, "exitCode": code})
		fmt.Fprintln(os.Stderr, string(b))
	} else {
		fmt.Fprintf(os.Stderr, "mytool: %s\n", msg)
	}
	panic(&exitError{code}) // recovered in Execute → os.Exit(code), runs defers
}
```

Keep JSON keys stable across versions — they're an API. Booleans like
`upToDate`, `updated`, `devBuild` read better than overloading strings.

## 5. Full-screen notices (bubbletea + lipgloss)

For an alt-screen TUI moment (a notice, a picker):

- Run with `tea.WithAltScreen()`; the screen is restored on quit.
- The alt screen leaves **no scrollback**, so anything the user should keep
  (e.g. the message they typed) must be re-printed to stdout *after* the program
  returns.
- Style with `lipgloss` adaptive colors (`lipgloss.AdaptiveColor{Light, Dark}`)
  so output suits light and dark terminals.
- Accept Space/Enter/Esc/q/Ctrl+C as dismissal; never trap the user.

```go
p := tea.NewProgram(model, tea.WithAltScreen())
if _, err := p.Run(); err != nil { return fmt.Errorf("run program: %w", err) }
fmt.Printf("message: %q\n", msg) // echo to scrollback after teardown
```

## 6. Errors & exit codes

`plaqq` centralizes exits through a `panic(&exitError{code})` that `Execute`
recovers and turns into `os.Exit(code)`. The payoff is one place that emits
JSON-vs-text errors and a chance to run `defer`s before exiting. If you don't
need that, the idiomatic baseline is fine: return errors from `RunE`, set
`SilenceErrors`, and print/`os.Exit` once in `main`. Either way, reserve exit
codes: `0` ok, `1` generic failure, `2` operational (network/install) failure.

## 7. Testing

- **Table tests** for pure logic (version compare, color parsing, config
  layering). See `internal/cmd/update_test.go` for the version-compare cases.
- **Inject side effects via package-level function vars** so commands are
  testable without network or a real install. `update.go` exposes
  `fetchLatestVersionFunc` and `installLatestVersionFn`; tests swap them and
  assert behavior. Restore them with `t.Cleanup`.
- **Drive config through the env override** (`t.Setenv("MYTOOL_CONFIG", tmp)`),
  never the real user config dir.
- Run `gofmt -l`, `go vet`, and `golangci-lint` in CI; the `justfile` wires
  these up locally.

```go
func TestUpdateDevBuildInstalls(t *testing.T) {
	orig := installLatestVersionFn
	t.Cleanup(func() { installLatestVersionFn = orig })
	installed := false
	fetchLatestVersionFunc = func() (string, error) { return "0.3.0", nil }
	installLatestVersionFn = func() error { installed = true; return nil }
	updateYes, version = true, "1a06be9" // bare hash that used to crash compare
	updateCmd.Run(updateCmd, nil)
	if !installed { t.Error("dev build should install the latest release") }
}
```

## 8. Distribution & self-update  ← the deep part

Single binary, GitHub Releases, one-line install, in-place upgrade. Full
scripts and code are in [references/self-update.md](references/self-update.md);
the essentials:

**Version embedding.** One source of truth (`VERSION`), injected at build time.

```
# release: goreleaser fills {{.Version}} from the git tag
ldflags: -s -w -X main.version={{.Version}}

# local dev build (justfile): VERSION file + short hash, marked as a dev build
version := shell('printf "%s-dev+%s" "$(tr -d "[:space:]" < VERSION)" "$(git rev-parse --short HEAD)"')
go build -ldflags "-X main.version={{version}}" -o bin/mytool ./cmd/mytool
```

A release build reports `0.3.1`; a local build reports `0.3.1-dev+1a06be9`. The
`-dev+hash` suffix is what lets the update command recognize a development build
instead of crashing on it.

**Release flow.** Bumping the `VERSION` file is the only manual step:

```
edit VERSION  →  push to main  →  auto-tag.yml (diffs VERSION, creates vX.Y.Z)
              →  release.yml (on tag)  →  goreleaser  →  GitHub Release + assets
```

Assets are named `<tool>_<version>_<os>_<arch>.tar.gz` (`.zip` on Windows) with
a `checksums.txt`. The install scripts and self-update command both rely on that
naming, so keep `.goreleaser.yaml`'s `name_template` and the scripts in sync.

**Atomically replacing a running binary** (the one that bites everyone). You
cannot `cp`-overwrite or untar over a binary that is *currently executing*:
Linux returns `ETXTBSY` ("text file busy"), Windows refuses outright. The fix is
**rename, don't truncate** — write the new binary to a temp file **on the same
filesystem** as the target, then `mv -f` it into place. `rename(2)` swaps the
directory entry without touching the in-use inode, so running copies keep
working and new invocations get the new binary. No `pkill` needed.

```sh
# install.sh — temp MUST be in the install dir (same fs), or mv falls back to a
# truncating copy and ETXTBSY returns.
TMP="$(mktemp "${INSTALL_DIR}/.mytool.XXXXXX")"
tar -xzOf "$ARCHIVE" mytool > "$TMP"
chmod +x "$TMP"
mv -f "$TMP" "$INSTALL_DIR/mytool"
```

On Windows, rename the locked exe aside (`mytool.old.exe`) before dropping the
new one in — see the reference. The `justfile install` recipe uses the same
temp-then-`mv` dance.

**The self-update command.** Fetch the latest release tag, compare versions, and
re-run the installer. The critical correctness point: **guard the comparison.**
A non-release version (empty, `dev`, a bare hash, or a `-dev`/`-dirty` build)
is not semver and must not reach a numeric compare.

```go
func isReleaseVersion(v string) bool { // strict X.Y.Z, optional leading v
	v = strings.TrimPrefix(v, "v")
	p := strings.Split(v, ".")
	if len(p) != 3 { return false }
	for _, n := range p {
		if n == "" { return false }
		if _, err := strconv.Atoi(n); err != nil { return false }
	}
	return true
}
```

Behavior: on a real release, compare and report up-to-date or offer the upgrade.
On a dev build, skip the compare, say "you're on a development build", and offer
to install the latest release anyway (with `--yes`, just do it). Also gate any
*background* "update available" check on `isReleaseVersion` so dev builds don't
make pointless network calls or nag. Full command in the reference.

## Pitfalls checklist

- **Never feed a non-semver version to a numeric compare.** Guard with
  `isReleaseVersion`; bare git hashes (`1a06be9`) and `-dev`/`-dirty` builds are
  not comparable. This is the bug that motivated `plaqq` v0.3.1.
- **Never overwrite a running binary in place.** Rename atomically; the temp
  file must be on the **same filesystem** as the target or cross-device `mv`
  re-introduces `ETXTBSY`.
- **Strip the leading `v` consistently** between git tags (`v0.3.1`) and the
  `VERSION` file (`0.3.1`) everywhere they meet.
- **GitHub API has rate limits.** Honor `GITHUB_TOKEN`/`GH_TOKEN`; give
  background checks a short timeout (~2s) and swallow every error silently.
- **Detect `--json-output` before cobra parses** so early errors are still JSON.
- **huh prompts need a TTY.** Detect non-interactive stdin and fail with a clear
  message instead of hanging.
- **Alt-screen TUIs leave no scrollback** — re-print anything worth keeping
  after the program exits.
- **Keep JSON keys stable** — they're a contract with scripts.
