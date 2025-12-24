# Proposal: Replace connection_mode with Enum (Breaking Change)

## Why

The current connection configuration uses two separate boolean/trilean fields ([`use_proxy`](../../../provisioner/ansible-navigator/provisioner.go:495) and [`ssh_tunnel_mode`](../../../provisioner/ansible-navigator/provisioner.go:535)) to control three distinct connection modes. This design is confusing and error-prone:

- Setting `use_proxy = false` + `ssh_tunnel_mode = true` creates a double-negative pattern
- The mutual exclusivity is enforced via validation but not expressed in the schema
- Users must disable one feature to enable another
- The relationship between these fields is not immediately clear from the configuration

This proposal removes both legacy fields entirely and replaces them with a single, explicit `connection_mode` string enum field.

**Project Context**: Per [`openspec/project.md`](../../project.md), backward compatibility is not a goal for this plugin. Breaking changes and removal of legacy configuration options are acceptable.

## What Changes

Replace the confusing `use_proxy` (Trilean) and `ssh_tunnel_mode` (bool) fields with a single, explicit `connection_mode` (string enum) field that clearly expresses connection intent without backwards compatibility.

### Summary of Changes

### 1. Remove Legacy Fields

**File**: [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go)

Remove these fields from the Config struct (around lines 495 and 535):

- `UseProxy config.Trilean` - Remove entirely
- `SSHTunnelMode bool` - Remove entirely

### 2. Add New `connection_mode` Field

Add new field to Config struct:

```go
// Connection mode determines how Ansible connects to the target machine.
// Valid values: "proxy", "ssh_tunnel", "direct"
// Default: "proxy"
//
// - "proxy": Use Packer's SSH proxy adapter (default, works for most builds)
// - "ssh_tunnel": Establish SSH tunnel through bastion host (for bastion-only access)
// - "direct": Connect directly without proxy (when target IP is directly accessible)
ConnectionMode string `mapstructure:"connection_mode"`
```

### 3. Update Validation Logic

**File**: [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go)

Replace the SSH tunnel validation section (currently around lines 571-607) with:

```go
// Validate connection_mode
validModes := []string{"proxy", "ssh_tunnel", "direct"}
if p.config.ConnectionMode == "" {
    p.config.ConnectionMode = "proxy" // Apply default
}
if !slices.Contains(validModes, p.config.ConnectionMode) {
    errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
        "connection_mode must be one of %v, got: %q", validModes, p.config.ConnectionMode))
}

// Validate bastion requirements when using ssh_tunnel mode
if p.config.ConnectionMode == "ssh_tunnel" {
    if p.config.BastionHost == "" {
        errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
            "bastion_host is required when connection_mode='ssh_tunnel'"))
    }
    if p.config.BastionUser == "" {
        errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
            "bastion_user is required when connection_mode='ssh_tunnel'"))
    }
    if p.config.BastionPrivateKeyFile == "" && p.config.BastionPassword == "" {
        errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
            "either bastion_private_key_file or bastion_password must be provided when connection_mode='ssh_tunnel'"))
    }
    if p.config.BastionPort < 1 || p.config.BastionPort > 65535 {
        errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
            "bastion_port must be between 1 and 65535, got %d", p.config.BastionPort))
    }
    if p.config.BastionPrivateKeyFile != "" {
        if err := validateFileConfig(p.config.BastionPrivateKeyFile, "bastion_private_key_file", true); err != nil {
            errs = packersdk.MultiErrorAppend(errs, err)
        }
    }
}
```

### 4. Update Provision() Connection Logic

**File**: [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go)

Replace the current connection branching logic (around lines 1361-1537) with a cleaner switch statement:

**Before** (fragmented logic with multiple if/else branches):

```go
if p.config.UseProxy.False() {
    // Check for fallback...
}
if p.config.SSHTunnelMode {
    // SSH tunnel mode...
} else if !p.config.UseProxy.False() {
    // Proxy adapter...
} else {
    // Direct mode...
}
```

**After** (clear switch on connection_mode):

```go
switch p.config.ConnectionMode {
case "ssh_tunnel":
    // SSH tunnel mode - establish tunnel through bastion
    ui.Message("Using SSH tunnel mode - connecting through bastion host")
    
    targetHost, ok := generatedData["Host"].(string)
    if !ok || targetHost == "" {
        return fmt.Errorf("SSH tunnel mode requires a valid target host")
    }
    
    // Extract and validate target port...
    // (existing port extraction code from lines 1389-1406)
    
    // Establish tunnel and get privKeyFile...
    // (existing tunnel setup code from lines 1408-1475)
    
case "direct":
    // Direct connection mode - use SSH keys from communicator
    ui.Message("Using direct connection mode - connecting without proxy")
    
    connType := generatedData["ConnType"].(string)
    switch connType {
    case "ssh":
        // (existing direct SSH logic from lines 1502-1532)
    case "winrm":
        ui.Message("Using WinRM password from Packer communicator")
    }
    
case "proxy":
    // Proxy adapter mode (default)
    ui.Message("Using proxy adapter mode")
    
    // (existing proxy setup logic from lines 1476-1497)
    
default:
    return fmt.Errorf("invalid connection_mode: %q (should have been caught by validation)", p.config.ConnectionMode)
}
```

### 5. Update createInventoryFile() Logic

**File**: [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go)

Replace `UseProxy` checks (around lines 1304-1316):

**Before**:

```go
if p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
    hostTemplate = DefaultWinRMInventoryFilev2
}
// ...
if !p.config.UseProxy.False() {
    ctxData["Host"] = p.config.AnsibleProxyHost
    ctxData["Port"] = p.config.LocalPort
}
```

**After**:

```go
if p.config.ConnectionMode == "direct" && p.generatedData["ConnType"] == "winrm" {
    hostTemplate = DefaultWinRMInventoryFilev2
}
// ...
if p.config.ConnectionMode == "proxy" || p.config.ConnectionMode == "ssh_tunnel" {
    ctxData["Host"] = p.config.AnsibleProxyHost
    ctxData["Port"] = p.config.LocalPort
}
```

### 6. Update createCmdArgs() Logic

**File**: [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go)

Replace `UseProxy` checks (around lines 1647-1652):

**Before**:

```go
if p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
    if password, ok := p.generatedData["Password"]; ok {
        extraVars["ansible_password"] = fmt.Sprint(password)
        ansiblePasswordSet = true
    }
}
```

**After**:

```go
if p.config.ConnectionMode == "direct" && p.generatedData["ConnType"] == "winrm" {
    if password, ok := p.generatedData["Password"]; ok {
        extraVars["ansible_password"] = fmt.Sprint(password)
        ansiblePasswordSet = true
    }
}
```

### 7. Regenerate HCL2 Specs

Update the `go:generate` directive (line 6) to include `ConnectionMode` if not already covered, then run:

```bash
make generate
```

This regenerates [`provisioner/ansible-navigator/provisioner.hcl2spec.go`](../../../provisioner/ansible-navigator/provisioner.hcl2spec.go) with the new field and without the removed fields.

### 8. Update Documentation Comments

Add comprehensive field-level documentation:

```go
// ConnectionMode determines how Ansible connects to the target machine.
//
// Valid values:
//   - "proxy" (default): Use Packer's SSH proxy adapter. Works for most builds including Docker.
//   - "ssh_tunnel": Establish SSH tunnel through a bastion host. Required when targets are only
//     accessible via jump host (common with WSL2 execution environments).
//   - "direct": Connect directly to the target without proxy. Use when the target IP is directly
//     accessible and proxy overhead is unnecessary.
//
// When using "ssh_tunnel", you must provide bastion_host, bastion_user, and either
// bastion_private_key_file or bastion_password.
ConnectionMode string `mapstructure:"connection_mode"`
```

## Impact

### Breaking Changes

**Existing configurations using `use_proxy` or `ssh_tunnel_mode` will break**. Users must update:

| Old Configuration | New Configuration |
|------------------|-------------------|
| `use_proxy = true` (or unset) | `connection_mode = "proxy"` |
| `use_proxy = false` + `ssh_tunnel_mode = false` | `connection_mode = "direct"` |
| `use_proxy = false` + `ssh_tunnel_mode = true` | `connection_mode = "ssh_tunnel"` |

### Migration Guide

Users should update their HCL configurations:

**Before:**

```hcl
provisioner "ansible-navigator" {
  use_proxy = false
  ssh_tunnel_mode = true
  bastion_host = "bastion.example.com"
  bastion_user = "user"
  bastion_private_key_file = "~/.ssh/id_rsa"
}
```

**After:**

```hcl
provisioner "ansible-navigator" {
  connection_mode = "ssh_tunnel"
  bastion_host = "bastion.example.com"
  bastion_user = "user"
  bastion_private_key_file = "~/.ssh/id_rsa"
}
```

## Dependencies

- Should be completed after [`fix-port-type-coercion`](../fix-port-type-coercion/proposal.md) if that change hasn't been archived yet
- Blocks prompt 03 (bastion block restructure) which will work better with the clearer `connection_mode` field
