Perfect — here’s your **`new-feature-json-logging.md`** feature spec, written in the same “Agents.md” style as your other design prompts.

It defines how to add **JSON output parsing and structured error reporting** to your `packer-plugin-ansible-navigator`, building on the `navigator_mode = "json"` feature we just discussed.

---

# 🧩 **new-feature-json-logging.md**

## 🎯 Objective: Add JSON Output Parsing for Structured Logging and Enhanced Error Reporting

We are extending **`packer-plugin-ansible-navigator`** to parse structured JSON output from `ansible-navigator` runs.
This allows the plugin to detect individual task results, report specific failures, and optionally produce structured logs for downstream tools and CI/CD integrations.

---

## 🧭 Context

* The plugin currently executes `ansible-navigator` in `--mode stdout` or `--mode json`.
* In JSON mode, `ansible-navigator` emits machine-readable event objects describing the Ansible run.
* This feature will consume that JSON event stream, enabling:

  * Fine-grained task-level failure detection
  * Improved build output clarity
  * Optional structured artifact creation (e.g., JSON summary file)

This enhancement is **fully compatible** with Go 1.25.3 and the new unified play execution system.

---

## ⚙️ Feature Overview: JSON Event Parsing

### Behavior

When `navigator_mode = "json"`:

1. The plugin captures `ansible-navigator`’s stdout as a JSON event stream.
2. Each event is parsed and evaluated in real time.
3. Task, play, and host outcomes are logged to Packer’s UI in a readable format.
4. Failures are summarized precisely at the end of the build.
5. A structured summary file may optionally be written to disk.

### Output Example

```
==> ansible-navigator: Play 'Configure SSH' started (2 tasks)
==> ansible-navigator: Task 'Ensure SSHD running' [web1.example.com]: OK
==> ansible-navigator: Task 'Ensure SSHD running' [web2.example.com]: FAILED (rc=2)
==> ansible-navigator: Play 'Configure SSH' failed on 1 host(s).
```

---

## 🧱 Configuration Schema Additions

Add the following fields to `Config`:

```go
type Config struct {
    NavigatorMode    string `mapstructure:"navigator_mode"`
    StructuredLogging bool   `mapstructure:"structured_logging"`
    LogOutputPath     string `mapstructure:"log_output_path"`
}
```

| Field                | Description                                    | Default         |
| -------------------- | ---------------------------------------------- | --------------- |
| `structured_logging` | Enables structured JSON parsing and reporting. | `false`         |
| `log_output_path`    | Optional path to write parsed summary JSON.    | `""` (disabled) |

---

## 🧩 Example HCL Configuration

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  navigator_mode = "json"
  structured_logging = true
  log_output_path = "./logs/ansible-summary.json"

  plays = [
    { name = "Provision base system", target = "geerlingguy.docker" },
    { name = "Deploy app", target = "myorg.webserver.deploy" }
  ]
}
```

---

## ⚙️  Implementation Tasks

Claude: perform these changes in a new branch `feature/json-logging`.

### 1️⃣ Capture JSON Stream

Modify the command execution for plays:

```go
cmd := exec.Command("ansible-navigator", "run", play.Target, "--mode", "json")
stdoutPipe, _ := cmd.StdoutPipe()
stderrPipe, _ := cmd.StderrPipe()
cmd.Start()
```

Use a streaming JSON decoder:

```go
decoder := json.NewDecoder(stdoutPipe)
for decoder.More() {
    var event NavigatorEvent
    if err := decoder.Decode(&event); err != nil {
        continue
    }
    handleNavigatorEvent(ui, &event, &summary)
}
```

---

### 2️⃣ Define Event Schema

Create `event.go`:

```go
type NavigatorEvent struct {
    Event      string                 `json:"event"`
    UUID       string                 `json:"uuid"`
    Counter    int                    `json:"counter"`
    Task       string                 `json:"task,omitempty"`
    Play       string                 `json:"play,omitempty"`
    Host       string                 `json:"host,omitempty"`
    Status     string                 `json:"status,omitempty"`
    Stdout     string                 `json:"stdout,omitempty"`
    StdErr     string                 `json:"stderr,omitempty"`
    Data       map[string]interface{} `json:"data,omitempty"`
}
```

These correspond to Ansible Runner event types like:

* `playbook_on_start`
* `runner_on_ok`
* `runner_on_failed`
* `playbook_on_stats`

---

### 3️⃣ Process and Report Events

In `handleNavigatorEvent`:

```go
func handleNavigatorEvent(ui packer.Ui, e *NavigatorEvent, summary *Summary) {
    switch e.Event {
    case "playbook_on_start":
        ui.Message(fmt.Sprintf("Play '%s' started.", e.Play))
    case "runner_on_ok":
        ui.Message(fmt.Sprintf("Task '%s' succeeded on %s", e.Task, e.Host))
    case "runner_on_failed":
        ui.Error(fmt.Sprintf("Task '%s' failed on %s", e.Task, e.Host))
        summary.FailedTasks = append(summary.FailedTasks, e)
    case "playbook_on_stats":
        ui.Message("Playbook completed.")
        summary.TotalTasks = e.Data["changed"].(int) + e.Data["ok"].(int)
    }
}
```

---

### 4️⃣ Summarize Results

Create a simple summary struct:

```go
type Summary struct {
    PlaysRun     int
    TasksTotal   int
    TasksFailed  int
    FailedTasks  []NavigatorEvent
}
```

At the end of execution:

```go
if len(summary.FailedTasks) > 0 {
    ui.Error(fmt.Sprintf("Summary: %d failed task(s) detected.", len(summary.FailedTasks)))
}
if c.LogOutputPath != "" {
    writeSummaryJSON(summary, c.LogOutputPath)
}
```

---

### 5️⃣ Write JSON Artifact (Optional)

```go
func writeSummaryJSON(summary *Summary, path string) {
    f, err := os.Create(path)
    if err != nil {
        return
    }
    defer f.Close()
    json.NewEncoder(f).Encode(summary)
}
```

Resulting file example:

```json
{
  "PlaysRun": 2,
  "TasksTotal": 24,
  "TasksFailed": 1,
  "FailedTasks": [
    {
      "Task": "Restart nginx",
      "Host": "web1.example.com",
      "Status": "failed"
    }
  ]
}
```

---

## 🧪 Testing Scenarios

Add `/provisioner/tests/json_logging_test.go`:

| Scenario                     | Expected Behavior                            |
| ---------------------------- | -------------------------------------------- |
| `structured_logging = false` | Plain output; no parsing.                    |
| `structured_logging = true`  | Decodes JSON events, produces summary.       |
| Single failed task           | Reports failure clearly, sets exit code ≠ 0. |
| Multiple plays               | Summarizes across plays.                     |
| `log_output_path` set        | File written successfully.                   |
| Invalid JSON                 | Skips gracefully with warning.               |

Use a mock JSON stream for unit testing — no need for real Ansible runs.

---

## ⚙️ Error Handling

| Condition          | Message                                                           |
| ------------------ | ----------------------------------------------------------------- |
| Invalid JSON       | `[Warning] Skipped malformed JSON event: %v`                      |
| No events received | `[Warning] No valid events parsed from ansible-navigator output.` |
| Failed tasks found | `[Error] %d task(s) failed during play execution.`                |
| Log write failure  | `[Warning] Could not write structured log to %s`                  |

All messages should use `ui.Message()` or `ui.Error()` to integrate with Packer’s logging subsystem.

---

## 📦 Versioning and Build

* Bump version to `v0.3.0`.
* Update `go.mod` and `README.md` to note JSON logging support.
* Ensure Go 1.25.3 is reflected in all build scripts and CI actions.

---

## 🧰 Developer Notes

* JSON parsing should stream — **never buffer full output** to avoid memory overhead on large playbooks.
* Keep event schema minimal; add fields as needed for error summaries.
* This lays groundwork for structured build dashboards or integration with telemetry (e.g., Prometheus exporters).
* Consider a future flag `structured_logging_format = "jsonl"` to output JSON lines instead of a single object.

---

## ✅ Completion Criteria

* `navigator_mode = "json"` triggers JSON parsing.
* When `structured_logging = true`, the plugin produces detailed per-task logs.
* Failed tasks are clearly reported with host and task name.
* Optional structured summary file written to disk if `log_output_path` provided.
* Non-JSON modes remain unchanged.
* All new code tested and documented.

---

### 📚 Reference

* [Ansible Runner Event Schema](https://ansible.readthedocs.io/projects/runner/en/latest/intro.html#event-data)
* [Ansible Navigator CLI Reference](https://ansible.readthedocs.io/projects/navigator/en/latest/)
* [Go JSON streaming decoder](https://pkg.go.dev/encoding/json#Decoder)
* [HashiCorp Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk)

---

**End of file — new-feature-json-logging.md**
This document defines the structured JSON parsing system for `ansible-navigator` output, providing task-level feedback, summarized reporting, and optional artifact creation for integration with Packer logs and CI pipelines.
