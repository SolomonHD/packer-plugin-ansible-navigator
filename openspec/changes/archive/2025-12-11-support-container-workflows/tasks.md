# Tasks: Support Container Workflows

## Phase 1: Configuration Updates

1. [x] **Add `ansible_proxy_bind_address` configuration**
   - Add field to `Config` struct in `provisioner/ansible-navigator/provisioner.go`
   - Default to `127.0.0.1`
   - Update `setupAdapter` to use this address in `net.Listen`

2. [x] **Add `ansible_proxy_host` configuration**
   - Add field to `Config` struct in `provisioner/ansible-navigator/provisioner.go`
   - Default to `127.0.0.1`
   - Update `createInventoryFile` to use this address for `ansible_host`

## Phase 2: Execution Environment Updates

3. [x] **Force unbuffered Python output**
   - Inject `PYTHONUNBUFFERED=1` into `ansible-navigator` environment

## Phase 3: Documentation

4. [x] **Update Troubleshooting Guide**
   - Add "WSL2 and Containers" section to `docs/TROUBLESHOOTING.md`
   - Explain how to configure the new options for WSL2
