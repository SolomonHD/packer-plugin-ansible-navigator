## ADDED Requirements

### Requirement: Per-Play extra_args escape hatch (remote provisioner)

The SSH-based provisioner SHALL support a per-play `extra_args` field (list(string)) that is appended to the `ansible-navigator run` invocation for that play.

#### Scenario: play.extra_args is accepted in HCL schema

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- **AND** a `play {}` block includes `extra_args = ["--check", "--diff"]`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the provisioner configuration SHALL include the `extra_args` values for that play

#### Scenario: play.extra_args is passed verbatim

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with a play:

  ```hcl
  play {
    target     = "site.yml"
    extra_args = ["--check", "--diff"]
  }
  ```

- **WHEN** the provisioner constructs the ansible-navigator command for that play
- **THEN** it SHALL include both `--check` and `--diff` as command arguments
- **AND** it SHALL not rewrite, split, or validate the `extra_args` values beyond basic type handling

#### Scenario: Deterministic argument ordering includes extra_args

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with:
  - one `play {}` block
  - `navigator_config.mode = "stdout"`
  - `play.extra_args = ["--check", "--diff"]`
- **WHEN** the provisioner constructs the ansible-navigator command arguments
- **THEN** the argument ordering SHALL be deterministic and consistent across executions:
  1. `run` subcommand
  2. enforced mode flag behavior (when configured), inserted immediately after `run`
  3. play-level `extra_args`
  4. provisioner-generated arguments (inventory, extra vars, tags, etc.)
  5. the play target (playbook path or generated playbook/role invocation)

