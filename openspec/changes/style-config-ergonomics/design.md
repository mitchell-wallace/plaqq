# Design: style-config-ergonomics

## Resolution precedence

`resolveStyle` (`internal/cmd/root.go`) layers sources per setting, lowest to
highest priority:

```
built-in defaults
  < user config file   (plaqq config — global, all panes)
  < session env vars    (PLAQQ_* — manual/ambient, this shell)
  < session state       (customise / plaqq config --session — this terminal)
  < CLI flags           (--font/--color/… — this invocation)
```

The interactive **customise** step applies to the current render *and* writes the
session-state layer, so it both styles the notice now and sticks for the rest of
the session. Each setting resolves independently (a pane may take colour from an
env var and font from the config file).

Session env vars and session state are both "this terminal" scope; session state
sits **above** env vars so that a customise choice is authoritative for the
session even if an env var is also set (the common case has no env var, so
customise simply sticks). A one-off `--flag` still overrides for that call.

## Strictness tracks authorship

How a bad value is handled depends on how deliberately the user aimed it at plaqq:

| source | on invalid value | why |
|---|---|---|
| CLI flag | **hard error** — list valid options + nearest-match suggestion | typed this instant; immediate feedback wanted |
| user config file | **hard error** — list valid options | authored for plaqq; `plaqq config` repairs it |
| session env var | **warn to stderr, ignore, fall through** | ambient/inherited, may be stale; must never brick the pane |
| session state | **warn to stderr, ignore, fall through** | machine-written, ephemeral |

Flags and the config file — things the user wrote *for* plaqq — fail loudly;
env vars and the session-state file — ambient or auto-written — degrade
gracefully. This keeps `refine-fonts-and-palette`'s hard-error stance for
flags/config while making the session layers forgiving, as requested. (An env var
is often inherited from a shell rc or another tool, so it shouldn't be able to
break an unrelated `plaqq`; a config file is a document the user deliberately
wrote and the picker can fix.) Warnings print to stderr before the interactive
form *and* the alt-screen program launch (not deferred on a channel like the
update-notice), so they land in scrollback rather than corrupting the TUI.

**`resolveStyle` becomes source-aware.** Because both the fail-loud and the
warn-and-skip paths now live inside resolution, `resolveStyle` must validate each
value *knowing which layer it came from*: an invalid flag/config value is returned
as an error; an invalid env/session value is warned and skipped. It therefore
can't rely on `font.Get`'s silent fallback or a single post-resolution check —
validation moves into the layering. This is the key seam shared with
`refine-fonts-and-palette` (which moves font/colour validation into resolution);
the two changes touch the same function and must agree that it is per-layer.

## Session env-var schema

Mirrors the config fields one-to-one:

| env var | maps to | parse |
|---|---|---|
| `PLAQQ_COLOR` | colour | preset \| hex \| ANSI index (same as `--color`) |
| `PLAQQ_FONT` | font | font name (same as `--font`) |
| `PLAQQ_BOLD` | bold | `strconv.ParseBool` |
| `PLAQQ_HINT` | hint text | string (verbatim) |
| `PLAQQ_NO_HINT` | hide hint | `strconv.ParseBool` |

Unset/empty → ignored (fall through). Invalid → warn + ignore. `PLAQQ_CONFIG` is
pre-existing and unrelated (it selects the config *file path*, not a style).

## "Customise sets the session" — honestly, a session-state file

A child process cannot mutate its parent shell's environment, so `plaqq` cannot
set an env var that a *later* `plaqq` in the same shell would inherit — and the
customise step couldn't anyway, since it ends by rendering the notice, not by
emitting shell. So "customise sets the env vars" is delivered by a small
**plaqq-managed session-state file** read as the session-state layer.
Functionally identical to a session env var: set it once, every later `plaqq` in
that terminal reuses it.

- **Keyed per terminal session** by the parent shell PID (`os.Getppid()`): every
  `plaqq` typed at one shell prompt shares it; a new pane (new shell) starts
  clean. (The controlling TTY is a more precise key; PPID is zero-dep and good
  enough — a possible refinement. Edge: PPID reuse after the shell exits — stamp
  the file with the shell start time to detect it, else accept + `--clear`.)
- **Stored** under `$XDG_RUNTIME_DIR/plaqq/` (cleared at logout) when present,
  else `os.TempDir()`; filename includes the session key.
- **Written by** the customise step and by `plaqq config --session`, which writes
  the file *directly* — no `eval`/shell integration needed, sidestepping the
  parent-shell problem entirely.
- **Cleared by** `plaqq config --session --clear` (and naturally when the runtime
  dir is wiped at logout).
- **Written atomically** (temp file + `rename`) with mode `0600`; on read, a
  malformed/invalid record warns and is ignored (it never blocks rendering).

Caveat of session-state-above-env: once customise has written session state, a
later `export PLAQQ_COLOR=…` in that pane appears to do nothing (session state
wins) until `--clear` or a fresh pane. That is the cost of "customise is
authoritative for the session"; it's documented, and `--clear` resets it. (The
alternative ordering — env above session — was rejected because it would stop a
customise from sticking whenever any `PLAQQ_*` is exported, defeating the feature.)

For users who specifically want real exported env vars (e.g. to share with other
tools), manual `export PLAQQ_*` still works; an optional `--export` form of
`plaqq config --session` that prints `export` lines for `eval` is an open
question, not core.

## Interactive flow (bare `plaqq`, no message arg), with back-navigation

One `huh.Form` with two groups, so the user can move *back* to the message after
entering customise (sequential `Run()` calls could not):

- **Group 1:** message `Input` + a Confirm/Customise `Select` (Confirm focused by
  default).
- **Group 2** (`WithHideFunc`, shown only when Customise is chosen): font
  `Select` + colour `Select`, each seeded from the fully-resolved style.

Confirm → group 2 stays hidden; submitting renders with the resolved style (fast
path ≈ Enter, Enter). Customise → group 2 appears; the user can `shift+tab` back
to group 1 to edit the message, then forward again. On submit with Customise:
write the chosen font+colour to the session-state file **and** render this notice
with them. Esc/Ctrl-C at any step exits cleanly. `plaqq "message"` (arg present)
bypasses the form and renders immediately.

Because the resolved style now includes the session-state layer, a later bare
`plaqq` that just confirms — or a `plaqq "msg"` — automatically shows the
customised font/colour without re-customising. Customise offers only font +
colour (the visual choices); bold/hint stay on flags/config/env to keep the form
to two fields.

## Non-interactive / `--json-output`

The confirm/customise form (like today's message prompt) needs an interactive
TTY. When bare `plaqq` is run with no message argument and stdin is **not** an
interactive TTY (piped, CI) or `--json-output` is set, plaqq must **not** launch
the form — it exits with a clear error telling the user to pass a message
argument. This also fixes a pre-existing latent bug: today `huh.NewInput()`
(`root.go:189`) runs regardless of `--json-output`/TTY and will fail or hang.
Detect interactivity with `golang.org/x/term.IsTerminal` (already an indirect
dep via charm).

## Open questions

- Session key: PPID (chosen) vs controlling TTY vs an injected `PLAQQ_SESSION` id.
- Optional `plaqq config --session --export` (emit `export` lines for `eval`) for
  users who want real env vars rather than the state file.
- Whether customise should also offer "save as user default" (write the config).
