# Change: Rename `plays` to `play` and Fix HCL2 Block Syntax

## Why

Two issues exist with the current `plays` configuration:

### Issue 1: Incorrect Documentation Syntax
The documentation (README.md and docs/*.md files) shows incorrect HCL2 syntax for the `plays` configuration option. The examples use array assignment syntax (`plays = [...]`) when the actual Packer SDK HCL2 implementation uses `BlockListSpec`, which requires HCL2 block syntax (`plays { ... }`).

This causes users following the documentation to encounter Packer errors:
```
Unsupported argument; An argument named "plays" is not expected here.
Did you mean to define a block of type "plays"?
```

### Issue 2: Non-Idiomatic Block Name
The block name `plays` (plural) violates HCL conventions. In Packer, Terraform, and other HCL-based tools, repeatable blocks use **singular names**:
- `provisioner` (not `provisioners`)
- `source` (not `sources`)
- `build` (not `builds`)
- `resource` (not `resources`)
- `variable` (not `variables`)

Using `plays` for a block that defines a single play is inconsistent and confusing.

**Root Cause:** The `plays` field in [`provisioner.hcl2spec.go:105`](provisioner/ansible-navigator/provisioner.hcl2spec.go:105) is defined as:
```go
"plays": &hcldec.BlockListSpec{TypeName: "plays", Nested: hcldec.ObjectSpec((*FlatPlay)(nil).HCL2Spec())},
```

## What Changes

### Code Changes
- **provisioner/ansible-navigator/provisioner.go**: Rename `Plays` field HCL tag from `plays` to `play`
- **provisioner/ansible-navigator/provisioner.hcl2spec.go**: Regenerate with `make generate` to update `BlockListSpec` TypeName from `"plays"` to `"play"`
- **Update all code references** that process the Plays field (validation, execution logic)

### Documentation Changes
- **README.md**: Update all examples to use `play { }` block syntax
- **docs/UNIFIED_PLAYS.md**: Update all examples and descriptions
- **docs/EXAMPLES.md**: Update all examples
- **docs/CONFIGURATION.md**: Update documentation and examples
- **AGENTS.md**: Update configuration examples
- **CHANGELOG.md**: Document the breaking change

## Impact

- **Breaking Change**: Yes - existing configurations using `plays { }` will need to change to `play { }`
- **Affected specs**: local-provisioner-capabilities (updated naming)
- **Affected code**:
  - provisioner/ansible-navigator/provisioner.go
  - provisioner/ansible-navigator/provisioner.hcl2spec.go
  - provisioner/ansible-navigator/provisioner_test.go
- **Affected files**:
  - README.md
  - AGENTS.md
  - CHANGELOG.md
  - docs/UNIFIED_PLAYS.md
  - docs/EXAMPLES.md
  - docs/CONFIGURATION.md
  - docs/provisioners/*.mdx

## Syntax Transformation

### INCORRECT (current documentation - array syntax)

```hcl
plays = [
  {
    name = "Play Name"
    target = "namespace.collection.role"
    become = true
  }
]
```

### DEPRECATED (current code - plural block name)

```hcl
plays {
  name   = "Play Name"
  target = "namespace.collection.role"
  become = true
}
```

### CORRECT (new syntax - singular block name)

```hcl
play {
  name   = "Play Name"
  target = "namespace.collection.role"
  become = true
}

play {
  name   = "Another Play"
  target = "another.collection.role"
}
```

Multiple plays are defined as repeated `play { }` blocks, following HCL idioms.

## Rationale for `play` over `run`

Alternative considered: `run` (to match `ansible-navigator run` CLI)

**Decision: Use `play`** because:
1. **HCL Convention Compliance** - Follows standard singular block naming
2. **Ansible Domain Alignment** - "Play" is fundamental Ansible terminology
3. **Self-Documenting** - Users instantly understand "this defines an Ansible play"
4. **Provisioner Name Provides Context** - The `ansible-navigator` provisioner name already indicates CLI alignment