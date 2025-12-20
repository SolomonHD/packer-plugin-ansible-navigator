# Proposal: Standardize Command-Line Argument Syntax

## Summary

Convert all `ansible-navigator` and `ansible-galaxy` command-line argument construction from `--flag value` (two separate array elements) to `--flag=value` (single element) syntax to prevent argument parsing failures when values start with `-` or contain spaces, especially when ansible-navigator passes arguments to ansible-playbook inside execution environments.

## Why

The plugin currently constructs command-line arguments using separate array elements for flags and their values:

```go
args = append(args, "--extra-vars", fmt.Sprintf("@%s", extraVarsFilePath))
args = append(args, "--ssh-extra-args", "-o IdentitiesOnly=yes")
args = append(args, "--limit", p.config.Limit)
args = append(args, "--tags", tag)
```

While Go's `exec.Command` bypasses shell interpretation, ansible-navigator internally invokes `ansible-playbook` which may re-parse arguments. When values start with `-` or contain spaces, they can be misinterpreted as new flags, causing execution failures.

## What Changes

Use `--flag=value` syntax for all flag+value pairs:

```go
args = append(args, fmt.Sprintf("--extra-vars=@%s", extraVarsFilePath))
args = append(args, fmt.Sprintf("--ssh-extra-args=-o IdentitiesOnly=yes"))
args = append(args, fmt.Sprintf("--limit=%s", p.config.Limit))
args = append(args, fmt.Sprintf("--tags=%s", tag))
```

## Scope

### In Scope

- All `args = append(args, "--flag", value)` patterns in:
  - [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go)
  - [`provisioner/ansible-navigator/galaxy.go`](../../../provisioner/ansible-navigator/galaxy.go)
  - [`provisioner/ansible-navigator-local/provisioner.go`](../../../provisioner/ansible-navigator-local/provisioner.go)
  - [`provisioner/ansible-navigator-local/galaxy.go`](../../../provisioner/ansible-navigator-local/galaxy.go)
- Short flag patterns (`-e`, `-i`, `-p`) are also converted

### Out of Scope

- Positional arguments (playbook paths, inventory files)
- Boolean flags without values (`--become`, `--force`, `--offline`)
- Environment variable handling
- Config file generation logic

## Technical Details

### Affected Patterns

1. **Long flags with values**:
   ```go
   // Before: args = append(args, "--extra-vars", fmt.Sprintf("@%s", path))
   // After:  args = append(args, fmt.Sprintf("--extra-vars=@%s", path))
   ```

2. **Short flags with values**:
   ```go
   // Before: args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
   // After:  args = append(args, fmt.Sprintf("-e=%s=%s", k, v))
   ```

3. **Flags with literal values**:
   ```go
   // Before: args = append(args, "--mode", "stdout")
   // After:  args = append(args, "--mode=stdout")
   ```

4. **Boolean flags (unchanged)**:
   ```go
   // Unchanged: args = append(args, "--become")
   // Unchanged: args = append(args, "--force")
   ```

### Assumptions

- ansible-navigator and ansible-galaxy accept `--flag=value` syntax for all flags currently using `--flag value`
- Short flags (`-e`, `-i`, `-p`) also accept `-f=value` syntax
- This change maintains backward compatibility with existing HCL configurations (no user-facing config changes)

## Benefits

1. **Robustness**: Prevents parsing failures when values contain special characters
2. **Consistency**: Single standard for argument construction
3. **Debugging**: Easier to identify complete flag+value pairs in logs
4. **Future-proofing**: Reduces risk of edge-case failures as configurations evolve

## Risks

- **Low risk**: If ansible-navigator/ansible-galaxy don't support `=` syntax for any specific flag, tests will catch it
- **Mitigation**: Comprehensive testing with existing test fixtures

## Validation Approach

1. Run `make generate` (regenerate HCL2 specs)
2. Run `go build ./...` (compile verification)
3. Run `go test ./...` (existing test suite)
4. Run `make plugin-check` (SDK compatibility)
5. Manual integration test with example HCL configurations

## Acceptance Criteria

- [ ] All `--flag value` patterns converted to `--flag=value` in provisioner.go files
- [ ] All `-f value` short flag patterns converted to `-f=value` in galaxy.go files
- [ ] Boolean flags remain unchanged
- [ ] Positional arguments remain as separate array elements
- [ ] Plugin builds successfully
- [ ] All unit tests pass
- [ ] Plugin check passes
- [ ] Manual integration test with builds succeeds without "argument expected" errors
