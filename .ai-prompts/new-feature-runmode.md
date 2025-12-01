Perfect — here’s your **`new-feature-runmode.md`** file written in the same *Agents instruction format* as your previous feature prompts.

This file adds the new `navigator_mode` option, ensures non-interactive behavior by default (`stdout`), and documents the rationale, fallback, and safe overrides.

---

# 🧩 **new-feature-runmode.md**

## 🎯 Objective: Add `navigator_mode` Support and Disable TUI by Default

We are extending **`packer-plugin-ansible-navigator`** to include a configuration option that controls how **`ansible-navigator`** executes — ensuring non-interactive operation within Packer builds.

By default, `ansible-navigator` launches a curses-based **TUI (Text User Interface)**, which is **incompatible with Packer’s non-interactive runtime**.
This feature introduces a `navigator_mode` parameter to enforce `--mode stdout` by default and optionally allow advanced users to specify alternate modes (e.g. `interactive`, `json`, or `yaml`).

---

## 🧭 Context

* **Default `ansible-navigator` behavior:** Launches a TUI that expects a human user.
* **Problem:** The TUI blocks execution and produces garbled output under Packer, CI/CD, or automated environments.
* **Solution:** Force `--mode stdout` unless explicitly overridden.

---

## ⚙️ Feature Overview: Controlled Run Mode

### New Config Parameter

Add a new configuration field to `Config`:

```go
type Config struct {
    NavigatorMode string `mapstructure:"navigator_mode"`
}
```

### Behavior Summary

| Field            | Description                                                                                    | Default    |
| ---------------- | ---------------------------------------------------------------------------------------------- | ---------- |
| `navigator_mode` | Execution mode for `ansible-navigator`. Valid values: `stdout`, `json`, `yaml`, `interactive`. | `"stdout"` |

If unset, the plugin defaults to non-interactive stdout mode.
If explicitly set, the value overrides the default.

---

## 🧱 Implementation Details

### 1️⃣ Update Configuration Parsing

In `config.go`, add:

```go
if c.NavigatorMode == "" {
    c.NavigatorMode = "stdout"
}
```

Validate that it’s one of the supported values:

```go
validModes := map[string]bool{
    "stdout": true,
    "json": true,
    "yaml": true,
    "interactive": true,
}
if !validModes[c.NavigatorMode] {
    return fmt.Errorf("invalid navigator_mode: %s (must be one of stdout, json, yaml, interactive)", c.NavigatorMode)
}
```

---

### 2️⃣ Apply Mode in Play Execution

In your play execution helper (`runPlays` or equivalent), when constructing the command:

```go
args := []string{"run", playbookPath, "--mode", c.NavigatorMode}
```

This must appear *before* other arguments (tags, vars, etc.).

If `NavigatorMode` is not explicitly set, it defaults to `"stdout"`.

---

### 3️⃣ Optional Environment Override (Safety Net)

If the environment variable `ANSIBLE_NAVIGATOR_MODE` is already set, prefer the config field but log a notice:

```go
if os.Getenv("ANSIBLE_NAVIGATOR_MODE") != "" {
    ui.Message(fmt.Sprintf("[Notice] Overriding ANSIBLE_NAVIGATOR_MODE with plugin value: %s", c.NavigatorMode))
}
os.Setenv("ANSIBLE_NAVIGATOR_MODE", c.NavigatorMode)
```

This ensures that even subprocesses launched from within Ansible respect the same mode.

---

### 4️⃣ Non-Interactive Behavior Enforcement

If the mode is set to `"interactive"` and no TTY is detected (i.e., running under Packer, CI, or other non-interactive contexts), warn and override automatically:

```go
if c.NavigatorMode == "interactive" && !term.IsTerminal(int(os.Stdout.Fd())) {
    ui.Message("[Warning] No TTY detected — switching ansible-navigator mode to 'stdout'.")
    c.NavigatorMode = "stdout"
}
```

This prevents builds from hanging or requiring manual input.

---

### 5️⃣ Integration with Other Features

The `navigator_mode` should apply consistently to **all play runs**, regardless of:

* Whether the target is a role or a playbook
* The number of plays in the `plays` list
* Dependency setup (collections/roles cache handling)

It should also be passed to any `ansible-navigator` command used internally (including validation or dry-run stages, if present).

---

## 🧩 Example Usage

### Default behavior (safe for Packer):

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  plays = [{ target = "site.yml" }]
}
```

Executes:

```bash
ansible-navigator run site.yml --mode stdout
```

### Explicit override:

```hcl
provisioner "ansible-navigator" {
  navigator_mode = "json"
  plays = [{ target = "site.yml" }]
}
```

Executes:

```bash
ansible-navigator run site.yml --mode json
```

### Interactive developer session:

```hcl
provisioner "ansible-navigator" {
  navigator_mode = "interactive"
  plays = [{ target = "site.yml" }]
}
```

If no TTY is detected, the plugin automatically switches to `stdout`.

---

## 🧪 Testing Requirements

Add unit tests under `/provisioner/tests/runmode_test.go`:

| Scenario                     | Expected Behavior                      |
| ---------------------------- | -------------------------------------- |
| No mode specified            | Defaults to `stdout`.                  |
| Explicit `json` mode         | Runs `--mode json`.                    |
| Invalid mode                 | Returns config validation error.       |
| Interactive mode with no TTY | Logs warning and switches to `stdout`. |
| Interactive mode with TTY    | Allows `--mode interactive`.           |

You can stub out `term.IsTerminal` for predictable test behavior.

---

## ⚙️ Error Handling and Reporting

| Scenario             | Message                                                                               |
| -------------------- | ------------------------------------------------------------------------------------- |
| Invalid mode         | `Error: invalid navigator_mode: foo (must be one of stdout, json, yaml, interactive)` |
| TTY unavailable      | `[Warning] No TTY detected — switching ansible-navigator mode to 'stdout'.`           |
| Environment override | `[Notice] Overriding ANSIBLE_NAVIGATOR_MODE with plugin value: stdout.`               |

All errors must surface via `ui.Error()` and warnings via `ui.Message()`.

---

## 📦 Versioning and Build

Ensure this feature is included in the next tagged release (e.g., `v0.2.0`), with the changelog entry:

```
### Added
- Added `navigator_mode` parameter to control ansible-navigator output mode.
- Default to `stdout` for non-interactive environments (packer-safe).
- Automatically switch from `interactive` → `stdout` when no TTY detected.
```

---

## 🧰 Developer Notes

* Use `--mode stdout` as the **hard default** to ensure compatibility with Packer and CI.
* Consider switching to `--mode json` by default in a future version if structured output parsing is added.
* Never launch `ansible-navigator` without an explicit `--mode` flag.
* This feature is mandatory for reliable automation — make it a core runtime setting.

---

## ✅ Completion Criteria

* `navigator_mode` is supported and validated.
* Builds no longer hang due to TUI mode.
* `--mode stdout` is always appended automatically.
* Interactive mode gracefully degrades in non-TTY sessions.
* Tests and documentation updated accordingly.

---

### 📚 Reference

* [Ansible Navigator CLI Reference](https://ansible.readthedocs.io/projects/navigator/en/latest/)
* [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk)
* [Go term.IsTerminal usage (golang.org/x/term)](https://pkg.go.dev/golang.org/x/term)
