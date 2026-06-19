# Design: style-config-ergonomics

## Resolution precedence

Insert a session env-var layer into the existing chain (`resolveStyle` in
`internal/cmd/root.go`). Lowest wins are overridden by higher:

```
built-in defaults
  < user config file        (plaqq config — global, all panes)
  < session env vars        (PLAQQ_* — per terminal/pane)
  < CLI flags               (--font/--color/… — this invocation)
  < interactive customise   (this single notice, bare-plaqq path only)
```

Each layer is independent per field: a pane may set only `PLAQQ_COLOR` while font
still comes from the config file. The interactive customise step sits *above*
everything but applies to one render and persists nothing.

## Session env-var schema

Mirror the config fields one-to-one:

| env var | maps to | parse |
|---|---|---|
| `PLAQQ_COLOR` | colour | preset name \| hex \| ANSI index (same as `--color`) |
| `PLAQQ_FONT` | font | font name (same as `--font`) |
| `PLAQQ_BOLD` | bold | `strconv.ParseBool` |
| `PLAQQ_HINT` | hint text | string (verbatim) |
| `PLAQQ_NO_HINT` | hide hint | `strconv.ParseBool` |

An unset/empty var is ignored (falls through to the config file). A *set but
invalid* var is a **hard error** naming the variable
(`PLAQQ_BOLD="maybe": invalid bool`), consistent with the flag/config hard-error
model from `refine-fonts-and-palette`. Tradeoff noted: a stale exported var
breaks every `plaqq` in that pane until fixed — acceptable, and the error says
exactly which var and why. (`PLAQQ_CONFIG` is pre-existing and unrelated — it
selects the config *file path*, not a style; keep it distinct.)

## Why env vars give "per-pane" for free

Each pane runs its own shell process; an `export` there is inherited by every
child (`plaqq`) and dies when the shell exits. New panes start fresh, so they
don't inherit another pane's overrides (tmux/screen `update-environment` aside).
No state file, no pane IDs to track — the shell already scopes it.

## The parent-shell constraint & the session helper

A child process can't mutate its parent shell's environment, so `plaqq` cannot
"set" a session var directly — it can only *emit* shell that the shell evaluates.
Standard pattern (direnv, zoxide, fnm, starship):

- `plaqq config --session` runs the interactive picker **rendered to stderr/TTY**
  (via `WithOutput(os.Stderr)`), then prints the chosen style as
  `export PLAQQ_COLOR=…; export PLAQQ_FONT=…` to **stdout**.
- The user wraps it: `eval "$(plaqq config --session)"`. The picker shows; the
  exports apply to the current pane.
- Non-interactive form for scripts/tests: `plaqq config --session --color alert
  --font heavy` skips the picker and just prints the exports.
- If stdout is a TTY (i.e. *not* captured by `eval`), print a one-line hint
  reminding the user to wrap the call in `eval "$(…)"`, so a bare
  `plaqq config --session` isn't silently useless.
- Manual `export PLAQQ_COLOR=alert PLAQQ_FONT=heavy` is always documented as the
  zero-magic fallback.

`plaqq config` with no flags keeps writing the **user-global config file**
unchanged — it is the per-pane base, not a session setter.

## Interactive customise flow (bare `plaqq`, no message arg)

Today: prompt for message → render. New:

1. `huh.Input` — message (unchanged).
2. A two-option choice — **Confirm** (default, focused) or **Customise**.
3. Confirm → render with the resolved style.
4. Customise → `huh.Select` font + `huh.Select` colour, each seeded from the
   resolved style → render with those for this notice only.

Esc/Ctrl-C at any step exits cleanly (current behaviour). Confirm being the
default keeps the fast path to ~two keystrokes (Enter through message, Enter on
Confirm). `plaqq "message"` (arg present) bypasses all of this and renders
immediately — the customise step never gates the scripted/quick path.

Customise only offers font + colour (the high-value, visual choices); bold/hint
stay on flags/config/env to keep the picker to two fields.

## Open questions

- **Session helper surface:** `plaqq config --session` (chosen here) vs a
  dedicated `plaqq env` subcommand. Leaning `--session` to keep one config entry
  point; revisit if it muddies `plaqq config`.
- **Should customise offer "save"?** It can't export to the parent shell, but it
  could offer "save as user default" (write config) at the end. Deferred — keep
  v1 one-shot to avoid surprising persistence.
- **Persisting hint/bold per session** is supported via env vars but not via the
  customise picker; revisit if users want them there.
