# Tasks: style-config-ergonomics

> Depends on **refine-fonts-and-palette landing first**: this change references the
> post-refine name sets (`alert`/`heavy` etc.), and its env/session warn-and-skip
> is the complement of refine's flag/config hard-fail. Both edit `resolveStyle`,
> which must end up **source-aware** (flag/config → error; env/session → warn).

## 1. Session env-var layer (warn, don't fail)

- [ ] 1.1 Add an env-reading step to `resolveStyle` (`internal/cmd/root.go`)
  between the config-file and session-state layers: `PLAQQ_COLOR`, `PLAQQ_FONT`,
  `PLAQQ_BOLD`, `PLAQQ_HINT`, `PLAQQ_NO_HINT`. Empty/unset → ignored. Make
  `resolveStyle` track each value's source so flag/config invalids error while
  env/session invalids warn (do not lean on `font.Get`'s silent fallback).
- [ ] 1.2 Parse `PLAQQ_BOLD` / `PLAQQ_NO_HINT` with `strconv.ParseBool`; validate
  font/colour with the existing helpers.
- [ ] 1.3 On any invalid env value: print a warning to stderr naming the variable,
  ignore it, and fall through (never abort). Collect warnings and emit them before
  the interactive form *and* the alt-screen program start — do **not** defer them
  on a channel like the update-notice (`root.go:252`), which would misorder them.
- [ ] 1.4 Tests: each var applies; invalid values warn + fall through (assert
  stderr + that the notice still resolves); empty falls through.

## 2. Per-terminal session-state store

- [ ] 2.1 Add a small store (e.g. `internal/session`) keyed by `os.Getppid()`,
  persisting at least font + colour as TOML under `$XDG_RUNTIME_DIR/plaqq/`
  (fallback `os.TempDir()`); `Load`/`Save`/`Clear`. Write atomically (temp file +
  `rename`, mode `0600`); a malformed/invalid record warns and is ignored.
- [ ] 2.2 Read it as the session-state layer in `resolveStyle` (above env, below
  flags); an invalid stored value warns + falls through (consistent with env).
- [ ] 2.3 Tests: write then load round-trips; precedence vs env and flags; a
  fresh key has no state; clear removes it.

## 3. `plaqq config --session`

- [ ] 3.1 Add `--session` and `--clear` to the `config` command
  (`internal/cmd/config.go`): `--session` writes the session-state record (from
  `--color`/`--font` if given, else via the interactive picker — reuse the
  tolerant picker so an invalid existing session/config value seeds as default
  rather than erroring), `--clear` removes it. No `eval` required.
- [ ] 3.2 Tests: `--session --color/--font` writes the record; `--clear` removes
  it.

## 4. Interactive confirm/customise flow with back-navigation

- [ ] 4.1 Rebuild the bare-invocation path (`root.go`) as one `huh.Form`:
  group 1 = message `Input` + Confirm/Customise `Select` (Confirm default);
  group 2 (font + colour `Select`s, seeded from the resolved style) shown via
  `WithHideFunc` only when Customise is chosen — so the user can `shift+tab` back
  to edit the message.
- [ ] 4.2 On submit with Customise: write font+colour to the session-state store
  AND render this notice with them. On Confirm: render with the resolved style.
- [ ] 4.3 Keep the message-argument path immediate (no choice step); preserve
  clean exit on Esc/Ctrl-C at every step.
- [ ] 4.4 Guard non-interactive use: when there is no message arg and stdin is not
  an interactive TTY (`golang.org/x/term.IsTerminal`) or `--json-output` is set,
  exit via `exit()` with "a message is required" instead of launching any form.
  (Also covers today's latent `huh.NewInput()` hang under `--json-output`.)
- [ ] 4.5 Tests where feasible (selection + persistence plumbing; non-TTY guard
  errors); manual check of the TUI flow incl. back-navigation.

## 5. Docs

- [ ] 5.1 README: precedence table (defaults → config → env → session → flags),
  the `PLAQQ_*` vars and warn-on-bad behaviour, the per-pane recipes
  (`plaqq config --session …`, customise, and manual `export` fallback), and the
  confirm/customise flow.
- [ ] 5.2 AGENTS.md: note the env + session-state layers in `resolveStyle`, the
  `--session` helper, and the session-state store location/keying.
- [ ] 5.3 Docs: clarify `PLAQQ_*` style vars vs the pre-existing `PLAQQ_CONFIG`
  (file path, unrelated), and the session-above-env caveat + `--clear` reset. Note
  this change does **not** bump `VERSION` (auto-tag releases on a VERSION bump).

## 6. Validate

- [ ] 6.1 `gofmt -w`, `go vet ./...`, `go build ./...`, `go test ./...` green.
- [ ] 6.2 Manual: bad `PLAQQ_*` warns + still renders; customise in one pane
  sticks for that pane only; back-navigation edits the message; `--session
  --clear` resets.
- [ ] 6.3 `openspec validate style-config-ergonomics --strict`.
