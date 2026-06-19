# Interactive Prompt

## ADDED Requirements

### Requirement: Bare invocation prompts for message then confirm or customise

The system SHALL, when invoked with no message argument, prompt for the message
text and then present a two-option choice between **Confirm** (display the notice
now) and **Customise** (adjust the style before displaying). **Confirm** SHALL be
the default selection so the fast path requires no extra navigation.

#### Scenario: Confirm shows the notice

- **WHEN** the user runs bare `plaqq`, enters a message, and chooses Confirm
- **THEN** the notice is displayed using the resolved style (defaults/config/env/session/flags)

#### Scenario: Confirm is the default

- **WHEN** the message has been entered and the choice is presented
- **THEN** Confirm is the focused/default option

### Requirement: Customise chooses font and colour and sticks for the session

The system SHALL, when the user chooses **Customise**, let them select a font and
a colour (each seeded from the currently resolved style), display the notice using
those selections, and write them to the session-state layer. Subsequent
invocations in the same terminal session SHALL therefore use the customised font
and colour without the user re-customising.

#### Scenario: Customised render

- **WHEN** the user chooses Customise and picks a font and colour
- **THEN** the notice renders with the chosen font and colour

#### Scenario: Customise sticks for the session

- **WHEN** the user customises a notice with a font and colour, then later runs `plaqq "next"` in the same terminal session
- **THEN** the second notice renders with the same font and colour without prompting to customise

#### Scenario: A separate pane is unaffected

- **WHEN** the user customises in one pane and runs `plaqq "hi"` in a different pane
- **THEN** the second pane renders from its own resolved style, not the first pane's customisation

### Requirement: Customise allows editing the message

The system SHALL allow the user, after choosing **Customise**, to navigate back to
the message field and change it before the notice is displayed.

#### Scenario: Edit message after entering customise

- **WHEN** the user enters a message, chooses Customise, navigates back, edits the message, and proceeds
- **THEN** the notice displays the edited message with the customised style

### Requirement: Message-argument path is immediate

The system SHALL, when a non-empty message argument is supplied, display the
notice immediately without presenting the confirm/customise choice.

#### Scenario: Argument bypasses the prompt

- **WHEN** the user runs `plaqq "deploy starting"`
- **THEN** the notice is displayed at once with no confirm/customise step

### Requirement: Non-interactive invocation requires a message

The system SHALL, when invoked with no message argument in a context where the
interactive prompt cannot run — standard input is not an interactive terminal, or
`--json-output` is set — exit with a non-zero status and an error instructing the
user to supply a message argument, instead of launching the interactive form.

#### Scenario: Piped stdin with no message

- **WHEN** `echo | plaqq` is run (stdin not a TTY) with no message argument
- **THEN** the program exits non-zero with an error asking for a message argument and does not launch the form

#### Scenario: JSON output with no message

- **WHEN** `plaqq --json-output` is run with no message argument
- **THEN** the program exits non-zero (structured error) without launching the form

### Requirement: Aborting the prompt exits cleanly

The system SHALL exit without error and without displaying a notice if the user
aborts (Esc / Ctrl-C) at any step of the interactive flow.

#### Scenario: Abort at the choice step

- **WHEN** the user enters a message and then presses Esc at the confirm/customise choice
- **THEN** the program exits cleanly with no notice displayed
