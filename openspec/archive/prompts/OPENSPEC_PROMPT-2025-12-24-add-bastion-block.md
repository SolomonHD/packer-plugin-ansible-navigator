# OpenSpec Prompt: Restructure Bastion Configuration as Nested Block

## Context

**Plugin**: packer-plugin-ansible-navigator  
**Files**:

- `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go`
- HCL2 spec generation via `go:generate` directives

Currently, bastion configuration uses flat top-level fields:

```hcl
provisioner "ansible-navigator" {
  connection_mode = "ssh_tunnel"
  bastion_host = "remote-access-aws.service.emory.edu"
  bastion_port = 22
  bastion_user = "shilli2"
  bastion_private_key_file = "~/.ssh/general_key.pem"
  bastion_password = ""  # Alternative to key file
}
```

This flat structure:

- Makes the configuration harder to scan visually
- Doesn't clearly group related settings
- Doesn't match the semantic structure (bastion is a related set of properties)
- Doesn't allow for an explicit enable/disable flag

## Goal

Restructure bastion configuration as a nested HCL block with sub-variables, making the configuration more readable and semantically aligned with the concept of "bastion configuration."

## Scope

### In-Scope

- Create a new `BastionConfig` struct to hold all bastion-related fields
- Add a `Bastion *BastionConfig` field to the Config struct
- Move existing flat bastion fields into the nested struct:
  - `BastionHost` → `Bastion.Host`
  - `BastionPort` → `Bastion.Port`
  - `BastionUser` → `Bastion.User`
  - `BastionPrivateKeyFile` → `Bastion.PrivateKeyFile`
  - `BastionPassword` → `Bastion.Password`
- Add optional `Enabled bool` field to BastionConfig for explicit opt-in
- Handle legacy flat fields with backward compatibility and deprecation warnings
- Update validation logic to check `Bastion` struct fields
- Update code that accesses bastion fields to use the new nested structure
- Update go:generate directive to include BastionConfig
- Update HOME path expansion for `Bastion.PrivateKeyFile`

### Out-of-Scope

- Removing legacy flat bastion_* fields completely (keep for backward compatibility)
- Changes to SSH tunnel implementation logic
- Additional bastion features (e.g., proxy jump chains)
- Changes to builder-level bastion configuration

## Desired Behavior

### New Configuration Syntax

```hcl
provisioner "ansible-navigator" {
  connection_mode = "ssh_tunnel"
  
  bastion {
    enabled              = true  # Optional, auto-enabled if host is set
    host                 = "remote-access-aws.service.emory.edu"
    user                 = "shilli2"
    private_key_file     = "~/.ssh/general_key.pem"
    port                 = 22  # Optional, defaults to 22
  }
}
```

### BastionConfig Struct Definition

```go
// BastionConfig represents SSH bastion (jump host) configuration for SSH tunnel mode.
// Used when connection_mode = "ssh_tunnel".
type BastionConfig struct {
    // Enable the bastion configuration. Auto-enabled if Host is set.
    Enabled bool `mapstructure:"enabled"`
    
    // Bastion host address (FQDN or IP)
    Host string `mapstructure:"host"`
    
    // SSH port on the bastion host. Defaults to 22.
    Port int `mapstructure:"port"`
    
    // SSH username for bastion authentication
    User string `mapstructure:"user"`
    
    // Path to SSH private key file for bastion authentication.
    // Either this or Password must be provided.
    // Supports HOME expansion (~ and ~/path).
    PrivateKeyFile string `mapstructure:"private_key_file"`
    
    // Password for bastion authentication.
    // Either this or PrivateKeyFile must be provided.
    Password string `mapstructure:"password"`
}
```

Add to Config struct:

```go
type Config struct {
    // ... existing fields ...
    
    // Bastion (jump host) configuration for SSH tunnel mode.
    // Used when connection_mode = "ssh_tunnel".
    Bastion *BastionConfig `mapstructure:"bastion"`
    
    // DEPRECATED: Use bastion.host instead
    BastionHost string `mapstructure:"bastion_host"`
    
    // DEPRECATED: Use bastion.port instead
    BastionPort int `mapstructure:"bastion_port"`
    
    // DEPRECATED: Use bastion.user instead
    BastionUser string `mapstructure:"bastion_user"`
    
    // DEPRECATED: Use bastion.private_key_file instead
    BastionPrivateKeyFile string `mapstructure:"bastion_private_key_file"`
    
    // DEPRECATED: Use bastion.password instead
    BastionPassword string `mapstructure:"bastion_password"`
}
```

### Backward Compatibility Migration

In `Prepare()`:

```go
// Migrate legacy flat bastion fields to nested bastion block
if p.config.BastionHost != "" || p.config.BastionUser != "" || p.config.BastionPrivateKeyFile != "" || p.config.BastionPassword != "" || p.config.BastionPort != 0 {
    ui.Message("Warning: bastion_host, bastion_port, bastion_user, bastion_private_key_file, and bastion_password are deprecated. Use the bastion {} block instead.")
    
    // Create bastion config if not already defined
    if p.config.Bastion == nil {
        p.config.Bastion = &BastionConfig{}
    }
    
    // Migrate fields (new block values take precedence)
    if p.config.Bastion.Host == "" && p.config.BastionHost != "" {
        p.config.Bastion.Host = p.config.BastionHost
    }
    if p.config.Bastion.Port == 0 && p.config.BastionPort != 0 {
        p.config.Bastion.Port = p.config.BastionPort
    }
    if p.config.Bastion.User == "" && p.config.BastionUser != "" {
        p.config.Bastion.User = p.config.BastionUser
    }
    if p.config.Bastion.PrivateKeyFile == "" && p.config.BastionPrivateKeyFile != "" {
        p.config.Bastion.PrivateKeyFile = p.config.BastionPrivateKeyFile
    }
    if p.config.Bastion.Password == "" && p.config.BastionPassword != "" {
        p.config.Bastion.Password = p.config.BastionPassword
    }
}

// Set bastion defaults
if p.config.Bastion != nil {
    if p.config.Bastion.Port == 0 {
        p.config.Bastion.Port = 22
    }
    // Auto-enable if host is set
    if p.config.Bastion.Host != "" {
        p.config.Bastion.Enabled = true
    }
}

// Apply HOME expansion to bastion private key file
if p.config.Bastion != nil && p.config.Bastion.PrivateKeyFile != "" {
    p.config.Bastion.PrivateKeyFile = expandUserPath(p.config.Bastion.PrivateKeyFile)
}
```

### Updated Validation

In `Validate()`:

```go
// Validate bastion configuration when using SSH tunnel mode
if c.ConnectionMode == "ssh_tunnel" {
    if c.Bastion == nil || c.Bastion.Host == "" {
        return fmt.Errorf("bastion.host is required when connection_mode='ssh_tunnel'")
    }
    
    if c.Bastion.User == "" {
        return fmt.Errorf("bastion.user is required when connection_mode='ssh_tunnel'")
    }
    
    // Require either key file or password
    if c.Bastion.PrivateKeyFile == "" && c.Bastion.Password == "" {
        return fmt.Errorf("either bastion.private_key_file or bastion.password must be provided when connection_mode='ssh_tunnel'")
    }
    
    // Validate port range
    if c.Bastion.Port < 1 || c.Bastion.Port > 65535 {
        return fmt.Errorf("bastion.port must be between 1 and 65535, got %d", c.Bastion.Port)
    }
    
    // Validate private key file exists if specified
    if c.Bastion.PrivateKeyFile != "" {
        if err := validateFileConfig(c.Bastion.PrivateKeyFile, "bastion.private_key_file", true); err != nil {
            return err
        }
    }
}
```

### Updated Code References

Replace all references like:

- `p.config.BastionHost` → `p.config.Bastion.Host`
- `p.config.BastionPort` → `p.config.Bastion.Port`
- `p.config.BastionUser` → `p.config.Bastion.User`
- `p.config.BastionPrivateKeyFile` → `p.config.Bastion.PrivateKeyFile`
- `p.config.BastionPassword` → `p.config.Bastion.Password`

Key locations:

- `Provision()` function - SSH tunnel setup (~line 1381-1427)
- `setupSSHTunnel()` function (~line 1122-1259)

Example in `setupSSHTunnel()`:

```go
// Before
bastionAddr := fmt.Sprintf("%s:%d", p.config.BastionHost, p.config.BastionPort)

// After
bastionAddr := fmt.Sprintf("%s:%d", p.config.Bastion.Host, p.config.Bastion.Port)
```

## Constraints & Assumptions

- Must maintain 100% backward compatibility with flat bastion_* fields
- Nested block values take precedence over legacy flat values if both are specified
- The `Enabled` field can be omitted - presence of `Host` implies enabled
- Default port remains 22
- HOME expansion (~/) must work for `private_key_file`
- Users should be able to configure bastion inline in a single block (no external files)

## Acceptance Criteria

### Must Have

- [ ] Create `BastionConfig` struct with fields: `Enabled`, `Host`, `Port`, `User`, `PrivateKeyFile`, `Password`
- [ ] Add `Bastion *BastionConfig` field to Config struct with `mapstructure:"bastion"` tag
- [ ] Mark legacy flat `Bastion*` fields as deprecated in comments
- [ ] Add migration logic in `Prepare()` that:
  - Detects legacy flat bastion fields
  - Migrates them to `Bastion` struct
  - Shows deprecation warning
  - Gives precedence to new block syntax if both are present
- [ ] Set default `Bastion.Port = 22` in `Prepare()` if not specified
- [ ] Auto-enable bastion (`Bastion.Enabled = true`) if `Bastion.Host` is set
- [ ] Apply HOME expansion to `Bastion.PrivateKeyFile` in `Prepare()`
- [ ] Update validation logic to check `!c.Bastion == nil && c.Bastion.Host` instead of `c.BastionHost`
- [ ] Replace all code references to `p.config.BastionHost` with `p.config.Bastion.Host` (and similar for other fields)
- [ ] Update `go:generate` directive to include `BastionConfig`
- [ ] Run `make generate` to regenerate HCL2 specs
- [ ] Code compiles with `go build ./...`

### Should Have

- [ ] Add struct-level documentation comment for `BastionConfig` explaining its purpose
- [ ] Add field-level documentation for each `BastionConfig` field
- [ ] Validation errors reference the new field names (`bastion.host` not `bastion_host`)

### Nice to Have

- [ ] Add example in code comments showing the nested block syntax
- [ ] Consider adding a helper method like `HasBastion() bool`

### Test Cases

- [ ] Config with nested `bastion {}` block → validates and uses bastion config
- [ ] Config with nested bastion + `connection_mode = "ssh_tunnel"` → works correctly
- [ ] Config with nested bastion but missing `host` → fails validation with clear error
- [ ] Config with nested bastion but missing credentials → fails validation (requires private_key_file OR password)
- [ ] Legacy config with flat `bastion_host`, `bastion_user`, etc. → migrates to nested block with warning
- [ ] Config with BOTH nested bastion and flat fields → nested values take precedence, warning shown
- [ ] Bastion port defaults to 22 when omitted
- [ ] Bastion `private_key_file` with `~/` expands to HOME directory
- [ ] Bastion validation fails for port < 1 or > 65535

## Expected Files Touched

- `provisioner/ansible-navigator/provisioner.go`
  - Add BastionConfig struct definition (~line 290+)
  - Config struct - add Bastion field, mark bastion_* fields deprecated (~line 320-557)
  - Validate() - update bastion validation (~line 570-607)
  - Prepare() - add migration logic, defaults, PATH expansion (~line 739-795)
  - Provision() - update bastion field references (~line 1408)
  - setupSSHTunnel() - update bastion field references (~line 1129-1158)
- Generated files after `make generate`:
  - `provisioner/ansible-navigator/provisioner.hcl2spec.go`

## Dependencies

- **Depends on**: 02-add-connection-mode-enum.md (validation checks `connection_mode == "ssh_tunnel"`)
- Soft dependency on 01-fix-port-type-coercion.md (independent changes, but both should be in the same release)
