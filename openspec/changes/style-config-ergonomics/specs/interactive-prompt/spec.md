# Interactive Prompt

## ADDED Requirements

### Requirement: Bare invocation prompts for message then confirm or customise

The system SHALL, when invoked with no message argument, prompt for the message
text and then present a two-option choice between **Confirm** (display the notice
now) and **Customise** (adjust the style before displaying). **Confirm** SHALL be
the default selection so the fast path requires no extra navigation.

#### Scenario: Confirm shows the notice

- **WHEN** the user runs bare `plaqq`, enters a message, and chooses Confirm
- **THEN** the notice is displayed using the resolved style (defaults/config/env/flags)

#### Scenario: Confirm is the default

- **WHEN** the message has been entered and the choice is presented
- **THEN** Confirm is the focused/default option

### Requirement: Customise chooses font and colour for one notice

The system SHALL, when the user chooses **Customise**, let them select a font and
a colour, each seeded from the currently resolved style, and then display the
notice using those selections. These selections SHALL apply to the current notice
only and SHALL NOT be written to the config file or environment.

#### Scenario: Customised render

- **WHEN** the user chooses Customise and picks a font and colour
- **THEN** the notice renders with the chosen font and colour

#### Scenario: Customise does not persist

- **WHEN** the user customises a notice and the program exits
- **THEN** the config file and environment are unchanged, and a subsequent bare `plaqq` again starts from the resolved style

### Requirement: Message-argument path is immediate

The system SHALL, when a non-empty message argument is supplied, display the
notice immediately without presenting the confirm/customise choice.

#### Scenario: Argument bypasses the prompt

- **WHEN** the user runs `plaqq "deploy starting"`
- **THEN** the notice is displayed at once with no confirm/customise step

### Requirement: Aborting the prompt exits cleanly

The system SHALL exit without error and without displaying a notice if the user
aborts (Esc / Ctrl-C) at any step of the interactive flow.

#### Scenario: Abort at the choice step

- **WHEN** the user enters a message and then presses Esc at the confirm/customise choice
- **THEN** the program exits cleanly with no notice displayed
