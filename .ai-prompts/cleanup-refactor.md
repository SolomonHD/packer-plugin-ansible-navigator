Perfect — here’s a cleanup-oriented **AI engineering prompt** you can feed directly to Claude Opus (or a similar large model) to do a *deep refactor and consistency pass* on your plugin repo.
It’s tuned for a repository that has evolved through incremental AI-generated features (like yours), and focuses on **reducing redundancy, ensuring naming consistency, tightening logic, and standardizing structure** without changing core behavior.

---

# 🧹 **cleanup-refactor.md**

## 🎯 Objective

Perform a **comprehensive cleanup and refactor** of the `packer-plugin-ansible-navigator` repository.
The goal is to improve readability, naming consistency, and internal structure — without changing functionality.

The plugin is written in **Go 1.25.3**, built as a **standalone Packer plugin** (not an inline provisioner), integrating with **Ansible Navigator** for provisioning automation.

---

## 🧭 Context

The codebase has evolved across several AI-generated feature branches (`unified-plays`, `runmode`, `json-logging`, etc.), which may have introduced:

* Redundant helper functions or overlapping logic (especially around command execution, JSON parsing, and UI handling).
* Inconsistent naming conventions based on incremental steps (`runPlays2`, `generateRolePlaybook_v3`, etc.).
* Fragmented configuration validation and inconsistent error reporting.
* Minor code duplication between role/playbook execution paths.
* Unused structs, debug prints, or experimental variables left behind.

You must perform a **safe, semantic cleanup** while preserving all working behaviors and feature coverage.

---

## 🧩 Tasks

### 1️⃣ Consolidate Functions

* Merge redundant or near-duplicate functions (e.g., multiple “run play” or “install deps” variants).
* Remove step-based suffixes (`_v2`, `_tmp`, `_new`, `_testonly`, etc.) and replace with **semantic names**.
* Ensure function names describe *intent* rather than *chronology*.
  Examples:

  * `runPlays3()` → `executePlays()`
  * `handleNavigatorEventNew()` → `processNavigatorEvent()`
  * `ensureCollections()` + `ensureRoles()` → `installRequirements()`
  * `writeSummaryJSON2()` → `writeJSONSummary()`

---

### 2️⃣ Refactor Configuration Structures

* Verify `Config` struct fields are grouped logically and alphabetically where possible.
* Ensure every config field is documented with a concise comment explaining its purpose.
* Remove any deprecated or unused fields (`playbook_file`, `collections_requirements`, etc.) if the deprecation window is complete.
* Centralize config validation in one function, e.g.:

  ```go
  func (c *Config) Validate() error
  ```

---

### 3️⃣ Unify Error Handling and Logging

* Standardize how errors and warnings are emitted:

  * Always use `ui.Error()` for failures.
  * Always use `ui.Message()` for informational output.

* Create helper functions like:

  ```go
  func logError(ui packer.Ui, msg string, err error)
  func logInfo(ui packer.Ui, msg string)
  ```

  Then use them consistently across the repo.

* Ensure all returned errors have context (`fmt.Errorf("failed to install requirements: %w", err)`).

---

### 4️⃣ Streamline Execution Path

* All play/role execution should go through a **single unified code path**.
* Detect `.yml/.yaml` vs role FQDN automatically — no duplicated switch blocks.
* Encapsulate subprocess execution in one helper:

  ```go
  func runNavigator(ui packer.Ui, args []string, mode string) error
  ```

  That handles environment variables, stdout/stderr wiring, and JSON parsing if enabled.

---

### 5️⃣ Clean Up Event and Summary Handling

* Ensure `NavigatorEvent` and `Summary` structs only contain necessary fields.
* Remove unused JSON tags, placeholder fields, or debugging properties.
* Extract event-handling logic into its own package or file (`events.go`).
* Ensure event constants are defined centrally if reused.

---

### 6️⃣ Remove Redundancies

* Identify repeated logic across play loops, JSON summaries, and dependency checks.
* Use helper functions to DRY up:

  * Environment variable setup (`ANSIBLE_COLLECTIONS_PATHS`, `ANSIBLE_ROLES_PATH`).
  * Temporary playbook generation for roles.
  * Summary reporting (success/failure counts).
* Verify no unused imports, commented code, or print/debug statements remain.

---

### 7️⃣ Improve Naming and Structure

* Functions, structs, and variables should follow Go idioms:

  * Functions: `camelCase` (verbs for actions).
  * Structs: `PascalCase`.
  * No numbering or temporary suffixes.

* Use consistent prefixes:

  * `runNavigator...` for anything that executes ansible-navigator.
  * `install...` for dependency management.
  * `parse...` for JSON event decoding.

* Rename test files and functions to be self-descriptive:

  * `plays_test.go` → `executor_test.go`
  * `TestUnifiedPlays()` → `TestExecutePlaysMultiple()`

---

### 8️⃣ Documentation and Comments

* Ensure every exported function, struct, and type has a GoDoc comment.
* Keep comments factual and concise — avoid duplication of obvious code logic.
* Update `README.md` and `docs/` references to reflect the cleaned config structure and parameters (`requirements_file`, `navigator_mode`, etc.).

---

### 9️⃣ Consistency in CLI and Logging Output

* Verify all CLI output lines start with a consistent prefix (e.g., `==> ansible-navigator:`).
* Ensure error summaries use the same phrasing across plays, roles, and dependencies:

  ```
  Error: Play 'Configure nginx' failed on 2 host(s).
  ```
* Use consistent capitalization and punctuation in messages.

---

### 🔟 Optional Improvements (if time allows)

* Modularize:

  * Move event parsing into `/internal/events/`.
  * Move requirement installation logic into `/internal/requirements/`.
  * Keep `/provisioner/` for orchestration and runtime logic.
* Simplify test data: replace mock JSON streams with small test fixtures under `/testdata/`.

---

## 🧪 Validation Checklist

Before cleanup completion:

* [ ] All unit tests pass (`go test ./...`).
* [ ] Plugin builds cleanly (`go build ./cmd/packer-plugin-ansible-navigator`).
* [ ] No unused functions or imports (`go vet`, `staticcheck`).
* [ ] All configs validated via `packer validate`.
* [ ] `go fmt`, `goimports`, and `golangci-lint` produce zero issues.
* [ ] Changelog updated with summary of cleanup and renamed functions.

---

## 📦 Deliverables

After cleanup, output a short summary file:

```
CLEANUP_SUMMARY.md
```

containing:

* Number of functions merged or renamed.
* List of removed redundant files/functions.
* Any renamed structs or exported symbols.
* Confirmation of passing tests.

---

## ✅ Completion Criteria

| Area          | Success Metric                                  |
| ------------- | ----------------------------------------------- |
| Redundancy    | No duplicate functions or config fields.        |
| Naming        | All identifiers semantic, no numbered suffixes. |
| Validation    | Unified validation logic.                       |
| Logging       | Consistent output formatting.                   |
| JSON Handling | Centralized and documented.                     |
| CI            | All builds/tests/lints pass cleanly.            |

---

## ⚙️ Environment

* Go version: **1.25.3**
* Packer Plugin SDK: latest compatible version
* `ansible-navigator`: assumed available in `$PATH`

---

## 📚 Notes for AI (Claude Opus)

* Do **not** remove features — cleanup only.
* Maintain all behavior from unified-plays, runmode, and json-logging features.
* You may reorder functions, refactor into helper files, or rename for clarity.
* Minimize file churn unless justified by consolidation.
* Run a full `go test ./...` check after refactoring.

---

**End of file — cleanup-refactor.md**
This prompt instructs a deep cleanup pass to remove redundancies, rename semantically, consolidate logic, and ensure clarity, while maintaining all existing functionality and Go 1.25.3 compatibility.

---

Would you like me to extend this with an **optional “refactor commit message template”** (for use after Opus finishes), so your final commit history stays clean and well-documented?
