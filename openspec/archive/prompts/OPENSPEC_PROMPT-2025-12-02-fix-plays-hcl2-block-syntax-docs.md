# OpenSpec change prompt

## Context

The `packer-plugin-ansible-navigator` README.md contains incorrect HCL2 syntax examples for the `plays` configuration. The documentation shows `plays` as an array assignment (`plays = [...]`) or array of objects, but the actual Packer SDK HCL2 implementation uses `BlockListSpec`, which requires HCL2 block syntax (`plays { ... }`).

This causes every user following the README to encounter Packer errors like:
```
Unsupported argument; An argument named "plays" is not expected here. 
Did you mean to define a block of type "plays"?
```

## Goal

Update all `plays` syntax examples in README.md (and any related documentation in `docs/`) to use the correct HCL2 block syntax that matches the generated `provisioner.hcl2spec.go`.

## Scope

- In scope:
  - `README.md` - All `plays` examples need correction
  - `docs/UNIFIED_PLAYS.md` - If it exists, check and correct
  - `docs/EXAMPLES.md` - If it exists, check and correct
  - `docs/CONFIGURATION.md` - If it exists, check and correct
  - Any other documentation showing `plays` syntax

- Out of scope:
  - Changing the Go code or HCL2 spec generation
  - Adding new features
  - Changing how `collections` or other array arguments work (those use `AttrSpec` and work correctly with assignment syntax)

## Desired behavior

After the change:
- All README examples should work when copy-pasted into a Packer HCL file
- Users understand that `plays` uses block syntax, not array assignment
- Multiple plays are defined as repeated `plays { }` blocks

## Constraints & assumptions

- Assumption: We are correcting documentation to match existing implementation, not changing implementation
- Assumption: The `collections` field using array syntax (`collections = [...]`) is correct and should not be changed
- Constraint: Do not modify any `.go` files or regenerate HCL2 specs

## Acceptance criteria

- [ ] All `plays = [...]` syntax in README.md replaced with `plays { ... }` block syntax
- [ ] All `plays = [{ target = "..." }]` replaced with repeated `plays { target = "..." }` blocks
- [ ] Example 2 (lines ~57-68) corrected from string array to proper block(s)
- [ ] Example 3 (lines ~73-98) corrected from object array to repeated blocks
- [ ] Container Images example (~104-127) corrected
- [ ] Multi-Stage Deployment example (~147-177) corrected
- [ ] Dual Invocation Mode section (~201-208) corrected
- [ ] Quick Reference table (~237) updated if needed
- [ ] Any docs/ files checked and corrected if they contain incorrect `plays` syntax

## Correct syntax reference

```hcl
# WRONG (what README currently shows):
plays = [
  "role.name"
]

plays = [
  {
    name = "Play Name"
    target = "namespace.collection.role"
    become = true
  }
]

# CORRECT (what it should show):
plays {
  name   = "Play Name"
  target = "namespace.collection.role"
  become = true
}

plays {
  name   = "Another Play"
  target = "another.collection.role"
}