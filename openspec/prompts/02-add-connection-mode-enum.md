# OpenSpec Prompt: Add connection_mode Enum to Replace use_proxy and ssh_tunnel_mode

## Context

**Plugin**: packer-plugin-ansible-navigator  
**Files**:

- `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go`
- HCL2 spec generation via `go:generate` directives

The current connection method configuration is confusing:

- Setting `use_proxy = false` + `ssh_tunnel_mode = true` creates a double-negative
- Users must explicitly disable one feature to enable another
- The relationship between these two fields is not immediately clear
- Validation logic checks mutual exclusivity, but the config structure doesn't express this

Example of current confusing configuration:

```hcl
provisioner "ansible-navigator" {
  use_proxy = false  # Must disable proxy...
  ssh_tunnel_mode = true  # ...to enable tunnel mode
  # What does use_proxy=true + ssh_tunnel_mode=false mean?
  # What does use_proxy=false + ssh_tunnel_mode=false mean?
}
```

## Goal

Replace the confusing `use_proxy` (Trilean) and `ssh_tunnel_mode` (bool) fields with a single,  explicit `connection_mode` (string enum) field that clearly expresses the user's intent.

## Scope

### In-Scope

- Add new `connection_mode` string field to Config struct
- Define and validate allowed connection_mode values: `"proxy"`, `"ssh_tunnel"`, `"direct"`
- Handle legacy `use_proxy` and `ssh_tunnel_mode` fields with:
  - Automatic migration to `connection_mode` in Prepare()
  - Deprecation warnings
  - Maintaining backward compatibility
- Update all code that checks `UseProxy` or `SSHTunnelMode` to check `ConnectionMode` instead
- Add validation for connection_mode enum values
- Update go:generate directive to regenerate HCL2 specs
- Document the new field and deprecations

### Out-of-Scope

- Removing the legacy fields (keep for backward compatibility)
- Restructuring bastion configuration (separate prompt)
- Changes to proxy adapter or SSH tunnel implementation logic
- Documentation website updates (code-level comments only)

## Desired Behavior

### Configuration Interface

Users should configure connection method with a single, explicit choice:

```hcl
provisioner "ansible-navigator" {
  # Clear, explicit connection mode (NEW)
  connection_mode = "ssh_tunnel"  # or "proxy" or "direct"
  
  # Bastion settings used when connection_mode = "ssh_tunnel"
  bastion_host = "remote-access-aws.service.emory.edu"
  bastion_user = "shilli2"
  bastion_private_key_file = "~/.ssh/general_key.pem"
}
```

### Valid connection_mode Values

| Value | Behavior |  Use Case |
|-------|----------|-----------|
| `"proxy"` | Use Packer's SSH proxy adapter (default) | Standard Packer builds, Docker builds |
| `"ssh_tunnel"` | Establish SSH tunnel through bastion | WSL2 + containers, bastion-only access |
| `"direct"` | Connect directly without proxy | When host IP is accessible and no proxy needed |

### Backward Compatibility

Old configurations should continue to work with a deprecation warning:

```hcl
# OLD (still works, but warns)
provisioner "ansible-navigator" {
  use_proxy = false
  ssh_tunnel_mode = true
}
```

Migration logic in Prepare():

```go
if p.config.SSHTunnelMode {
    ui.Message("Warning: ssh_tunnel_mode is deprecated. Use connection_mode='ssh_tunnel' instead.")
    p.config.ConnectionMode = "ssh_tunnel"
} else if p.config.UseProxy.False() {
    ui.Message("Warning: use_proxy is deprecated. Use connection_mode='direct' instead.")
    p.config.ConnectionMode = "direct"
} else if p.config.UseProxy.True() {
    p.config.ConnectionMode = "proxy"
}

// Default to proxy mode if not specified
if p.config.ConnectionMode == "" {
    p.config.ConnectionMode = "proxy"
}
```

### Validation Logic

Replace existing mutual exclusivity check with enum validation:

```go
// Validate connection mode
validModes := []string{"proxy", "ssh_tunnel", "direct"}
if !slices.Contains(validModes, c.ConnectionMode) {
    return fmt.Errorf("connection_mode must be one of %v, got: %q", validModes, c.ConnectionMode)
}

// Validate bastion requirements
if c.ConnectionMode == "ssh_tunnel" {
    if c.BastionHost == "" {
        return fmt.Errorf("bastion_host is required when connection_mode='ssh_tunnel'")
    }
    // ... other bastion field validation
}
```

### Code Changes Throughout Plugin

Replace all checks of `UseProxy` and `SSHTunnelMode`:

**Before:**

```go
if p.config.SSHTunnelMode {
    // SSH tunnel code
} else if !p.config.UseProxy.False() {
    // Proxy code
} else {
    // Direct code
}
```

**After:**

```go
switch p.config.ConnectionMode {
case "ssh_tunnel":
    // SSH tunnel code
case "proxy":
    // Proxy code
case "direct":
    // Direct code
}
```

## Constraints & Assumptions

- Must maintain 100% backward compatibility with existing HCL configurations
- Deprecation warnings should guide users to the new field
- The default behavior (proxy mode) must remain unchanged
- HCL2 spec files are auto-generated, so changes to Config struct require `make generate`
- Connection mode determines which other fields are required (e.g., bastion_* for ssh_tunnel)
- The three modes cover all current and reasonable future connection scenarios

## Acceptance Criteria

### Must Have

- [ ] Add `ConnectionMode string` field to Config struct with `mapstructure:"connection_mode"` tag
- [ ] Mark `UseProxy` and `SSHTunnelMode` fields as deprecated in comments
- [ ] Add migration logic in `Prepare()` that converts legacy fields to `connection_mode`
- [ ] Display deprecation warnings when legacy fields are used
- [ ] Implement enum validation in `Validate()` that checks `connection_mode` is one of: `{"proxy", "ssh_tunnel", "direct"}`
- [ ] Default `ConnectionMode` to `"proxy"` when not explicitly set
- [ ] Replace all `if p.config.SSHTunnelMode` checks with `if p.config.ConnectionMode == "ssh_tunnel"`
- [ ] Replace all `if p.config.UseProxy.False()` checks with `if p.config.ConnectionMode == "direct"`
- [ ] Replace all `if !p.config. UseProxy.False()` checks with `if p.config.ConnectionMode == "proxy"`
- [ ] Update `go:generate` directive on line 6 to include `ConnectionMode` if not already covered
- [ ] Run `make generate` to regenerate HCL2 spec files
- [ ] Code compiles with `go build ./...`

### Should Have

- [ ] Add field-level documentation comments explaining each connection mode value
- [ ] Use a constant slice for valid modes instead of repeating the list
- [ ] Consider adding helper methods like `IsProxyMode() bool`, `IsSSHTunnelMode() bool` for readability
- [ ] Validation error messages include examples of valid values

### Nice to Have

- [ ] Add examples in code comments showing the new vs old configuration
- [ ] Document the migration path in a code comment block

### Test Cases

- [ ] Config with `connection_mode = "proxy"` works (proxy adapter used)
- [ ] Config with `connection_mode = "ssh_tunnel"` + bastion settings works (tunnel created)
- [ ] Config with `connection_mode = "direct"` works (no proxy/tunnel)
- [ ] Config with `connection_mode = "invalid"` fails validation with clear error
- [ ] Legacy config with `use_proxy = false` + `ssh_tunnel_mode = true` → migrates to `connection_mode = "ssh_tunnel"` with warning
- [ ] Legacy config with `use_proxy = false` + `ssh_tunnel_mode = false` → migrates to `connection_mode = "direct"` with warning
- [ ] Legacy config with `use_proxy = true` → migrates to `connection_mode = "proxy"` with warning
- [ ] Config with no connection fields specified → defaults to `connection_mode = "proxy"`
- [ ] Config with `connection_mode = "ssh_tunnel"` but missing `bastion_host` → fails validation

## Expected Files Touched

- `provisioner/ansible-navigator/provisioner.go`
  - Config struct definition (~line 320-557)
  - Validate() function (~line 560-716)
  - Prepare() function (~line 739-873)
  - Provision() function (~line 1351-1544) - connection mode logic (~line 1380-1522)
- Generated files after `make generate`:
  - `provisioner/ansible-navigator/provisioner.hcl2spec.go`

## Dependencies

- **Depends on**: 01-fix-port-type-coercion.md (should be completed first, but not strictly required)
- **Blocks**: 03-add-bastion-block.md (bastion block restructure will work with connection_mode)
