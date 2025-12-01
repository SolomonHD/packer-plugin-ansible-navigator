# Agent Prompt: Upgrade Packer Plugin to Latest Go 1.x and Packer SDK

## Goal
Upgrade this Packer plugin repository to:
- The latest stable **Go 1.x** toolchain (Go 1.23+ as of 2024/2025).
- The latest **packer-plugin-sdk** (v0.7.x or newer).
- Ensure build artifacts match the **Packer ≥ 1.10 plugin protocol (x5)** naming standard.
- Modernize Go code using current Go 1.x best practices (generics, improved error handling, etc.).
- Ensure the project builds cleanly via **Go modules** and **GoReleaser v2**.
- Produce release assets suitable for `packer plugins install`.

---

## Context

### Current Project
This repository is a Packer plugin named **`packer-plugin-ansible-navigator`**.
It currently:
- Uses an invalid Go version (1.25.3 - which doesn't exist).
- Depends on `packer-plugin-sdk v0.6.4` (outdated).
- Has GoReleaser v2 config with x5 naming (already correct).

### Target Behavior
When the user runs:
```bash
packer plugins install github.com/solomonhd/ansible-navigator vX.Y.Z
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

   * Update `go.mod` to use latest stable Go 1.x version (e.g., `go 1.23` or `go 1.24`).
   * Run `go mod tidy` and `go get -u ./...` to update dependencies.
   * Update `.go-version` file to match.
   * Fix any deprecated syntax or API usage.

2. **Update SDK**

   * Upgrade to latest `github.com/hashicorp/packer-plugin-sdk` (check for v0.7.x+).
   * Update imports if SDK structure changed.
   * Verify plugin registration pattern matches SDK version:

     ```go
     pps := plugin.NewSet()
     pps.RegisterProvisioner("ansible-navigator", new(ansible.Provisioner))
     pps.SetVersion(version.PluginVersion)
     err := pps.Run()
     ```

3. **Modernize Go Code**

   * Apply current Go best practices (proper error handling, context usage).
   * Use generics where beneficial (Go 1.18+).
   * Ensure proper use of `context.Context` throughout.
   * Apply linter recommendations.

4. **Verify GoReleaser Config**

   * Confirm `version: 2` is set (✓ already correct).
   * Verify artifact naming uses x5 protocol (✓ already correct):

     ```yaml
     name_template: "{{ .ProjectName }}_v{{ .Version }}_x5_{{ .Os }}_{{ .Arch }}.zip"
     ```

5. **Validate Release**

   * Run `goreleaser release --snapshot --clean`.
   * Confirm correct filenames and checksum output.
   * Verify installation via:

     ```bash
     packer plugins install github.com/solomonhd/ansible-navigator vX.Y.Z
     ```

6. **Update Documentation**

   * Update README.md to state minimum requirements:

     * Go 1.23+ (or current stable version)
     * Packer ≥ 1.10
     * Current Packer Plugin SDK version
   * Verify `packer plugins install` instructions are accurate.

---

## Deliverables

* ✅ Updated `go.mod` to valid Go 1.x version (1.23 or 1.24).
* ✅ Updated `.go-version` to match.
* ✅ Upgraded to latest `packer-plugin-sdk`.
* ✅ Verified `.goreleaser.yaml` uses x5 protocol naming (already correct).
* ✅ Verified snapshot build with `goreleaser release --snapshot --clean`.
* ✅ Updated documentation reflecting current requirements.

---

## Notes

* **Go "2.x" does not exist** - Go evolves within the 1.x version line (currently 1.23-1.24).
* The project currently has an invalid Go version (1.25.3) that needs correction.
* GoReleaser config already uses correct x5 naming - no changes needed there.
* Focus on upgrading SDK and fixing Go version to valid, current release.
* Maintain backward compatibility where possible.
* Ensure reproducible builds for release automation.

---

**End of Prompt**
