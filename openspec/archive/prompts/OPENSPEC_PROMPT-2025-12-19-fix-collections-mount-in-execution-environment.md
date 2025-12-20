# OpenSpec Change Prompt

## Context

The Packer ansible-navigator plugin successfully installs Ansible collections to `/home/solomong/.packer.d/ansible_collections_cache`, but when Ansible runs inside the execution environment container, it cannot find collection roles. Analysis of the error shows:

1. Collection `integration.common_tasks` installs successfully to `/home/solomong/.packer.d/ansible_collections_cache/ansible_collections/integration/common_tasks`
2. When running role `integration.common_tasks.etc_profile_d`, Ansible searches in `/tmp/roles:/home/solomong/.packer.d/ansible_roles_cache:/tmp`
3. The collections cache directory is **not** in ansible's search path
4. There's a deprecation warning about `ANSIBLE_COLLECTIONS_PATHS` (plural) vs `ANSIBLE_COLLECTIONS_PATH` (singular)

## Goal

Fix the packer-plugin-ansible-navigator to properly expose Ansible collections installed in `~/.packer.d/ansible_collections_cache` to the execution environment container, and fix the deprecation warning.

## Scope

**In scope:**
- Mount the collections cache directory as a volume in the execution environment
- Set `ANSIBLE_COLLECTIONS_PATH` (singular, not plural) environment variable to point to the mounted path
- Configure volume mounts to make host collections accessible inside the container
- Handle both EE-enabled and EE-disabled scenarios
- Use the correct singular form of the environment variable to fix deprecation warning

**Out of scope:**
- Changes to how collections are installed
- Changes to the roles cache directory (separate concern)
- Modifying user HCL configuration requirements

## Desired Behavior

When a user configures the plugin like this:

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Testing Navigator roles" 
    target = "integration.common_tasks.etc_profile_d"
  }
  requirements_file = "./requirements.yml"
  
  navigator_config {
    execution_environment {
      enabled = true
      image = "packer-ansible-ee:2.16"
    }
  }
}
```

The plugin should:

1. Install collections from `requirements_file` to `~/.packer.d/ansible_collections_cache/ansible_collections/`
2. Mount `~/.packer.d/ansible_collections_cache/ansible_collections` as a volume in the EE container
3. Set `ANSIBLE_COLLECTIONS_PATH` (singular) inside the container to point to the mounted collections
4. Ansible inside the EE container can find and execute collection roles like `integration.common_tasks.etc_profile_d`

## Constraints & Assumptions

- Assumption: Collections cache location is `~/.packer.d/ansible_collections_cache/ansible_collections` (based on plugin code)
- Assumption: The plugin generates `ansible-navigator.yml` configuration that controls volume mounts
- Assumption: Volume mounts need to be added to `navigator_config.execution_environment.volume_mounts` section
- Constraint: Must work with Docker and potentially Podman container engines
- Constraint: Must not break existing functionality when execution environment is disabled
- Constraint: Use `ANSIBLE_COLLECTIONS_PATH` (singular) not `ANSIBLE_COLLECTIONS_PATHS` (plural) to avoid deprecation warning

## Acceptance Criteria

- [ ] Collections installed to `~/.packer.d/ansible_collections_cache` are accessible inside the EE container
- [ ] Ansible can resolve collection roles like `integration.common_tasks.etc_profile_d`
- [ ] Volume mount is automatically added without requiring user HCL changes
- [ ] `ANSIBLE_COLLECTIONS_PATH` (singular) is set correctly in the EE container environment
- [ ] Deprecation warning about `ANSIBLE_COLLECTIONS_PATHS` is eliminated
- [ ] No regressions when EE is disabled
- [ ] Local test build succeeds with `describe` test
