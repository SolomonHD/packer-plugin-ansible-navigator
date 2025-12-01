# Design: Align Local Provisioner with Remote Provisioner

## Overview

This document describes the architectural approach for aligning the local provisioner (`ansible-navigator`) with the remote provisioner (`ansible-navigator-remote`) to provide consistent modern Ansible Navigator functionality across both execution modes.

## Current Architecture

### Local Provisioner (ansible-navigator)

```
provisioner/ansible-navigator/
├── provisioner.go      # Legacy Config struct, ansible-playbook default
├── provisioner_test.go
└── provisioner.hcl2spec.go
```

**Current characteristics:**
- Default command: `ANSIBLE_FORCE_COLOR=1 PYTHONUNBUFFERED=1 ansible-playbook`
- Config struct: ~15 fields (legacy layout)
- No Play struct
- No validation for mutual exclusivity
- No GalaxyManager integration
- Executes commands remotely via communicator

### Remote Provisioner (ansible-navigator-remote)

```
provisioner/ansible-navigator-remote/
├── provisioner.go      # Modern Config struct, ansible-navigator run
├── provisioner_test.go
├── provisioner.hcl2spec.go
├── galaxy.go           # GalaxyManager for dependency management
├── galaxy_test.go
├── json_logging.go     # Structured logging support
└── flat.go
```

**Current characteristics:**
- Default command: `ansible-navigator run`
- Config struct: ~40 fields (modern layout)
- Play struct for multi-play execution
- Comprehensive `Config.Validate()` method
- GalaxyManager for roles/collections
- Executes commands locally via SSH proxy

## Architectural Decision

### Approach: Feature Parity with Shared Components

Rather than merging the two provisioners or creating a shared base, we will:

1. **Port the Config structure** - Add all modern fields to the local provisioner's Config struct
2. **Port the Play struct** - Copy the Play struct definition to enable multi-play execution
3. **Reuse GalaxyManager** - Import and use the GalaxyManager from the remote provisioner package, or duplicate it for isolation
4. **Adapt execution logic** - Modify the execution flow to work with communicator-based remote execution

### Key Design Decisions

#### Decision 1: Keep Provisioners Separate

**Rationale:** The two provisioners have fundamentally different execution models:
- Local: Uploads files to remote, executes ansible via communicator
- Remote: Sets up SSH proxy, executes ansible locally against the proxy

Merging would create complex conditional logic and increase maintenance burden.

#### Decision 2: Port Config Fields, Not Code Logic

**Rationale:** The execution flow differs significantly. We port the data structures (Config, Play) but adapt the execution logic to the local provisioner's communicator-based model.

#### Decision 3: GalaxyManager Duplication vs Import

**Options:**
1. Import GalaxyManager from remote package
2. Duplicate GalaxyManager into local package
3. Extract to shared package

**Decision:** Duplicate GalaxyManager into the local package initially. This:
- Avoids circular dependencies
- Allows adaptation for local execution context
- Can be refactored to shared package later if beneficial

#### Decision 4: Backward Compatibility

**Legacy fields to preserve:**
- `playbook_file` / `playbook_files` - Continue supporting
- `staging_directory` - Keep for local provisioner context
- `galaxy_file` - Keep as alias for `requirements_file`

**Deprecation strategy:**
- Add warnings for legacy-only patterns
- Document migration path in README

## Component Changes

### 1. Config Struct Updates

```go
// ADDED fields to local provisioner Config
type Config struct {
    // ... existing fields ...
    
    // Modern Ansible Navigator fields (ADDED)
    NavigatorMode        string   `mapstructure:"navigator_mode"`
    ExecutionEnvironment string   `mapstructure:"execution_environment"`
    WorkDir              string   `mapstructure:"work_dir"`
    KeepGoing            bool     `mapstructure:"keep_going"`
    StructuredLogging    bool     `mapstructure:"structured_logging"`
    LogOutputPath        string   `mapstructure:"log_output_path"`
    
    // Play-based execution (ADDED)
    Plays            []Play   `mapstructure:"plays"`
    RequirementsFile string   `mapstructure:"requirements_file"`
    
    // Collections support (ADDED)
    Collections             []string `mapstructure:"collections"`
    CollectionsCacheDir     string   `mapstructure:"collections_cache_dir"`
    CollectionsOffline      bool     `mapstructure:"collections_offline"`
    CollectionsForceUpdate  bool     `mapstructure:"collections_force_update"`
    
    // Group management (ADDED)
    Groups []string `mapstructure:"groups"`
}
```

### 2. Play Struct

```go
// ADDED - Same structure as remote provisioner
type Play struct {
    Name      string            `mapstructure:"name"`
    Target    string            `mapstructure:"target"`
    ExtraVars map[string]string `mapstructure:"extra_vars"`
    Tags      []string          `mapstructure:"tags"`
    VarsFiles []string          `mapstructure:"vars_files"`
    Become    bool              `mapstructure:"become"`
}
```

### 3. Validation

```go
// ADDED - Config.Validate() method
func (c *Config) Validate() error {
    var errs *packersdk.MultiError
    
    // Mutual exclusivity check
    if c.PlaybookFile != "" && len(c.Plays) > 0 {
        errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
            "you may specify only one of `playbook_file` or `plays`"))
    }
    
    // ... additional validation ...
}
```

### 4. Command Default Change

```go
// MODIFIED - in Prepare()
if p.config.Command == "" {
    // OLD: p.config.Command = "ANSIBLE_FORCE_COLOR=1 PYTHONUNBUFFERED=1 ansible-playbook"
    p.config.Command = "ansible-navigator run"  // NEW
}
```

### 5. Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     Provision() Method                       │
├─────────────────────────────────────────────────────────────┤
│  1. Upload files to remote (playbook_dir, playbooks, etc.)  │
│  2. Install requirements via GalaxyManager (on remote)       │
│  3. Generate inventory file                                  │
│  4. Execute plays/playbooks via communicator                 │
│  5. Cleanup staging directory if configured                  │
└─────────────────────────────────────────────────────────────┘
```

## File Changes Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `provisioner/ansible-navigator/provisioner.go` | MODIFIED | Add Config fields, Play struct, Validate(), update default command |
| `provisioner/ansible-navigator/provisioner.hcl2spec.go` | REGENERATED | Auto-generated by `packer-sdc` |
| `provisioner/ansible-navigator/galaxy.go` | ADDED | Port GalaxyManager from remote |
| `provisioner/ansible-navigator/json_logging.go` | ADDED | Port structured logging support |
| `provisioner/ansible-navigator/provisioner_test.go` | MODIFIED | Add tests for new functionality |

## Migration Path

### For Users

1. Existing `playbook_file` configurations continue to work
2. Users can gradually adopt `plays` array syntax
3. `ansible-navigator` must now be installed instead of just `ansible-playbook`

### Fallback Strategy

For users who need legacy `ansible-playbook` behavior:
```hcl
provisioner "ansible-navigator" {
  command = "ansible-playbook"  # Override default
  playbook_file = "site.yml"
}
```

## Testing Strategy

1. **Unit tests** - Validate Config parsing, Play struct handling, validation logic
2. **Integration tests** - Verify ansible-navigator execution with various configurations
3. **Backward compatibility tests** - Ensure legacy playbook_file configurations work
4. **Plugin check** - `make plugin-check` validates SDK compatibility

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking existing configs | Keep legacy fields, provide override option |
| ansible-navigator not installed | Clear error message with installation guidance |
| Galaxy dependency differences | Test thoroughly with various requirement patterns |
| HCL2 spec drift | Run `make generate` as part of CI |