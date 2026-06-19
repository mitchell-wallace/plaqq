# Per-pane style config & interactive customise

> Builds on the resolution model from
> [`refine-fonts-and-palette`](../refine-fonts-and-palette/proposal.md). Design
> decisions live in [`design.md`](./design.md).

## Why

plaqq is dropped into many panes at once, and a useful habit is "this pane's
alerts are red/heavy, that pane's are teal/block". Today the only persistence is
a single user-global config file, so per-pane intent has to be retyped as flags
every time. We want **per-pane style that sticks for that terminal session**
without adding friction to the common case — bare `plaqq` should stay a
two-keystroke "type message, go".

Environment variables fit "ambient, per-shell" overrides. And the interactive
path should let you style *one* notice (or make it stick) right from the prompt,
with the message still editable — so a quick "customise" picks font + colour and
quietly makes that the session's look.

## What Changes

- **New session env-var layer.** plaqq reads `PLAQQ_COLOR`, `PLAQQ_FONT`,
  `PLAQQ_BOLD`, `PLAQQ_HINT`, `PLAQQ_NO_HINT` between the config file and CLI
  flags. **A bad env value warns and is ignored** (falls through) — it never
  aborts. (CLI flags, by contrast, hard-fail with the valid options + a
  nearest-match suggestion; the config file stays hard-fail too.)
- **Resolution precedence becomes:** defaults → user config file → session env
  vars → session state → CLI flags.
- **Customise sticks for the session.** Choosing *Customise* writes the chosen
  font + colour to a per-terminal **session-state file** (a child process can't
  set its parent shell's env vars, so this file is plaqq's honest stand-in), so
  the next `plaqq` in that pane reuses them without re-customising.
- **`plaqq config --session`** writes that session-state file directly (no `eval`
  needed); `--clear` resets it. `plaqq config` (no flag) still sets the
  user-global default.
- **Interactive flow with back-navigation.** Bare `plaqq` (no message arg)
  prompts for the message, then a two-option choice — **Confirm** (show it now,
  default) or **Customise** (pick font + colour). Built as one form so the user
  can navigate **back to edit the message** after entering customise. The
  message-argument path (`plaqq "msg"`) stays immediate.

## Impact

- **Affected specs:** new `style-resolution` and `interactive-prompt` capability
  requirements.
- **Affected code:** `internal/cmd/root.go` (`resolveStyle` gains env +
  session-state layers, env warnings; bare-invocation flow becomes a two-group
  form; customise writes session state), `internal/cmd/config.go` (`--session`
  write/clear), new session-state store keyed by shell PID under
  `$XDG_RUNTIME_DIR`/temp, `internal/config/` (env-var names / parsing),
  README + AGENTS docs (precedence table, env vars, per-pane recipe).
- **Compatibility:** additive. Existing flags, config file, and `plaqq "msg"`
  behave as before; the env layer, session state, and customise step are opt-in.
