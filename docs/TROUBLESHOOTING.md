# Troubleshooting

## Configuration errors

### Error: "at least one `play` block must be defined"

Add one or more `play { ... }` blocks:

```hcl
provisioner "ansible-navigator" {
  play {
    target = "site.yml"
  }
}
```

### Error: command must be only the executable name or path (no arguments)

`command` must be an executable only (no embedded args like `run`).

✅ Correct:

```hcl
command = "~/.asdf/shims/ansible-navigator"
```

❌ Wrong:

```hcl
command = "~/.asdf/shims/ansible-navigator run"
```

If you need to pass flags, use provisioner fields or the `navigator_config` block.

## Version check hangs / timeouts

### Automatic shim resolution (v0.4.0+)

The plugin automatically detects and resolves version manager shims (asdf, rbenv, pyenv). In most cases, no manual configuration is needed.

**When automatic resolution works:**

```bash
# Your version manager is configured
which ansible-navigator
# → ~/.asdf/shims/ansible-navigator (a shim)

asdf which ansible-navigator  # or rbenv/pyenv which
# → /home/user/.asdf/installs/ansible-navigator/2.3.0/bin/ansible-navigator
```

The plugin detects the shim and automatically uses the real binary path.

**When manual configuration is needed:**

If automatic resolution fails (version manager not in PATH, or `which` command fails), you'll see a clear error message with solutions:

```hcl
provisioner "ansible-navigator" {
  # Option 1: Specify the full path directly
  command = "/home/user/.asdf/installs/ansible-navigator/2.3.0/bin/ansible-navigator"
  # Find your path with: asdf which ansible-navigator

  play { target = "site.yml" }
}
```

**Or add the directory to PATH:**

```hcl
provisioner "ansible-navigator" {
  # Option 2: Add directories to be prepended to PATH
  ansible_navigator_path = ["/home/user/.asdf/installs/ansible-navigator/2.3.0/bin"]

  play { target = "site.yml" }
}
```

### Other timeout causes

If timeouts occur for reasons other than shims (network delays, container image pulls):

```hcl
provisioner "ansible-navigator" {
  version_check_timeout = "120s"  # Increase timeout
  # skip_version_check = true      # Last resort only

  play { target = "site.yml" }
}
```

## Execution environment permissions (`/.ansible/tmp`)

If Ansible in the execution environment fails due to non-root temp dirs, set `ansible_cfg` (or rely on the provisioner defaults when `execution_environment` is set):

```hcl
provisioner "ansible-navigator" {
  execution_environment = "quay.io/ansible/creator-ee:latest"

  ansible_cfg = {
    defaults = {
      remote_tmp = "/tmp/.ansible/tmp"
      local_tmp  = "/tmp/.ansible-local"
    }
  }

  play { target = "site.yml" }
}
```

## Dependencies not installed

If roles/collections are missing, ensure `requirements_file` exists and includes the needed items:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  play { target = "site.yml" }
}
```

## Authentication failures

The plugin delegates all authentication to `ansible-galaxy` and git. If collections cannot be installed from private sources, check external authentication setup.

### Error: "HTTP Error 401: Unauthorized"

**Symptom**: `ansible-galaxy install` fails with:
```
ERROR! Unknown error when attempting to call Galaxy at 'https://automation-hub.example.com/api/galaxy/': HTTP Error 401: Unauthorized
```

**Root Cause**: Invalid or missing API token for Private Automation Hub or custom Galaxy server.

**Resolution**:

1. **Verify API token** is valid:
   - Log into your Automation Hub web interface
   - Go to User Preferences → API Token
   - Regenerate if necessary

2. **Check `ansible.cfg` configuration** in your Packer template:
   ```hcl
   provisioner "ansible-navigator" {
     ansible_cfg = <<-EOF
       [galaxy]
       server_list = automation_hub
       
       [galaxy_server.automation_hub]
       url = https://automation-hub.example.com/api/galaxy/
       token = ${var.automation_hub_token}
     EOF
     
     play { target = "site.yml" }
   }
   ```

3. **Ensure token variable is set** when running Packer:
   ```bash
   packer build -var "automation_hub_token=YOUR_TOKEN" template.pkr.hcl
   ```

See [GALAXY_AUTHENTICATION.md - Private Automation Hub](GALAXY_AUTHENTICATION.md#private-automation-hub--red-hat-automation-hub) for detailed setup instructions.

### Error: "Permission denied (publickey)"

**Symptom**: Git clone fails during `ansible-galaxy install` with:
```
fatal: Could not read from remote repository.
Permission denied (publickey).
```

**Root Cause**: SSH key not configured, not added to GitHub/GitLab, or ssh-agent not running.

**Resolution**:

1. **Verify SSH key is added to GitHub/GitLab**:
   ```bash
   ssh -T git@github.com
   # Should show: "Hi username! You've successfully authenticated..."
   ```

2. **Start ssh-agent and add key**:
   ```bash
   eval "$(ssh-agent -s)"
   ssh-add ~/.ssh/id_ed25519  # or your key path
   ```

3. **Verify key is loaded**:
   ```bash
   ssh-add -l
   # Should list your key fingerprint
   ```

4. **For CI/CD environments**, ensure ssh-agent is set up in your build script (see [GALAXY_AUTHENTICATION.md - CI/CD Integration](GALAXY_AUTHENTICATION.md#cicd-integration-patterns))

See [GALAXY_AUTHENTICATION.md - Private GitHub (SSH)](GALAXY_AUTHENTICATION.md#private-github-repositories-ssh) for complete setup instructions.

### Error: "fatal: Authentication failed"

**Symptom**: Git HTTPS clone fails during `ansible-galaxy install` with:
```
fatal: Authentication failed for 'https://github.com/myorg/ansible-collection.git/'
```

**Root Cause**: Git credential helper not configured or GitHub Personal Access Token invalid/expired.

**Resolution**:

1. **Configure git credential helper**:
   ```bash
   # Cache credentials for 1 hour (recommended for development)
   git config --global credential.helper cache
   git config --global credential.helper 'cache --timeout=3600'
   ```

2. **Or set environment variables**:
   ```bash
   export GIT_USERNAME="your-github-username"
   export GIT_PASSWORD="ghp_YourPersonalAccessToken"
   packer build template.pkr.hcl
   ```

3. **Verify token has correct permissions**:
   - GitHub: Settings → Developer settings → Personal access tokens
   - Required scope: `repo` (for private repositories)

4. **For CI/CD**, inject token via secrets:
   ```yaml
   # GitHub Actions example
   - name: Run Packer
     env:
       GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
     run: |
       git config --global credential.helper store
       echo "https://${GITHUB_TOKEN}@github.com" > ~/.git-credentials
       packer build template.pkr.hcl
   ```

See [GALAXY_AUTHENTICATION.md - Private GitHub (HTTPS)](GALAXY_AUTHENTICATION.md#private-github-repositories-https-with-tokens) for complete setup instructions.

### Error: "Could not find/install packages"

**Symptom**: `ansible-galaxy install` reports:
```
ERROR! Could not find/install packages: namespace.collection_name
```

**Root Cause**: Multiple possible causes - authentication issues if collections are from private sources, wrong collection name, network issues, or offline mode enabled.

**Resolution**:

1. **If collections are from private sources**, verify authentication is configured:
   - Private GitHub: Check SSH keys or git credential helpers
   - Private Automation Hub: Verify API token in `ansible.cfg`
   - Custom Galaxy server: Check `galaxy_args` configuration
   - See [GALAXY_AUTHENTICATION.md](GALAXY_AUTHENTICATION.md) for detailed authentication setup

2. **Verify collection name and source**:
   ```yaml
   # requirements.yml
   collections:
     - name: namespace.collection_name  # Check spelling
       source: automation_hub            # Ensure source is correct
   ```

3. **Check network connectivity**:
   ```bash
   # Test access to Galaxy server
   curl -I https://galaxy.ansible.com
   
   # Or Private Automation Hub
   curl -I https://automation-hub.example.com/api/galaxy/
   ```

4. **Verify offline mode is not enabled**:
   ```hcl
   provisioner "ansible-navigator" {
     # Make sure offline is NOT enabled
     # offline = true  # ← Remove this if present
     
     requirements_file = "./requirements.yml"
     play { target = "site.yml" }
   }
   ```

## Collections not found inside execution environment

### Error: "unable to find role" or "collection not found" when using EE

**Symptom**: Collections work fine with `execution_environment.enabled = false`, but fail when EE is enabled:
```
ERROR! the role 'namespace.collection.role_name' was not found
```

**Root Cause** (before v6.1.0): Collections installed on the host were not mounted into the execution environment container.

**Resolution**:

Starting with v6.1.0, collections are automatically mounted when using execution environments. No manual configuration needed:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  
  play {
    # Collections from requirements.yml are automatically mounted and available
    target = "namespace.collection.role_name"
  }
}
```

**What happens automatically:**

- Collections installed to `~/.packer.d/ansible_collections_cache/ansible_collections` are mounted read-only into the container
- `ANSIBLE_COLLECTIONS_PATH` is set inside the container to point to the mounted collections
- No manual volume configuration needed

**For v6.0.0 and earlier**, you had to manually configure volume mounts. If you're on an older version, upgrade to v6.1.0+ or use manual mounts:

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      # Manual mount (not needed in v6.1.0+)
      volume_mounts {
        src = "/home/user/.packer.d/ansible_collections_cache/ansible_collections"
        dest = "/tmp/.packer_ansible/collections"
        options = "ro"
      }
      environment_variables {
        set = {
          ANSIBLE_COLLECTIONS_PATH = "/tmp/.packer_ansible/collections"
        }
      }
    }
  }
}
```

## ansible-navigator version warnings

### "Settings file format version needs migration"

**Symptom**: ansible-navigator displays warnings about configuration version migration during Packer builds.

**Resolution**: Starting with v3.1.0, the plugin automatically generates Version 2 format configuration files that are immediately recognized by ansible-navigator 25.x+ without triggering migration prompts.

If you're using an older version of the plugin, upgrade to v3.1.0+ to eliminate these warnings. No configuration changes are required on your part - the plugin handles Version 2 format generation automatically.

---

[← Back to docs index](README.md)
