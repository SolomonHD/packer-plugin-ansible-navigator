# GitHub Actions builds

The **Build plugin** workflow in `.github/workflows/build.yml` produces
downloadable development binaries. It runs for pull requests, pushes to `main`
and `release/**`, and manual runs. It can also be called by another workflow.
It does not publish a GitHub Release, create a tag, or change the plugin version.

The workflow reads Go from `.go-version`, reads the base plugin version from
`version/VERSION`, runs tests/vet and the SDK plugin check, and discovers the
plugin API from the native binary. Packages use the correct
`packer-plugin-ansible-navigator` name and contain both provisioners.

## Download a build

1. Open [Actions → Build plugin](https://github.com/SolomonHD/packer-plugin-ansible-navigator/actions/workflows/build.yml).
2. Choose a successful run for the commit you want. To request a build manually,
   choose **Run workflow** and select a branch. Manual dispatch is available once
   this workflow is on the default branch.
3. Under **Artifacts**, download `packer-plugin-ansible-navigator-<os>-<arch>`.
4. Extract the downloaded GitHub artifact. It contains a plugin ZIP,
   `SHA256SUMS`, and `build-info.json` with the exact source commit, version,
   platform, and hashes.

| Machine | Artifact platform |
| --- | --- |
| WSL2 or native Linux on Intel/AMD | `linux-amd64` |
| Native Linux on ARM64 | `linux-arm64` |
| Intel Mac | `darwin-amd64` |
| Apple Silicon Mac | `darwin-arm64` |
| Windows on Intel/AMD | `windows-amd64` |

The existing Linux ARM/386, Windows 386, FreeBSD/NetBSD/OpenBSD amd64/386/ARM,
and Solaris amd64 cross-builds are retained. A successful cross-build establishes
compilation and packaging, not Navigator/container-runtime acceptance on every OS.
Artifacts are retained for 30 days, subject to repository retention limits.

## Verify and install

On Linux/WSL, from the extracted artifact directory:

```bash
sha256sum --check SHA256SUMS
unzip packer-plugin-ansible-navigator_v9.0.0-dev_x5.0_linux_amd64.zip
chmod +x packer-plugin-ansible-navigator_v9.0.0-dev_x5.0_linux_amd64
export PACKER_PLUGIN_PATH="$PWD/packer-dev-plugins"
packer plugins install \
  --path ./packer-plugin-ansible-navigator_v9.0.0-dev_x5.0_linux_amd64 \
  github.com/solomonhd/ansible-navigator
packer plugins installed
```

Use the actual version/API/platform in your artifact's filename. On macOS use
`shasum -a 256 -c SHA256SUMS`. On Windows, verify the ZIP with
`Get-FileHash -Algorithm SHA256`, extract it, and pass the resulting `.exe` to
the same `packer plugins install --path` command.

Packer accepts only the `-dev` prerelease convention, so development packages use
the base version plus `-dev`; the source SHA belongs in `build-info.json`.
Use an isolated `PACKER_PLUGIN_PATH`, as above, and keep it set for subsequent
`packer validate` and `packer build` commands. An installed stable binary of the
same base version takes precedence in a shared plugin directory. Packer HCL
constraints use the base version; a constraint containing `-dev` is rejected.
This configuration selects the development build installed in the isolated path:

```hcl
packer {
  required_plugins {
    ansible-navigator = {
      source  = "github.com/solomonhd/ansible-navigator"
      version = "= 9.0.0"
    }
  }
}
```

Actions artifacts are installed with `--path`; `packer init` does not download
Actions artifacts. Automatic remote installation needs compatible assets on a
published release, which remains a separate workflow. No signing or notarization
is performed by the artifact build; checksums detect changed download contents.

## Build the same package locally

Go and Python 3 are required. From a Git checkout:

```bash
python3 scripts/build_plugin.py \
  --os linux --arch amd64 --output-dir dist/linux-amd64
```

Use a separate output directory for each platform. The helper discovers the API
with a native build when `--api-version` is omitted, cross-compiles with CGO
disabled, and packages the binary with its executable permission preserved.
The workflow passes the API from its verified native build. `source_dirty` in
the build metadata distinguishes local edited checkouts from committed builds.

See [Packer's loading specification](https://developer.hashicorp.com/packer/docs/plugins/creation/plugin-load-spec)
for naming/version rules and [GitHub's artifact documentation](https://docs.github.com/en/actions/how-tos/writing-workflows/choosing-what-your-workflow-does/storing-and-sharing-data-from-a-workflow)
for downloading and retention.
