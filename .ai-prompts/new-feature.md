Perfect — here’s a complete **`new-feature.md`** you can drop into your repo.
It’s written in the same *Agents/AI instruction style* as your previous `Agents.md`, but tailored for **Claude Sonnet** (or any structured AI developer assistant).

This version defines the *“Managed Collections”* feature, the extra behavioral flags, and the environment updates for Go 1.25.3 throughout the repo.

---

# 🧩 **new-feature.md**

## 🎯 Objective: Add Managed Ansible Collections Support

We are extending **`packer-plugin-ansible-navigator`** to support **managed Ansible collections**, improving reproducibility and flexibility for users who rely on custom or local collections.

---

## 🧭 Context

This plugin is written in Go **1.25.3**, using the latest Packer Plugin SDK.

Claude: ensure that all build references, `go.mod` declarations, and documentation in this repository reflect this version.
Specifically:

* Update `go.mod` to use `go 1.25.3`.
* Update the `README.md`, `Agents.md`, and `website` docs to mention Go 1.25.3 where applicable.
* Ensure the GitHub Actions or CI configuration pins `setup-go@v5` (or later) to `go-version: 1.25.3`.

---

## ⚙️ Feature Overview: “Managed Collections”

### 🧱 Purpose

Allow users to define collections declaratively in their Packer HCL templates.
The plugin will:

* Automatically install any required collections using `ansible-galaxy collection install`.
* Cache downloaded collections for reuse.
* Support local paths and pinned versions.
* Expose extended behaviors for offline or force-update modes.

---

## 🧩 Implementation Tasks

Claude: perform the following in a new branch (`feature/managed-collections`).

### 1. 🧭 Update Config Schema

Modify `config.go` to include:

```go
type Config struct {
    Collections           []string `mapstructure:"collections"`
    CollectionsCacheDir   string   `mapstructure:"collections_cache_dir"`
    CollectionsOffline    bool     `mapstructure:"collections_offline"`
    CollectionsForceUpdate bool    `mapstructure:"collections_force_update"`
    CollectionsRequirements string `mapstructure:"collections_requirements"`
}
```

#### Behavior summary:

| Field                      | Description                                                                              |
| -------------------------- | ---------------------------------------------------------------------------------------- |
| `collections`              | List of collections (name[:version] or name@path).                                       |
| `collections_cache_dir`    | Directory to store cached collections. Default: `~/.packer.d/ansible_collections_cache`. |
| `collections_offline`      | Skip network fetch; fail if collection not present locally.                              |
| `collections_force_update` | Always reinstall collections even if cached.                                             |
| `collections_requirements` | Path to a `requirements.yml` file, alternative to inline list.                           |

---

### 2. 🛠️ Add Installer Logic (`install_collections.go`)

Create a helper implementing:

```go
func ensureCollections(ui packer.Ui, c *Config) error
```

Core steps:

1. Resolve `CollectionsCacheDir` → absolute path (`~` → user home).
2. If `collections_requirements` is set:

   ```bash
   ansible-galaxy collection install -r <file> -p <cacheDir>
   ```
3. Else iterate over each entry in `Collections`:

   * Parse version/path syntax.
   * Skip if offline mode and missing.
   * If `CollectionsForceUpdate` or not cached, run:

     ```bash
     ansible-galaxy collection install <spec> -p <cacheDir>
     ```
4. On failure, return clear error:

   ```
   ERROR: Failed to install collection community.general:5.11.0
   Reason: exit status 2
   ```
5. Validate install by checking for `MANIFEST.json` under:

   ```
   <cacheDir>/ansible_collections/<namespace>/<name>/
   ```

All stdout/stderr should be streamed to Packer’s UI (`ui.Message`, `ui.Error`).

---

### 3. 🌍 Environment Integration

Before executing `ansible-navigator run`, set:

```go
os.Setenv("ANSIBLE_COLLECTIONS_PATHS", c.CollectionsCacheDir)
```

If the variable already exists, append:

```go
os.Setenv("ANSIBLE_COLLECTIONS_PATHS", c.CollectionsCacheDir+":"+os.Getenv("ANSIBLE_COLLECTIONS_PATHS"))
```

---

### 4. 🧪 Validation and Testing

Add unit tests under `/provisioner/tests/`:

* ✅ Missing collection in offline mode → error.
* ✅ Cached collection → skip reinstall.
* ✅ Force update → reinstall regardless.
* ✅ Local path install (`@/path/to/collection`).
* ✅ Requirements file install.

Example:

```go
func TestEnsureCollections_OfflineMissing(t *testing.T) {
  cfg := Config{
    Collections: []string{"myorg.missing"},
    CollectionsOffline: true,
    CollectionsCacheDir: t.TempDir(),
  }
  err := ensureCollections(testUi, &cfg)
  if err == nil {
    t.Fatalf("expected error in offline mode for missing collection")
  }
}
```

---

### 5. 🧩 Docs and Examples

Update `/website/docs/provisioner/ansible-navigator.mdx` and `README.md` to include examples:

```hcl
provisioner "ansible-navigator" {
  collections = [
    "community.general:5.11.0",
    "myorg.mycollection@/opt/custom/collections"
  ]
  collections_cache_dir = "~/.packer.d/ansible_collections_cache"
  collections_force_update = true
  playbook_file = "deploy.yml"
}
```

and:

```hcl
# requirements.yml example
provisioner "ansible-navigator" {
  collections_requirements = "./requirements.yml"
}
```

---

### 6. 📦 Build and Go Version Update

* Ensure `go.mod` contains:

  ```go
  go 1.25.3
  ```
* Update `.github/workflows/build.yml` or equivalent CI to include:

  ```yaml
  - uses: actions/setup-go@v5
    with:
      go-version: '1.25.3'
  ```
* Run:

  ```bash
  go mod tidy
  go build -o ~/.packer.d/plugins/packer-plugin-ansible-navigator
  ```

---

## ⚠️ Error Reporting and User Feedback

Claude: implement clear, structured error messages.

| Scenario           | Message                                                                     |
| ------------------ | --------------------------------------------------------------------------- |
| Missing binary     | `Error: ansible-galaxy not found in PATH.`                                  |
| Missing collection | `Collection 'community.general' not found and offline mode is enabled.`     |
| Install failed     | `Failed to install collection 'myorg.tools': exit code 1.`                  |
| Cache invalid      | `Reinstalling collection 'community.general' due to missing MANIFEST.json.` |

All messages should appear via:

```go
ui.Error(msg)
return fmt.Errorf(msg)
```

---

### 🧠 Notes for AI Agent Behavior

* Maintain backward compatibility for users not defining collections.
* Ensure all added parameters are optional.
* Keep error handling explicit and fail-fast.
* No implicit network operations when `collections_offline` is true.
* Use idiomatic Go patterns, no shell wrappers unless necessary.

---

### ✅ Completion Criteria

* The plugin builds cleanly under Go 1.25.3.
* All new configuration fields are recognized and documented.
* Collections install, cache, and resolve correctly in both online and offline modes.
* `go test ./...` passes for all new functionality.

---

### 📚 Reference

* [Ansible Collections Install Guide](https://docs.ansible.com/ansible/latest/collections_guide/collections_installing.html)
* [Ansible Navigator CLI Reference](https://ansible.readthedocs.io/projects/navigator/en/latest/)
* [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk)
