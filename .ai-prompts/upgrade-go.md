Here’s exactly what you asked for — a **prompt file** for an AI (such as GitHub Copilot Workspace, ChatGPT Code Interpreter, or Cursor Agent) that instructs it to upgrade your Packer plugin project to the **latest Go 2.x toolchain**, **modernize build logic**, and **ensure correct Packer 1.10+ artifact format**.

Below is the ready-to-use file contents.
Save it as:

```
upgrade-go.md
```

---

````markdown
# Agent Prompt: Upgrade Packer Plugin to Latest Go 2.x and Packer SDK Format

## Goal
Upgrade this Packer plugin repository to:
- The latest stable **Go 2.x** toolchain (or latest preview if Go 2 is not GA).
- The latest **packer-plugin-sdk v2.x** (or newer).
- Generate build artifacts that match the **Packer ≥ 1.10 plugin protocol (x5)** naming standard.
- Simplify Go code using any new Go 2.x language or tooling features (e.g. improved generics, error handling, collections, `for`/`if` syntax, modules layout, etc.).
- Ensure the project builds cleanly via **Go modules** and **GoReleaser v2**.
- Produce release assets suitable for `packer plugins install`.

---

## Context

### Current Project
This repository is a Packer plugin named **`packer-plugin-ansible-navigator`** written in Go 1.x.  
It currently:
- Uses a legacy Go toolchain.
- Depends on an older version of `packer-plugin-sdk`.
- Uses an outdated GoReleaser config (missing or incorrect `_x5_` artifact naming).

### Target Behavior
When the user runs:
```bash
packer plugins install github.com/SolomonHD/ansible-navigator vX.Y.Z
````

Packer should correctly fetch:

```
packer-plugin-ansible-navigator_vX.Y.Z_x5_linux_amd64.zip
packer-plugin-ansible-navigator_vX.Y.Z_x5_SHA256SUMS
```

and install successfully.

---

## Tasks

1. **Upgrade Toolchain**

   * Migrate `go.mod` and `go.sum` to Go 2.x (`go mod tidy`, `go mod upgrade`).
   * Ensure module syntax is updated (`go 2.0` or current stable Go version line).
   * Refactor any deprecated or incompatible Go 1.x syntax.

2. **Update SDK**

   * Require `github.com/hashicorp/packer-plugin-sdk/v2` at latest release.
   * Refactor imports to `packer-plugin-sdk/v2/...`.
   * Update plugin registration:

     ```go
     plugin.Serve(&plugin.ServeOpts{
       Name: "github.com/SolomonHD/ansible-navigator",
       Version: "vX.Y.Z",
       Components: []interface{}{
         new(mybuilder.Builder),
       },
     })
     ```

3. **Modernize Go Code**

   * Replace repetitive boilerplate with new Go 2.x features (simplified generics, improved error values, new `for`/`if` forms, unified type parameters, etc.).
   * Simplify error handling and concurrency if applicable.

4. **Revise GoReleaser Config**

   * Ensure `version: 2`.
   * Produce artifacts named:

     ```
     packer-plugin-ansible-navigator_v{{ .Version }}_x5_{{ .Os }}_{{ .Arch }}.zip
     packer-plugin-ansible-navigator_v{{ .Version }}_x5_SHA256SUMS
     ```
   * Linux amd64 build only.
   * Keep binary root-level (`wrap_in_directory: false`).

5. **Validate Release**

   * Run `goreleaser release --snapshot --clean`.
   * Confirm correct filenames and checksum output.
   * Verify installation via:

     ```bash
     packer plugins install github.com/SolomonHD/ansible-navigator vX.Y.Z
     ```

6. **Update Documentation**

   * Modify README.md to state minimum:

     * Go 2.x+
     * Packer ≥ 1.10
     * Packer Plugin SDK v2.x
   * Include usage snippet for `packer plugins install`.

---

## Deliverables

* ✅ Updated `go.mod` using Go 2.x syntax.
* ✅ Updated plugin source compatible with `packer-plugin-sdk/v2`.
* ✅ Updated `.goreleaser.yml` compliant with 2025 Packer plugin naming (`_x5_`).
* ✅ Verified snapshot build artifacts in `dist/`.
* ✅ Updated `README.md` reflecting new toolchain and SDK requirements.

---

## Notes

* If Go 2.x introduces new syntax not yet fully supported by the `packer-plugin-sdk`, preserve compatibility by pinning SDK v2 while using Go 2.x compiler features.
* Prefer clarity and maintainability over heavy use of new features.
* Maintain binary reproducibility for release automation.

---

**End of Prompt**

```markdown
---

Save that as `upgrade-go.md` and feed it directly into your AI workspace or automation agent — it contains all the instructions needed for a clean modernization pass.
```
