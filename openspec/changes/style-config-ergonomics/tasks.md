# Tasks: style-config-ergonomics

## 1. Session env-var layer

- [ ] 1.1 Add an env-reading step to `resolveStyle` (`internal/cmd/root.go`)
  between the config-file and flag layers: `PLAQQ_COLOR`, `PLAQQ_FONT`,
  `PLAQQ_BOLD`, `PLAQQ_HINT`, `PLAQQ_NO_HINT`. Empty/unset → ignored.
- [ ] 1.2 Parse `PLAQQ_BOLD` / `PLAQQ_NO_HINT` with `strconv.ParseBool`; on
  failure return a hard error naming the variable.
- [ ] 1.3 Reuse the existing font/colour validation so invalid `PLAQQ_FONT` /
  `PLAQQ_COLOR` hard-error and name the variable + valid options.
- [ ] 1.4 Tests: each var applies; precedence (flag > env > config > default) per
  setting; per-setting independence; invalid env values error with the var name.

## 2. Session setup helper

- [ ] 2.1 Add `--session` to the `config` command (`internal/cmd/config.go`):
  render the picker via `WithOutput(os.Stderr)`, then print
  `export PLAQQ_…=…` lines to stdout.
- [ ] 2.2 Support a non-interactive form (`--session` + `--color/--font/…`) that
  skips the picker and prints exports directly (scriptable + testable).
- [ ] 2.3 When stdout is a TTY (not captured by `eval`), print a hint to wrap the
  call in `eval "$(plaqq config --session)"`.
- [ ] 2.4 Tests: non-interactive `--session` emits correct `export` statements;
  quoting is shell-safe (use `strconv.Quote`/`%q` appropriately).

## 3. Interactive confirm/customise flow

- [ ] 3.1 In the bare-invocation path (`root.go`), after the message input, add a
  two-option `huh` choice (Confirm default / Customise).
- [ ] 3.2 On Customise, present font + colour `huh.Select`s seeded from the
  resolved style; apply the selections to this render only.
- [ ] 3.3 Keep the message-argument path immediate (no choice step); preserve
  clean exit on Esc/Ctrl-C at every step.
- [ ] 3.4 Tests where feasible (selection plumbing); manual check of the TUI flow.

## 4. Docs

- [ ] 4.1 README: precedence table (defaults → config → env → flags), the
  `PLAQQ_*` vars, the per-pane recipe (`eval "$(plaqq config --session)"` and the
  manual `export` fallback), and the confirm/customise flow.
- [ ] 4.2 AGENTS.md: note the env layer in `resolveStyle` and the `--session`
  helper. Update the config-file template comment to mention env overrides.

## 5. Validate

- [ ] 5.1 `gofmt -w`, `go vet ./...`, `go build ./...`, `go test ./...` green.
- [ ] 5.2 Manual: export `PLAQQ_*` in one pane, confirm it styles `plaqq` there
  and not in a fresh pane; run the confirm/customise flow end-to-end.
- [ ] 5.3 `openspec validate style-config-ergonomics --strict`.
