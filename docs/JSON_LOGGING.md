# JSON Logging and Structured Output

The ansible-navigator provisioner supports parsing structured JSON output from ansible-navigator runs, enabling enhanced error reporting, task-level feedback, and integration with CI/CD pipelines.

## Overview

When `navigator_mode = "json"` and `structured_logging = true`, the provisioner will:

1. Parse ansible-navigator's JSON event stream in real-time
2. Display detailed task-level status information
3. Provide specific failure reporting with host and task names
4. Optionally write a structured summary file for downstream tools

## Configuration

### Basic Configuration

```hcl
provisioner "ansible-navigator" {
  requirements_file   = "./requirements.yml"
  navigator_mode      = "json"
  structured_logging  = true

  play {
    name   = "Setup base system"
    target = "geerlingguy.docker"
  }

  play {
    name   = "Deploy application"
    target = "myorg.webserver.deploy"
  }
}
```

### With Summary File Output

```hcl
provisioner "ansible-navigator" {
  requirements_file   = "./requirements.yml"
  navigator_mode      = "json"
  structured_logging  = true
  log_output_path     = "./logs/ansible-summary.json"

  play {
    name   = "Provision base system"
    target = "geerlingguy.docker"
  }

  play {
    name   = "Deploy app"
    target = "myorg.webserver.deploy"
  }
}
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `navigator_mode` | string | No | `"stdout"` | Must be set to `"json"` for structured logging to work |
| `structured_logging` | boolean | No | `false` | Enable JSON event parsing and enhanced reporting |
| `log_output_path` | string | No | `""` | Path to write structured summary JSON file (disabled if empty) |

## Output Examples

### Console Output

When structured logging is enabled, you'll see detailed task-level feedback:

```
==> ansible-navigator: Play 'Configure SSH' started
==> ansible-navigator: Task 'Ensure SSHD running' [web1.example.com]: OK
==> ansible-navigator: Task 'Ensure SSHD running' [web2.example.com]: FAILED
==> ansible-navigator: Play 'Configure SSH' completed
[Error] 1 task(s) failed during play execution.
  - Task 'Ensure SSHD running' on host 'web2.example.com'
Summary: 1 play(s) executed, 2 task(s) total, 1 failed
```

### Summary JSON File

When `log_output_path` is specified, a structured summary file is created:

```json
{
  "plays_run": 2,
  "tasks_total": 24,
  "tasks_failed": 1,
  "failed_tasks": [
    {
      "event": "runner_on_failed",
      "task": "Restart nginx",
      "host": "web1.example.com",
      "status": "failed",
      "uuid": "abc-123-def",
      "counter": 15
    }
  ]
}
```

## Event Types

The provisioner recognizes and processes the following ansible-navigator event types:

- `playbook_on_start` - Playbook execution begins
- `playbook_on_play_start` - Individual play starts
- `runner_on_ok` - Task succeeds on a host
- `runner_on_failed` - Task fails on a host
- `runner_on_skipped` - Task is skipped
- `runner_on_unreachable` - Host is unreachable
- `playbook_on_stats` - Final statistics
- `playbook_on_task_start` - Task execution begins

## Use Cases

### CI/CD Integration

Use the structured summary file for integration with CI/CD pipelines:

```hcl
provisioner "ansible-navigator" {
  navigator_mode      = "json"
  structured_logging  = true
  log_output_path     = "${path.root}/build-artifacts/ansible-summary.json"

  play {
    name   = "Deploy"
    target = "deploy.yml"
  }
}
```

The summary file can be parsed by your CI/CD system to:

- Generate detailed build reports
- Track deployment metrics
- Alert on specific failures
- Archive execution history

### Debugging and Troubleshooting

Enable structured logging during development to get detailed feedback:

```hcl
provisioner "ansible-navigator" {
  navigator_mode      = "json"
  structured_logging  = true
  log_output_path     = "./debug/ansible-run-${timestamp()}.json"

  play {
    name   = "Debug playbook"
    target = "debug.yml"
  }
}
```

### Production Auditing

Maintain execution logs for compliance and auditing:

```hcl
provisioner "ansible-navigator" {
  navigator_mode      = "json"
  structured_logging  = true
  log_output_path     = "/var/log/packer/ansible-${build.ID}.json"

  play {
    name   = "Production deployment"
    target = "production.yml"
  }
}
```

## Error Handling

### Invalid JSON

If ansible-navigator outputs malformed JSON, the provisioner will:

- Log a warning message
- Skip the invalid event
- Continue processing subsequent events

Example warning:

```
[Warning] Skipped malformed JSON event: unexpected EOF
```

### No Events Received

If no valid events are parsed:

```
[Warning] No valid events parsed from ansible-navigator output.
```

### Task Failures

Failed tasks are reported with details:

```
[Error] 2 task(s) failed during play execution.
  - Task 'Configure firewall' on host 'web1.example.com'
  - Task 'Restart service' on host 'web2.example.com'
```

## Backward Compatibility

Structured logging is opt-in and fully backward compatible:

- Default behavior (`structured_logging = false`) remains unchanged
- Works only when `navigator_mode = "json"`
- When disabled, output is streamed line-by-line as before

## Best Practices

1. **Enable for Production Builds**: Use structured logging in production to capture detailed execution data
2. **Archive Summary Files**: Store summary JSON files for historical analysis
3. **Monitor File Size**: Be aware that log files grow with playbook complexity
4. **Use Relative Paths**: When possible, use relative paths for `log_output_path` for portability
5. **Combine with Other Logs**: Use structured logs alongside Packer's native logging for comprehensive troubleshooting

## Limitations

- Structured logging only works when `navigator_mode = "json"`
- Summary file writes are best-effort (failures are logged but don't stop the build)
- Very large playbooks may generate large summary files
- Event parsing is based on Ansible Runner event schema

## Further Reading

- [Ansible Runner Event Schema](https://ansible.readthedocs.io/projects/runner/en/latest/intro.html#event-data)
- [Ansible Navigator CLI Reference](https://ansible.readthedocs.io/projects/navigator/en/latest/)
- [Packer Plugin SDK Documentation](https://github.com/hashicorp/packer-plugin-sdk)
