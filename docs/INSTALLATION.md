# Installation Guide

This repository provides two Packer provisioners:

- `ansible-navigator` (SSH-based; runs ansible-navigator on the machine where Packer runs)
- `ansible-navigator-local` (on-target; runs ansible-navigator via the communicator on the build target)

## Requirements

- Packer with plugin protocol **x5** support
- `ansible-navigator` available where it will run:
  - for `ansible-navigator`: on your Packer host
  - for `ansible-navigator-local`: on the build target

## Install for consumers (recommended): `required_plugins` + `packer init`

```hcl
packer {
  required_plugins {
    ansible-navigator = {
      source  = "github.com/solomonhd/ansible-navigator"
      version = ">= 1.0.0"
    }
  }
}
```

Then:

```bash
packer init .
```

## Local / air-gapped / dev installs: `packer plugins install --path`

If you have a plugin binary built locally (or downloaded from a release artifact), install it with Packer’s plugin installer:

```bash
packer plugins install \
  --path ./dist/packer-plugin-ansible-navigator_<version>_x5.0_<os>_<arch> \
  github.com/solomonhd/ansible-navigator
```

Notes:

- Use the binary naming conventions expected by Packer plugins.
- Prefer release automation (GoReleaser) to produce the `dist/` binaries + `_SHA256SUM` files.
- Avoid copying binaries directly into `~/.packer.d/plugins`; use the installer so Packer manages checksums and layout.

## Verify install

```bash
packer plugins installed | grep ansible-navigator
```

---

[← Back to docs index](README.md)
