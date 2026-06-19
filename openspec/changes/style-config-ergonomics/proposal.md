# Per-pane style config & interactive customise

> Builds on the hard-error resolution model from
> [`refine-fonts-and-palette`](../refine-fonts-and-palette/proposal.md). Design
> decisions live in [`design.md`](./design.md).

## Why

plaqq is dropped into many panes at once, and a useful habit is "this pane's
alerts are red/heavy, that pane's are teal/block". Today the only persistence is
a single user-global config file, so per-pane intent has to be retyped as flags
every time. We want **per-pane style that sticks for that terminal session**
without adding friction to the common case — bare `plaqq` should stay a
two-keystroke "type message, go".

The natural fit for "sticks with this pane/session" is **environment variables**:
they inherit to every command in a shell and die with it. Layer them between the
user-global config file and one-off flags, and a single `export` (or a small
helper) styles every `plaqq` in that pane. For one-off variation, the interactive
path grows an opt-in **customise** step so you can pick font + colour for a single
notice without touching any persistent setting.

## What Changes

- **New session env-var layer.** plaqq reads `PLAQQ_COLOR`, `PLAQQ_FONT`,
  `PLAQQ_BOLD`, `PLAQQ_HINT`, `PLAQQ_NO_HINT` and applies them above the config
  file but below CLI flags. Set them in a pane → every `plaqq` there is styled.
- **Resolution precedence becomes:** built-in defaults → user config file →
  session env vars → CLI flags. (`plaqq config` continues to set the user-global
  default — the base for panes with no env override.)
- **Invalid env values are a hard error**, consistent with the flag/config
  behaviour, with a message naming the offending variable.
- **Session setup helper.** `plaqq config --session` emits shell `export`
  statements for the chosen style on stdout (UI on stderr), so
  `eval "$(plaqq config --session)"` styles the current pane. Manual
  `export PLAQQ_*=…` always works as the zero-magic fallback.
- **Interactive customise flow.** Bare `plaqq` (no message arg) prompts for the
  message, then offers a two-option choice — **Confirm** (show it now, the
  default) or **Customise** (pick font + colour for *this* notice, then show).
  The message-argument path (`plaqq "msg"`) stays immediate — no extra prompts.

## Impact

- **Affected specs:** new `style-resolution` and `interactive-prompt` capability
  requirements.
- **Affected code:** `internal/cmd/root.go` (`resolveStyle` gains the env layer;
  bare-invocation flow gains confirm/customise), `internal/cmd/config.go`
  (`--session` export mode), `internal/config/` (env-var names / parsing helper),
  README + AGENTS docs (precedence table, env vars, per-pane recipe).
- **Compatibility:** additive. Existing flags, config file, and `plaqq "msg"`
  behave as before; the new env layer and customise step are opt-in.
