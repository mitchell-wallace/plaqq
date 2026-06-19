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

### Requirement: Invalid environment value warns and continues

The system SHALL, when a styling environment variable holds an invalid value
(unknown font/colour name, malformed colour, or non-boolean for a boolean
variable), print a warning to standard error, ignore that variable, and continue
using the next-lower-precedence source. It SHALL NOT abort.

#### Scenario: Bad boolean warns and falls through

- **WHEN** `PLAQQ_BOLD=maybe` is exported and `plaqq "hi"` is run
- **THEN** a warning naming `PLAQQ_BOLD` is printed to stderr and the notice is still displayed using the resolved bold default

#### Scenario: Unknown font warns and falls through

- **WHEN** `PLAQQ_FONT=nonsense` is exported and `plaqq "hi"` is run
- **THEN** a warning naming `PLAQQ_FONT` is printed to stderr and the notice is displayed using the lower-precedence font

### Requirement: Per-terminal session state

The system SHALL persist a per-terminal-session styling record (at least font and
colour) and read it as the session-state layer. The record SHALL be keyed to the
terminal session so that separate panes do not share it, and a fresh session SHALL
start without one. An invalid value in the record SHALL warn and fall through
rather than abort.

#### Scenario: A fresh session has no state

- **WHEN** `plaqq "hi"` is run in a terminal where no session state has been written
- **THEN** the notice resolves from defaults, config, and env only (no session-state influence)

#### Scenario: Session state styles later invocations

- **WHEN** a session-state record sets `font = "heavy"` for the current terminal
- **THEN** a subsequent `plaqq "hi"` in that terminal renders with the `heavy` font

### Requirement: Resolution precedence

The system SHALL resolve each notice setting by layering sources in increasing
priority: built-in default, user config file, session environment variable,
session state, then CLI flag. A higher-priority source that is set SHALL override
a lower one for that setting independently of the other settings.

#### Scenario: Flag beats session state

- **WHEN** session state sets `color = "alert"` and `plaqq --color ok "hi"` is run
- **THEN** the notice uses the `ok` colour

#### Scenario: Session state beats env var

- **WHEN** `PLAQQ_COLOR=alert` is exported and session state sets `color = "ok"`
- **THEN** the notice uses the `ok` colour

#### Scenario: Env var beats config file

- **WHEN** the config file sets `font = "block"` and `PLAQQ_FONT=heavy` is exported
- **THEN** the notice uses the `heavy` font

#### Scenario: Per-setting independence

- **WHEN** the config file sets `color` and only `PLAQQ_FONT` sets the font
- **THEN** the colour comes from the config file and the font from the env var

### Requirement: Session helper writes and clears session state

The system SHALL provide `plaqq config --session` to write the current terminal's
session-state record, and `plaqq config --session --clear` to remove it. Writing
the record SHALL NOT require shell evaluation or integration.

#### Scenario: Write session state

- **WHEN** `plaqq config --session --color alert --font heavy` is run
- **THEN** the current terminal's session-state record is set to that colour and font, and a subsequent `plaqq "hi"` uses them

#### Scenario: Interactive session write

- **WHEN** `plaqq config --session` is run with no style flags in an interactive terminal
- **THEN** it opens the picker and writes the chosen font/colour to the current terminal's session-state record

#### Scenario: Clear session state

- **WHEN** `plaqq config --session --clear` is run after a session record exists
- **THEN** the record is removed and a subsequent `plaqq "hi"` no longer reflects it
