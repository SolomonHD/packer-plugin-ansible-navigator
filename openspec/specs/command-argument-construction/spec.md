# command-argument-construction Specification

## Purpose
TBD - created by archiving change standardize-cmd-arg-syntax. Update Purpose after archive.
## Requirements
### Requirement: Flag-Value Argument Syntax

All command-line arguments for `ansible-navigator` and `ansible-galaxy` commands MUST use `--flag=value` syntax (single array element) instead of `--flag value` (two separate elements) to prevent argument parsing failures.

#### Scenario: Long flag with dynamic value
**Given:** A configuration requires passing `--extra-vars` with a file path  
**When:** Constructing arguments for `ansible-navigator run`  
**Then:** The argument must be formatted as a single element: `--extra-vars=@/path/to/file.yml`  
**And:** The value /path/to/file.yml may contain special characters or start with `-`  
**And:** The argument list contains one element, not two

#### Scenario: Short flag with dynamic value
**Given:** A configuration requires passing `-e` with a key=value pair  
**When:** Constructing arguments for `ansible-galaxy` commands  
**Then:** The argument must be formatted as: `-e=key=value`  
**And:** The combined format preserves the nested `=` for Ansible's key-value syntax

#### Scenario: Multiple flag-value pairs
**Given:** A play configuration specifies multiple tags  
**When:** Constructing arguments for each tag  
**Then:** Each tag argument must be a single element: `--tags=value1`, `--tags=value2`  
**And:** Multiple occurrences of the same flag are preserved

#### Scenario: Flag with literal string value
**Given:** NavigatorConfig specifies mode as "stdout"  
**When:** Adding the mode argument  
**Then:** The argument must be formatted as: `--mode=stdout`

#### Scenario: Boolean flags without values
**Given:** A play configuration enables become  
**When:** Constructing the argument list  
**Then:** Boolean flags must remain as single elements without values: `--become`  
**And:** No `=` is appended for boolean flags

#### Scenario: Positional arguments
**Given:** An ansible-navigator command requires a playbook path  
**When:** Adding the playbook path to arguments  
**Then:** Positional arguments must remain as separate array elements  
**And:** The playbook path follows all flag arguments

### Requirement: Flag Pattern Identification

The plugin MUST correctly identify and format different argument patterns.

#### Scenario: Distinguishing flags from positional arguments
**Given:** A command requires both flags and positional arguments  
**When:** Constructing the complete argument array  
**Then:** All flags (starting with `-` or `--`) with values use `=` syntax  
**And:** Positional arguments remain separate elements after all flags  
**And:** The order is: [flags with values, boolean flags, positional arguments]

#### Scenario: Handling values that start with hyphen
**Given:** A configuration value starts with `-` (e.g., `-o IdentitiesOnly=yes`)  
**When:** Constructing `--ssh-extra-args` with this value  
**Then:** The formatted argument is: `--ssh-extra-args=-o IdentitiesOnly=yes`  
**And:** The value's leading `-` does not cause parsing confusion

#### Scenario: Handling values with spaces
**Given:** A limit filter contains spaces (e.g., `web01 OR web02`)  
**When:** Constructing `--limit` with this value  
**Then:** The formatted argument is: `--limit=web01 OR web02`  
**And:** The spaces are preserved within the single argument element

### Requirement: Implementation Consistency

All argument construction across provisioner and galaxy files MUST follow the same pattern.

#### Scenario: Consistent formatting in provisioner.go
**Given:** The ansible-navigator provisioner constructs arguments  
**When:** Adding any flag with a value  
**Then:** All flag-value pairs must use `fmt.Sprintf("--flag=%s", value)` pattern  
**And:** No mixed patterns exist in the codebase

#### Scenario: Consistent formatting in galaxy.go
**Given:** The galaxy manager constructs arguments  
**When:** Adding flag-value pairs for install commands  
**Then:** All patterns match the provisioner formatting standard  
**And:** Short flags use `-f=value` format

#### Scenario: Test fixture compatibility
**Given:** Existing test fixtures validate argument construction  
**When:** Tests execute after standardization  
**Then:** All tests pass with the new format  
**And:** Test assertions check for `--flag=value` format  
**And:** No tests expect the old `--flag value` format

