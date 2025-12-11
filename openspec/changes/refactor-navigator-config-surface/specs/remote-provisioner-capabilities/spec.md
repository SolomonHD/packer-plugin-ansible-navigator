# remote-provisioner-capabilities Specification Deltas

## ADDED Requirements

### Requirement: Navigator Config File Generation

The SSH-based provisioner SHALL support generating ansible-navigator.yml configuration files from a declarative HCL map and using them to control all ansible-navigator behavior.

#### Scenario: User-specified navigator_config generates YAML file

- **GIVEN** a configuration with:

  ```hcl
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
    }
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
        }
      }
    }
  }
  ```

- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL generate a temporary file named `/tmp/packer-navigator-cfg-<uuid>.yml` (or equivalent in system temp directory)
- **AND** the file SHALL contain valid YAML matching the ansible-navigator.yml schema
- **AND** the file path SHALL be added to the cleanup list

#### Scenario: ANSIBLE_NAVIGATOR_CONFIG environment variable set

- **GIVEN** a generated ansible-navigator.yml file at path `/tmp/packer-navigator-cfg-ABC123.yml`
- **WHEN** the provisioner executes ansible-navigator
- **THEN** it SHALL set `ANSIBLE_NAVIGATOR_CONFIG=/tmp/packer-navigator-cfg-ABC123.yml` in the command environment
- **AND** the environment variable SHALL be present for both version checks and play execution

#### Scenario: Cleanup after provisioning

- **GIVEN** a generated ansible-navigator.yml file
- **WHEN** provisioning completes (success or failure)
- **THEN** the temporary ansible-navigator.yml file SHALL be deleted
- **AND** the cleanup SHALL occur in a deferred function to ensure execution

#### Scenario: Automatic EE defaults when execution environment enabled

- **GIVEN** a configuration with:

  ```hcl
  navigator_config = {
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```

- **AND** the user has NOT specified `ansible.config.defaults.remote_tmp` or `ansible.config.defaults.local_tmp`
- **AND** the user has NOT specified `execution-environment.environment-variables`
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** it SHALL automatically include:

  ```yaml
  ansible:
    config:
      defaults:
        remote_tmp: "/tmp/.ansible/tmp"
        local_tmp: "/tmp/.ansible-local"
  execution-environment:
    environment-variables:
      ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
      ANSIBLE_LOCAL_TMP: "/tmp/.ansible-local"
  ```

- **AND** these defaults SHALL prevent "Permission denied: /.ansible/tmp" errors in EE containers

#### Scenario: User-provided values override automatic defaults

- **GIVEN** a configuration with `execution-environment.enabled = true`
- **AND** the user has explicitly specified temp directory settings in navigator_config
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** it SHALL use the user-specified values
- **AND** it SHALL NOT apply automatic defaults
- **AND** the user's configuration SHALL take full precedence

#### Scenario: No config file generation when navigator_config not specified

- **GIVEN** a configuration without `navigator_config`
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL NOT generate an ansible-navigator.yml file
- **AND** it SHALL NOT set the ANSIBLE_NAVIGATOR_CONFIG environment variable
- **AND** ansible-navigator SHALL use its normal configuration search order

#### Scenario: Empty navigator_config rejected

- **GIVEN** a configuration with `navigator_config = {}`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail with a clear error message
- **AND** the error SHALL explain that navigator_config must contain at least one configuration section

#### Scenario: Complex nested structure preserved

- **GIVEN** a configuration with deeply nested navigator_config:

  ```hcl
  navigator_config = {
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
          host_key_checking = "False"
        }
        ssh_connection = {
          pipelining = "True"
          timeout = "30"
        }
      }
    }
    execution-environment = {
      enabled = true
      image = "custom:latest"
      environment-variables = {
        ANSIBLE_REMOTE_TMP = "/custom/tmp"
        CUSTOM_VAR = "value"
      }
    }
  }
  ```

- **WHEN** the YAML file is generated
- **THEN** the nested structure SHALL be preserved exactly
- **AND** all keys and values SHALL be written correctly

#### Scenario: Navigator config works with SSH proxy

- **GIVEN** a configuration with both `navigator_config` and SSH proxy settings
- **WHEN** the provisioner executes with an execution environment
- **THEN** the EE container SHALL use the generated ansible-navigator.yml
- **AND** the EE container SHALL be able to reach the SSH proxy via the configured `ansible_proxy_host`
- **AND** all ansible.cfg settings in navigator_config SHALL apply inside the container

## REMOVED Requirements

### Requirement: Execution Environment Support (via string field)

**Reason**: Configuration via string `execution_environment` field is replaced by `navigator_config.execution-environment` object.

**Migration**: Users should move from:

```hcl
execution_environment = "quay.io/ansible/creator-ee:latest"
```

To:

```hcl
navigator_config = {
  execution-environment = {
    enabled = true
    image = "quay.io/ansible/creator-ee:latest"
  }
}
```

### Requirement: Navigator Mode Support (via string field)

**Reason**: Configuration via `navigator_mode` field is replaced by `navigator_config.mode`.

**Migration**: Users should move from:

```hcl
navigator_mode = "stdout"
```

To:

```hcl
navigator_config = {
  mode = "stdout"
}
```

### Requirement: Ansible Configuration File Generation (via ansible_cfg map)

**Reason**: Configuration via `ansible_cfg` map is replaced by `navigator_config.ansible.config`.

**Migration**: Users should move from:

```hcl
ansible_cfg = {
  defaults = {
    remote_tmp = "/tmp/.ansible/tmp"
  }
}
```

To:

```hcl
navigator_config = {
  ansible = {
    config = {
      defaults = {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }
}
```

## MODIFIED Requirements

### Requirement: SSH-Based Remote Execution

The SSH-based `ansible-navigator` provisioner SHALL run ansible-navigator from the local machine (where Packer is executed) and connect to the target via SSH.

#### Scenario: Default execution mode

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **WHEN** the provisioner executes
- **THEN** it SHALL run on the local machine (not the target)
- **AND** it SHALL connect to the target via SSH using the configured communicator
- **AND** this matches the behavior of the official `ansible` provisioner

#### Scenario: Execution environment support via navigator_config

- **GIVEN** a configuration with:

  ```hcl
  navigator_config = {
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
    }
  }
  ```

- **WHEN** the provisioner executes
- **THEN** ansible-navigator SHALL use the specified execution environment
- **AND** all settings SHALL be controlled via the generated ansible-navigator.yml file
- **AND** the container SHALL have network access to reach the target via SSH

#### Scenario: No execution environment specified

- **GIVEN** a configuration without `navigator_config.execution-environment`
- **WHEN** the provisioner executes
- **THEN** ansible-navigator SHALL use its default execution environment behavior
- **AND** no EE-specific settings SHALL be in the generated config file

### Requirement: Default Navigator Command

The SSH-based provisioner SHALL use ansible-navigator as its default executable, pass `run` as the first argument, and treat the `command` field strictly as the ansible-navigator executable name or path (without embedded arguments).

#### Scenario: Default command when unspecified

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **AND** the `command` field is not specified
- **WHEN** the provisioner constructs the command to run ansible-navigator
- **THEN** it SHALL invoke ansible-navigator using `exec.Command("ansible-navigator", "run", ...)` on the local machine
- **AND** configuration SHALL be controlled via ANSIBLE_NAVIGATOR_CONFIG environment variable
- **AND** NO mode, EE, or other CLI flags SHALL be passed (all settings come from config file)

#### Scenario: Command construction with navigator_config

- **GIVEN** a configuration with `navigator_config` specified
- **WHEN** the provisioner constructs the ansible-navigator command
- **THEN** it SHALL invoke: `ansible-navigator run <playbook>` with ANSIBLE_NAVIGATOR_CONFIG env var set
- **AND** it SHALL NOT pass `--mode`, `--ee`, `--eei`, or other flags that are controlled by the config file

### Requirement: Navigator Mode Support (via config file)

The SSH-based provisioner SHALL support configuring the ansible-navigator execution mode via navigator_config.

#### Scenario: Mode specified in navigator_config

- **GIVEN** a configuration with:

  ```hcl
  navigator_config = {
    mode = "json"
  }
  ```

- **AND** `structured_logging = true`
- **WHEN** the provisioner executes
- **THEN** ansible-navigator SHALL use JSON mode based on the config file
- **AND** the provisioner SHALL parse JSON events and provide detailed task-level reporting
