Perfect — here’s your new **`new-feature-plays.md`** document, written in the same structured “AI agent instruction” style as your earlier `new-feature.md`, but updated for the *unified requirements system*, *structured plays*, *role FQDN support*, and *runtime flexibility*.

---

# 🧩 **new-feature-plays.md**

## 🎯 Objective: Unify Play and Role Execution Model

We are extending **`packer-plugin-ansible-navigator`** to introduce a unified play execution model.
This update replaces `playbook_file` with a new **`plays`** array supporting both **Ansible playbooks and roles (including FQDNs)**, while merging dependency management into a single `requirements_file`.

---

## 🧭 Context

This repository is built with **Go 1.25.3** — ensure all build, CI, and documentation references are consistent with that version.
The provisioner already implements managed **collections** and **roles** caching; this feature expands the execution layer to handle multiple plays with per-play customization.

---

## ⚙️ Feature Overview: Unified Play Execution

The plugin must now:

1. Replace `playbook_file` with a **`plays`** array supporting structured play definitions.
2. Accept either playbooks (`.yml/.yaml`) or fully-qualified role names (FQRNs).
3. Support per-play flags (become, tags, vars, etc.).
4. Merge dependency management (collections + roles) under one field:

   ```hcl
   requirements_file = "./requirements.yml"
   ```
5. Maintain backward compatibility with `playbook_file` if `plays` is not defined.

---

## 🧱 Configuration Schema Additions

Update `config.go`:

```go
type Config struct {
    RequirementsFile string `mapstructure:"requirements_file"`

    CollectionsCacheDir string `mapstructure:"collections_cache_dir"`
    RolesCacheDir       string `mapstructure:"roles_cache_dir"`
    OfflineMode         bool   `mapstructure:"offline_mode"`
    ForceUpdate         bool   `mapstructure:"force_update"`

    Plays []Play `mapstructure:"plays"`
}

type Play struct {
    Name       string            `mapstructure:"name"`
    Target     string            `mapstructure:"target"`
    ExtraVars  map[string]string `mapstructure:"extra_vars"`
    Tags       []string          `mapstructure:"tags"`
    VarsFiles  []string          `mapstructure:"vars_files"`
    Become     bool              `mapstructure:"become"`
}
```

### Behavior summary

| Field                                       | Description                                                             |
| ------------------------------------------- | ----------------------------------------------------------------------- |
| `requirements_file`                         | Path to a YAML file containing both roles and collections requirements. |
| `plays`                                     | Array of play definitions (playbook paths or role FQRNs).               |
| `offline_mode`                              | Prevent network fetches, fail if missing in cache.                      |
| `force_update`                              | Force reinstall dependencies.                                           |
| `collections_cache_dir` / `roles_cache_dir` | Cache directories for Galaxy installs.                                  |

---

## 🧩 Example Usage in HCL

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  collections_cache_dir = "~/.packer.d/ansible_collections_cache"
  roles_cache_dir       = "~/.packer.d/ansible_roles_cache"
  offline_mode          = false
  force_update          = false

  plays = [
    {
      name   = "Setup base system"
      target = "geerlingguy.docker"
      extra_vars = { docker_install_compose = true }
    },
    {
      name   = "Deploy web stack"
      target = "myorg.webserver.deploy"
      become = true
      vars_files = ["vars/web.yml"]
    },
    {
      name   = "Custom playbook test"
      target = "site.yml"
      tags   = ["test"]
    }
  ]
}
```

---

## 🔧 Implementation Tasks

Claude: perform these steps in a new branch `feature/unified-plays`.

### 1️⃣ Update Configuration Parsing

* Implement new structs in `config.go`.
* Maintain backward compatibility:

  * If `plays` is empty and `playbook_file` exists, convert it to one `Play` entry internally.
* Validate mutually exclusive fields at load time.

---

### 2️⃣ Dependency Installer Update

Modify your dependency preparation logic (currently in `ensureCollections` / `ensureRoles`) to also:

* Accept a **single `requirements_file`**:

  ```bash
  ansible-galaxy install -r requirements.yml
  ```
* Detect and install both roles and collections automatically.
* Use cache dirs and honor `offline_mode` / `force_update`.

---

### 3️⃣ Play Execution Logic

Implement a helper in `provisioner.go`:

```go
func runPlays(ui packer.Ui, c *Config) error
```

Loop through each play:

```go
for _, play := range c.Plays {
    args := []string{"run"}

    // Determine target type
    if strings.HasSuffix(play.Target, ".yml") || strings.HasSuffix(play.Target, ".yaml") {
        args = append(args, play.Target)
    } else {
        // Treat as role (FQDN or short role name)
        tmpPlaybook := generateRolePlaybook(play.Target, play.ExtraVars)
        args = append(args, tmpPlaybook)
    }

    if play.Become {
        args = append(args, "--become")
    }
    for _, tag := range play.Tags {
        args = append(args, "--tags", tag)
    }
    for k, v := range play.ExtraVars {
        args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", k, v))
    }

    cmd := exec.Command("ansible-navigator", args...)
    cmd.Stdout = uiWriter
    cmd.Stderr = uiWriter

    if err := cmd.Run(); err != nil {
        ui.Error(fmt.Sprintf("Play '%s' failed: %v", play.Name, err))
        return err
    }
}
```

### Helper for roles

```go
func generateRolePlaybook(role string, vars map[string]string) string {
    // Create temp YAML file like:
    // - hosts: all
    //   roles:
    //     - role: geerlingguy.docker
    //       vars:
    //         key: value
}
```

---

### 4️⃣ Environment Setup

Ensure both dependency paths are available before running any play:

```go
os.Setenv("ANSIBLE_COLLECTIONS_PATHS", c.CollectionsCacheDir)
os.Setenv("ANSIBLE_ROLES_PATH", c.RolesCacheDir)
```

If either variable already exists, prepend rather than overwrite.

---

### 5️⃣ Error Handling and Reporting

Provide detailed, contextual messages:

| Scenario            | Message                                                             |
| ------------------- | ------------------------------------------------------------------- |
| Missing dependency  | `Error: Unable to locate required dependency from requirements.yml` |
| Invalid play target | `Invalid play target 'foo.txt' — must be .yml/.yaml or role FQDN.`  |
| Play failure        | `Play 'Deploy web stack' failed with exit code 2`                   |
| Offline missing     | `Role 'geerlingguy.docker' not found and offline mode enabled.`     |

All error messages must be printed via `ui.Error` and returned as Go errors.

---

### 6️⃣ Unit Tests

Add `/provisioner/tests/plays_test.go`:

* ✅ Runs single playbook.
* ✅ Runs multiple FQDN roles.
* ✅ Handles per-play variables.
* ✅ Validates backward compatibility for `playbook_file`.
* ✅ Fails gracefully in offline mode with missing dependencies.

---

## 🧪 Testing Scenarios

| Scenario                                              | Expected Result                          |
| ----------------------------------------------------- | ---------------------------------------- |
| `plays` contains .yml                                 | Executes playbook normally               |
| `plays` contains role FQDN                            | Generates temporary playbook and runs it |
| `offline_mode = true` with missing dep                | Fails early                              |
| `force_update = true`                                 | Reinstalls all deps                      |
| `requirements_file` contains both roles & collections | Installs both types successfully         |

---

## 📦 Build and Version

Ensure:

```go
go 1.25.3
```

and GitHub Actions:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.25.3'
```

Update all documentation and `go.mod` references accordingly.

---

## 🧰 Developer Notes

* Keep `playbook_file` as a deprecated field for one minor version; print a warning if used.
* Maintain consistent caching semantics between roles and collections.
* Use `ansible-navigator` for execution, but prefer `ansible-galaxy` directly for installs (faster).
* Log all dependency and play activity via `ui.Message()` for better user feedback.

---

## ✅ Completion Criteria

* The plugin executes multiple plays in sequence (playbook and FQDN roles).
* `requirements_file` fully replaces the older dual requirements variables.
* Offline and force-update modes function identically for both dependency types.
* Backward compatibility for `playbook_file` maintained (with deprecation warning).
* Tests and documentation reflect new schema.

---

### 📚 Reference

* [Ansible Galaxy requirements syntax](https://docs.ansible.com/ansible/latest/collections_guide/collections_installing.html)
* [Ansible Navigator CLI Reference](https://ansible.readthedocs.io/projects/navigator/en/latest/)
* [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk)
