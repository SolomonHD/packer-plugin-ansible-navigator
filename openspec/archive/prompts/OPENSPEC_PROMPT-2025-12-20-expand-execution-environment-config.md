# OpenSpec Change Prompt: Expand Execution Environment Configuration

## Context

The plugin currently supports a limited subset of ansible-navigator's `execution-environment` configuration block. Users receive "Unsupported argument" errors when trying to configure valid ansible-navigator v3.x EE options like `container-options`, `volume-mounts`, or `pull-arguments`. The current [`NavigatorConfig.ExecutionEnvironment`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct only covers basic fields (enabled, image, pull_policy) and basic environment variable support.

## Goal

Expand the [`ExecutionEnvironment`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct to support all documented ansible-navigator v3.x execution-environment configuration options, providing full HCL configuration parity with the ansible-navigator.yml `execution-environment` section.

## Scope

**In scope:**
- Add missing execution-environment fields to the Go struct:
  - `container_engine` (string: "auto", "podman", "docker")
  - `container_options` (list of strings for additional container runtime flags)
  - `volume_mounts` (list of volume mount specifications)
  - `pull_arguments` (list of strings for pull command flags)
  - Complete `environment_variables` structure (currently has partial support via `pass`/`set` but needs validation)
- Update `//go:generate packer-sdc mapstructure-to-hcl2` type list to include any new nested struct types
- Regenerate HCL2 spec files ([`*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/*.hcl2spec.go:1))
- Update YAML generation in [`generateNavigatorYAML()`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) to correctly emit all new execution-environment fields with proper hyphenation and nesting
- Add tests for HCL decoding and YAML generation of execution-environment options

**Out of scope:**
- Expanding other navigator_config sections (ansible_config, logging, playbook-artifact, mode-settings, etc.) - those are separate changes
- Changing existing automatic defaults for execution environments (safe temp directories, collections mounting)
- Adding runtime EE validation beyond what ansible-navigator itself provides

## Desired Behavior

A user can configure any documented ansible-navigator v3.x execution-environment option through [`navigator_config { execution_environment { ... } }`](packer/plugins/packer-plugin-ansible-navigator/README.md:1) in Packer HCL:

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    execution_environment {
      enabled           = true
      image             = "quay.io/ansible/creator-ee:latest"
      pull_policy       = "missing"
      container_engine  = "docker"
      
      container_options = [
        "--net=host",
        "--security-opt", "label=disable"
      ]
      
      volume_mounts = [
        {
          src  = "/host/path"
          dest = "/container/path"
          options = "ro"
        }
      ]
      
      pull_arguments = [
        "--creds", "user:pass"
      ]
      
      environment_variables {
        pass = ["SSH_AUTH_SOCK"]
        set = {
          ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        }
      }
    }
  }
}
```

The generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) correctly includes:

```yaml
execution-environment:
  enabled: true
  image: quay.io/ansible/creator-ee:latest
  pull:
    policy: missing
    arguments:
      - "--creds"
      - "user:pass"
  container-engine: docker
  container-options:
    - "--net=host"
    - "--security-opt"
    - "label=disable"
  volume-mounts:
    - src: /host/path
      dest: /container/path
      options: ro
  environment-variables:
    pass:
      - SSH_AUTH_SOCK
    set:
      ANSIBLE_REMOTE_TMP: /tmp/.ansible/tmp
```

## Constraints & Assumptions

- **Assumption:** ansible-navigator v3.x is the target version (not experimental/undocumented options).
- **Constraint:** HCL field names use underscores (`container_options`, `volume_mounts`), YAML keys use hyphens (`container-options`, `volume-mounts`).
- **Constraint:** The `pull` section in YAML has nested `policy` and `arguments` - the struct and YAML generation must correctly represent this nesting.
- **Constraint:** Volume mounts may need a typed struct (with `src`, `dest`, `options` fields) rather than plain strings - determine correct structure from ansible-navigator v3.x schema.
- **Constraint:** Preserve backward compatibility - existing minimal execution_environment configs must continue to work.
- **Constraint:** Use typed structs (no `map[string]interface{}` with `cty.DynamicPseudoType`) for RPC serialization compatibility.
- **Constraint:** Follow the plugin's existing pattern: struct fields use `mapstructure` tags, HCL2 specs are generated via `packer-sdc`.

## Acceptance Criteria

- [ ] [`ExecutionEnvironment`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct includes all ansible-navigator v3.x execution-environment documented fields.
- [ ] HCL configuration with `container_options`, `volume_mounts`, and `pull_arguments` decodes without error.
- [ ] Generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) includes correct key names (hyphenated) and nesting (e.g., `pull.policy`, `pull.arguments`).
- [ ] A minimal Packer template using new execution-environment options passes `packer validate`.
- [ ] Existing execution-environment configurations continue to work unchanged (backward compatibility test).
- [ ] [`make generate`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) regenerates HCL2 specs successfully.
- [ ] [`go build ./...`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) compiles without errors.
- [ ] Unit tests validate HCL decoding and YAML structure for representative execution-environment configs.

## Files Expected to Change

- [`provisioner/ansible-navigator/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) (struct definitions, YAML generation)
- [`provisioner/ansible-navigator-local/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:1) (if separate, likely same structs)
- [`provisioner/ansible-navigator/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/*.hcl2spec.go:1) (generated)
- [`provisioner/ansible-navigator-local/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/*.hcl2spec.go:1) (generated)
- Test files under `provisioner/*/` (new tests for execution-environment expansion)
- Possibly [`example/`](packer/plugins/packer-plugin-ansible-navigator/example/) HCL templates (to demonstrate new options)

## Dependencies

None (this is the first in a series of navigator_config expansions).
