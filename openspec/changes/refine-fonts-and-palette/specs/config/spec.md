# Config

## ADDED Requirements

### Requirement: Interactive picker tolerates invalid stored values

The interactive `plaqq config` picker SHALL open and remain usable even when the
config file holds a font or colour name that is no longer valid. An invalid
stored value SHALL be presented as the default selection rather than causing an
error, so the picker can be used to repair the config.

#### Scenario: Stored font no longer exists

- **WHEN** the config file sets a removed font name and `plaqq config` is launched
- **THEN** the picker opens with the font field defaulted to `block` (no error)

#### Scenario: Stored colour no longer exists

- **WHEN** the config file sets a removed colour name and `plaqq config` is launched
- **THEN** the picker opens with the colour field defaulted to `info` (no error)

#### Scenario: Saving repairs the config

- **WHEN** the user selects valid options in the picker and saves
- **THEN** the written config contains only valid font/colour values
