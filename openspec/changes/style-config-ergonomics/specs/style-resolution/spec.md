# Style Resolution

## ADDED Requirements

### Requirement: Session styling via environment variables

The system SHALL read the styling environment variables `PLAQQ_COLOR`,
`PLAQQ_FONT`, `PLAQQ_BOLD`, `PLAQQ_HINT`, and `PLAQQ_NO_HINT`, applying each set
variable to the corresponding notice setting. An unset or empty variable SHALL be
ignored so the setting falls through to the lower-precedence layer.

#### Scenario: Env var styles the notice

- **WHEN** `PLAQQ_COLOR=alert` and `PLAQQ_FONT=heavy` are exported and `plaqq "hi"` is run with no relevant flags
- **THEN** the notice renders with the `alert` colour and `heavy` font

#### Scenario: Empty var falls through

- **WHEN** `PLAQQ_COLOR=` (empty) is exported and the config file sets `color = "ok"`
- **THEN** the notice uses the `ok` colour from the config file

### Requirement: Resolution precedence

The system SHALL resolve each notice setting by layering sources in increasing
priority: built-in default, user config file, session environment variable, then
CLI flag. A higher-priority source that is set SHALL override a lower one for
that setting independently of the other settings.

#### Scenario: Flag beats env var

- **WHEN** `PLAQQ_COLOR=alert` is exported and `plaqq --color ok "hi"` is run
- **THEN** the notice uses the `ok` colour

#### Scenario: Env var beats config file

- **WHEN** the config file sets `font = "block"` and `PLAQQ_FONT=heavy` is exported
- **THEN** the notice uses the `heavy` font

#### Scenario: Per-setting independence

- **WHEN** the config file sets `color` and `PLAQQ_FONT` sets only the font
- **THEN** the colour comes from the config file and the font from the env var

### Requirement: Invalid environment value is a hard error

The system SHALL reject a styling environment variable whose value is invalid
(an unknown font/colour name, a malformed colour, or a non-boolean for a boolean
variable), failing with a non-zero exit and a message naming the offending
variable.

#### Scenario: Bad boolean

- **WHEN** `PLAQQ_BOLD=maybe` is exported and `plaqq "hi"` is run
- **THEN** the command exits non-zero and the error names `PLAQQ_BOLD`

#### Scenario: Unknown font name

- **WHEN** `PLAQQ_FONT=nonsense` is exported and `plaqq "hi"` is run
- **THEN** the command exits non-zero and the error names `PLAQQ_FONT` and lists the valid fonts

### Requirement: Session setup helper emits shell exports

The system SHALL provide a session mode (`plaqq config --session`) that prints
shell `export` statements for the selected styling to standard output, so that
`eval "$(plaqq config --session)"` configures the current terminal session. The
interactive picker SHALL render to standard error so standard output carries only
the export statements.

#### Scenario: Non-interactive export for scripts

- **WHEN** `plaqq config --session --color alert --font heavy` is run
- **THEN** standard output contains `export PLAQQ_COLOR` and `export PLAQQ_FONT` assignments reflecting those choices

#### Scenario: Eval applies to the session

- **WHEN** the user runs `eval "$(plaqq config --session --color alert)"`
- **THEN** `PLAQQ_COLOR` is set to `alert` in that shell and subsequent `plaqq` calls use it

#### Scenario: Reminder when not captured

- **WHEN** `plaqq config --session` is run with standard output attached to a TTY (not captured by `eval`)
- **THEN** it prints a hint to wrap the command in `eval "$(…)"`
