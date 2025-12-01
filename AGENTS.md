<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

Perfect — turning your AI bootstrap prompt into an **`Agents.md`** document is an excellent idea. It gives both human and AI collaborators (like me or any GitHub Copilot-like assistant) a **clear mission brief** inside the repository itself.

Here’s a full, production-ready `Agents.md` you can drop into the root of your `packer-plugin-ansible-navigator` repo:

---

# 🧭 **Agents.md**

## 🏗️ Project Overview

This repository, **`packer-plugin-ansible-navigator`**, is a standalone HashiCorp Packer plugin written in Go.
It originated from the HashiCorp [`packer-plugin-ansible`](https://github.com/hashicorp/packer-plugin-ansible) codebase but is now an **independent project** developed under:

```
github.com/solomonhd/packer-plugin-ansible-navigator
```

The plugin enables provisioning of images using **Ansible Navigator** (`ansible-navigator run`), extending beyond traditional Ansible playbook execution.

---

## 🎯 Mission for AI Agents

Your task is to **evolve this plugin** into a fully featured, production-grade Packer provisioner that supports the full Ansible Navigator workflow, modern error handling, and clearer developer experience.

Follow the goals, constraints, and conventions outlined below.

---

## 🧩 Primary Objectives

1. **Rename and Rebrand**

   * Replace all identifiers, module paths, and documentation references from:

     ```
     packer-plugin-ansible → packer-plugin-ansible-navigator
     ```
   * Update plugin registration and display name:

     ```go
     "ansible" → "ansible-navigator"
     ```
   * Update Go module path in `go.mod` to:

     ```go
     module github.com/solomonhd/packer-plugin-ansible-navigator
     ```

2. **License**

   * Replace existing license with **Apache License 2.0**.
   * Add a `LICENSE` file with the full Apache 2.0 text.
   * Include this header in all new Go files:

     ```go
     // Licensed under the Apache License, Version 2.0 (the "License");
     // you may not use this file except in compliance with the License.
     // You may obtain a copy of the License at
     // http://www.apache.org/licenses/LICENSE-2.0
     ```

3. **Core Functionality**

   * Continue to use HashiCorp’s [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk).
   * Maintain SSH and WinRM communicator compatibility.
   * Replace invocations of:

     ```bash
     ansible-playbook
     ```

     with:

     ```bash
     ansible-navigator run
     ```

---

## 🧠 New Functionality

### 1. Dual Invocation Mode

Support two mutually exclusive configuration paths:

```hcl
# Option A: Traditional playbook file
playbook_file = "site.yml"

# Option B: Collection plays with structured configuration
plays = [
  {
    name = "Migrate Node"
    target = "integration.portainer.migrate_node"
    extra_vars = {
      environment = "production"
    }
  },
  {
    name = "Configure Firewall"
    target = "acme.firewall.configure_rules"
    vars_files = ["firewall_vars.yml"]
  }
]
```

* If both are set, return a config error:

  > “You may specify only one of `playbook_file` or `plays`.”
* If neither is set, return:

  > “Either `playbook_file` or `plays` must be defined.”

---

### 2. Error Handling and Reporting

Implement detailed error and progress feedback for AI and users.

#### On Play Execution Failure:

If running multiple plays (from `plays` array), report which play failed:

```plaintext
ERROR: Play 'integration.portainer.migrate_node' failed (exit code 2)
```

Continue printing:

```plaintext
Aborting remaining plays. Check the above output for the failing play.
```

#### On Config Validation Failure:

Provide user-focused messages during `Prepare()`:

```plaintext
Invalid configuration: both playbook_file and plays are set.
```

#### On Missing Dependencies:

If `ansible-navigator` binary is missing or not executable:

```plaintext
Error: ansible-navigator not found in PATH. Please install it before running this provisioner.
```

#### On Execution Errors:

* Stream `ansible-navigator`’s stdout/stderr directly to the Packer UI.
* Include context in Go errors:

  ```go
  return fmt.Errorf("ansible-navigator run failed for %s: %w", play, err)
  ```

---

### 3. Logging and UI Integration

Use the `packer.Ui` interface for consistent output:

```go
ui.Say(fmt.Sprintf("Running Ansible Navigator play: %s", play))
ui.Message(fmt.Sprintf("Execution environment: %s", c.ExecutionEnvironment))
ui.Error(fmt.Sprintf("Play %s failed: %v", play, err))
```

All error conditions must be surfaced through both:

* The UI (for console logs)
* Go `error` returns (for CI/CD and automation)

---

### 4. Testing and Validation

* Add unit tests under `/provisioner/ansible_navigator/tests/`
* Include scenarios for:

  * Multiple plays (success + failure)
  * Missing `ansible-navigator` binary
  * Invalid config (both fields set)
  * Normal single playbook execution

Example test pattern:

```go
func TestValidateConfig_MutualExclusion(t *testing.T) {
  cfg := Config{PlaybookFile: "a.yml", Plays: []string{"b.play"}}
  _, err := cfg.Prepare()
  if err == nil {
    t.Fatalf("expected error for mutual exclusivity")
  }
}
```

---

### 5. Documentation

Generate or update docs under `/website/docs/provisioner/ansible-navigator.mdx`:

**Example HCL:**

```hcl
provisioner "ansible-navigator" {
  plays = [
    {
      name = "Migrate Node"
      target = "integration.portainer.migrate_node"
      extra_vars = {
        node_type = "worker"
        cluster_id = "prod-01"
      }
      vars_files = ["production.yml"]
    }
  ]
  execution_environment = "ansible-execution-env:latest"
  extra_arguments = ["--mode", "stdout"]
  inventory_directory = "inventory/"
  work_dir = "/tmp/ansible"
}
```

Include error-handling notes and example log output.

---

## 🧰 Build and Local Test Checklist

Run the following to verify a successful build:

```bash
go mod tidy
go build -o ~/.packer.d/plugins/packer-plugin-ansible-navigator
packer init .
packer build template.pkr.hcl
```

Confirm output shows:

```
Running Ansible Navigator play: integration.portainer.migrate_node
...
Build completed successfully!
```

---

## 🧩 Summary

| Area            | Requirement                                            |
| --------------- | ------------------------------------------------------ |
| License         | Apache 2.0                                             |
| Module Path     | `github.com/solomonhd/packer-plugin-ansible-navigator` |
| Plugin Name     | `ansible-navigator`                                    |
| Error Reporting | Detailed per-play failure logs                         |
| Compatibility   | SSH + WinRM, Packer SDK                                |
| Primary Command | `ansible-navigator run`                                |
| Config Schema   | `playbook_file` or `plays[]` (mutually exclusive)      |

---

## 💬 Final Instruction for AI Agents

> Maintain clean, idiomatic Go code following Packer plugin conventions.
> Handle all errors explicitly, always surface failing play names, and ensure compatibility with the Packer Plugin SDK.
> Prioritize clear user messaging over silent failures.
