# Change: Add ansible.cfg Generation Feature

## Why

Users of the Ansible Navigator plugin face a recurring issue when using execution environments (containers): ansible-runner defaults to creating temporary directories at `/.ansible/tmp`, which fails when the container runs as a non-root user (Permission denied error).

Current workarounds are error-prone and require users to understand both Packer plugin configuration AND Ansible configuration:

- Manually creating an `ansible.cfg` file in their working directory
- Using `extra_arguments` to pass `-e ansible_remote_tmp=/tmp/.ansible/tmp`
- Setting environment variables that may not be properly inherited by containers

This creates a poor user experience and scatters Ansible configuration across multiple mechanisms instead of providing a unified, declarative configuration option within the Packer plugin.

## What Changes

Add a new `ansible_cfg` configuration option that:

1. **Accepts a map/object of Ansible configuration settings** organized by INI sections (e.g., `defaults`, `ssh_connection`)
2. **Automatically generates a temporary `ansible.cfg` file** before provisioning begins
3. **Passes the path to ansible-navigator via `ANSIBLE_CONFIG` environment variable**
4. **Provides sensible defaults** for execution environment use cases (fixes the non-root container issue)
5. **Cleans up the temporary file** after provisioning completes (success or failure)

This applies to both provisioners:

- **Remote provisioner (`ansible-navigator`)**: generates ansible.cfg on local machine, sets ANSIBLE_CONFIG in local environment
- **Local provisioner (`ansible-navigator-local`)**: generates ansible.cfg on local machine, uploads to target, sets ANSIBLE_CONFIG in remote shell command

### Configuration Example

```hcl
provisioner "ansible-navigator" {
  play {
    name = "Configure system"
    target = "integration.common_tasks.setup"
  }
  
  execution_environment = "containers.github.service.emory.edu/myorg/ee:latest"
  
  # NEW: ansible_cfg option
  ansible_cfg = {
    defaults = {
      remote_tmp      = "/tmp/.ansible/tmp"
      local_tmp       = "/tmp/.ansible-local"
      host_key_checking = "False"
      timeout         = "30"
    }
    ssh_connection = {
      pipelining      = "True"
      ssh_args        = "-o ControlMaster=auto -o ControlPersist=60s"
    }
  }
}
```

### Generated File

The plugin generates `/tmp/packer-ansible-cfg-<random>.ini`:

```ini
[defaults]
remote_tmp = /tmp/.ansible/tmp
local_tmp = /tmp/.ansible-local
host_key_checking = False
timeout = 30

[ssh_connection]
pipelining = True
ssh_args = -o ControlMaster=auto -o ControlPersist=60s
```

### Default Behavior When execution_environment Is Set

When `execution_environment` is configured and `ansible_cfg` is NOT explicitly set, the plugin SHALL apply these defaults automatically:

```hcl
ansible_cfg = {
  defaults = {
    remote_tmp = "/tmp/.ansible/tmp"
    local_tmp  = "/tmp/.ansible-local"
  }
}
```

This eliminates the "Permission denied: /.ansible" error for non-root container users without requiring manual configuration.

## Impact

### Affected specs

- `local-provisioner-capabilities` - New ansible.cfg generation requirement
- `remote-provisioner-capabilities` - New ansible.cfg generation requirement

### Affected code

- `provisioner/ansible-navigator/provisioner.go` - Config struct, generation logic, environment setup
- `provisioner/ansible-navigator-local/provisioner.go` - Config struct, generation logic, file upload, remote environment setup
- Both provisioners' HCL2 spec files (will need `make generate`)

### Backward Compatibility

This change is **100% backward compatible**:

- Existing configurations without `ansible_cfg` work unchanged
- Users with manual `ansible.cfg` files: their files take precedence (Ansible's normal search order)
- Users using `extra_arguments` or environment variables: those still work
- No breaking changes to existing configuration options

### Benefits

- **Fixes execution environment non-root user errors** with zero configuration
- **Eliminates scattered configuration** (consolidates Ansible settings into one place)
- **Improves user experience** (declarative, self-documenting configuration)
- **Reduces support burden** (fewer workarounds needed)
- **Enables advanced use cases** (users can configure any Ansible setting via HCL)
