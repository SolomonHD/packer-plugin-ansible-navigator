# 🐛 Troubleshooting Guide

Common issues, error messages, and solutions for the Packer Plugin Ansible Navigator.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Common Error Messages](#common-error-messages)
- [Installation Issues](#installation-issues)
- [Configuration Problems](#configuration-problems)
- [Execution Environment Issues](#execution-environment-issues)
- [Connection Problems](#connection-problems)
- [Performance Issues](#performance-issues)
- [Debugging Techniques](#debugging-techniques)
- [FAQ](#faq)

## Quick Diagnostics

Run this diagnostic checklist first:

```bash
# 1. Check Packer version
packer version
# Should be >= 1.7.0

# 2. Check plugin is installed
packer plugins installed | grep ansible-navigator
# Should show: github.com/solomonhd/packer-plugin-ansible-navigator

# 3. Check ansible-navigator is available
ansible-navigator --version
# Should show version information

# 4. Check Docker/Podman is running (if using containerized EE)
docker info || podman info
# Should show system information

# 5. Validate your template
packer validate your-template.pkr.hcl
# Should show: The configuration is valid.
```

## Version Check Issues

### Version Check Hangs or Times Out

**Symptoms:**

- Packer hangs during ansible-navigator version check
- Error: "ansible-navigator version check timed out after 60s"
- Build stalls at the beginning with no visible progress

**Common Causes:**

- ansible-navigator not installed or not in PATH
- ansible-navigator installed via asdf with shim issues
- Container runtime (Docker/Podman) not running or slow
- Network delays downloading execution environment images

**Solutions:**

1. **Verify ansible-navigator is installed and accessible:**

```bash
# Check if ansible-navigator is in PATH
which ansible-navigator

# Test version check manually
ansible-navigator --version

# Check execution time
time ansible-navigator --version
```

2. **For asdf users - specify full path to shim:**

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  command = "~/.asdf/shims/ansible-navigator run"
  
  # or add to PATH
  ansible_navigator_path = ["~/.asdf/shims"]
}
```

3. **Increase timeout for slow environments:**

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  version_check_timeout = "120s"  # 2 minutes instead of default 60s
}
```

4. **Skip version check if you're sure it's installed:**

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  skip_version_check = true
}
```

5. **Pre-pull execution environment image:**

```bash
# Pull the default or specified execution environment ahead of time
docker pull quay.io/ansible/creator-ee:latest

# Or for custom image
docker pull myregistry.io/my-ee:v1.0
```

### asdf-Specific Issues

**Problem:** ansible-navigator installed via asdf doesn't work in Packer subprocess.

**Why:** asdf uses shim scripts that may not have the correct environment when called from Packer's subprocess context.

**Solutions:**

**Option 1: Use full path to asdf shim**

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  command = "/home/user/.asdf/shims/ansible-navigator run"
  version_check_timeout = "90s"  # asdf may be slower
}
```

**Option 2: Add asdf shims to PATH**

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  ansible_navigator_path = ["~/.asdf/shims"]
}
```

**Option 3: Use pip/pipx instead of asdf**

```bash
# Install via pipx (recommended for CLI tools)
pipx install ansible-navigator

# Or via pip
pip install --user ansible-navigator
```

**Verify asdf setup:**

```bash
# Check asdf version
asdf current ansible-navigator

# Test shim directly
~/.asdf/shims/ansible-navigator --version

# Check shim path
ls -l ~/.asdf/shims/ansible-navigator
```

### Interaction with skip_version_check

The `version_check_timeout` option is only used when `skip_version_check` is `false` (the default). If you set `skip_version_check = true`, the timeout is ignored and no version check is performed.

**HCL2**

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  
  # These work together
  skip_version_check = false  # default - perform version check
  version_check_timeout = "90s"  # timeout for the check
}

# OR skip version check entirely
# provisioner "ansible-navigator" {
#   playbook_file = "site.yml"
#   skip_version_check = true  # no version check, timeout ignored
# }
```

**JSON**

```json
{
  "type": "ansible-navigator",
  "playbook_file": "site.yml",
  "skip_version_check": false,
  "version_check_timeout": "90s"
}
```

## Common Error Messages

### Error: "You may specify only one of `playbook_file` or `play` blocks"

**Problem**: Both `playbook_file` and one or more `play` blocks are specified in configuration.

**Solution**: Remove one of them. They are mutually exclusive.

❌ **Wrong:**

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"

  play { # CONFLICT!
    name   = "My Play"
    target = "namespace.collection.play"
  }
}
```

✅ **Correct:**

```hcl
# Option A: Use playbook file
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}

# Option B: Use collection plays (repeated play blocks)
provisioner "ansible-navigator" {
  play {
    name   = "My Play"
    target = "namespace.collection.play"
    extra_vars = {
      environment = "production"
    }
  }
}
```

### Error: "Either `playbook_file` or `play` blocks must be defined"

**Problem**: Neither `playbook_file` nor `plays` is specified.

**Solution**: Add one of them to your configuration.

```hcl
provisioner "ansible-navigator" {
  # Add either:
  playbook_file = "site.yml"

  # OR one or more play blocks:
  play {
    name   = "Configure System"
    target = "namespace.collection.play"
  }
}
```

### Error: "ansible-navigator not found in PATH"

**Problem**: ansible-navigator is not installed or not in PATH.

**Solutions:**

1. Install ansible-navigator:

```bash
pip install ansible-navigator
# or
pipx install ansible-navigator
```

2. Add to PATH if installed locally:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

3. Specify full path in configuration:

```hcl
provisioner "ansible-navigator" {
  command = "/usr/local/bin/ansible-navigator run"
}
```

### Error: "Play 'namespace.collection.play_name' failed (exit code 2)"

**Problem**: A specific play failed during execution.

**Solutions:**

1. Check the play exists in the collection:

```bash
ansible-doc -l | grep namespace.collection
```

2. Verify collection is installed:

```bash
ansible-galaxy collection list
```

3. Enable verbose logging for details:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = ["-vvv"]
  
  # Enable JSON logging for detailed output
  navigator_mode = "json"
  structured_logging = true
  verbose_task_output = true
}
```

### Error: "Failed to pull execution environment image"

**Problem**: Cannot download the container image for ansible-navigator.

**Solutions:**

1. Check Docker/Podman is running:

```bash
docker info
# or
podman info
```

2. Pull image manually:

```bash
docker pull quay.io/ansible/creator-ee:latest
```

3. Use local image:

```hcl
provisioner "ansible-navigator" {
  execution_environment = "localhost:5000/my-ee:latest"
}
```

4. Disable image pulling:

```bash
export ANSIBLE_NAVIGATOR_PULL_POLICY=never
```

## Installation Issues

### Plugin Not Found After Installation

**Symptoms:**

- `packer plugins installed` doesn't show the plugin
- Template validation fails with "unknown provisioner"

**Solutions:**

1. Check plugin location:

```bash
# Linux/macOS
ls -la ~/.packer.d/plugins/packer-plugin-ansible-navigator

# Windows
dir %APPDATA%\packer.d\plugins
```

2. Verify file permissions:

```bash
chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator
```

3. Check plugin name format:

- Must start with `packer-plugin-`
- Binary name: `packer-plugin-ansible-navigator` (exactly)

4. Clear plugin cache:

```bash
rm -rf ~/.config/packer/plugins/
packer init your-template.pkr.hcl
```

### Version Conflicts

**Symptoms:**

- "Version constraint not satisfied"
- Multiple versions installed

**Solution:**

```bash
# Remove all versions
rm ~/.packer.d/plugins/packer-plugin-ansible-navigator*

# Reinstall specific version
packer init -upgrade your-template.pkr.hcl
```

### Go Version Issues (Building from Source)

**Symptoms:**

- Build fails with Go errors
- "undefined" errors during compilation

**Solutions:**

1. Update Go version:

```bash
go version  # Should be >= 1.25.3
```

2. Clean module cache:

```bash
go clean -modcache
go mod download
```

3. Use correct build command:

```bash
go build -o packer-plugin-ansible-navigator
```

## Configuration Problems

### Collections Not Installing

**Problem**: Collections specified in configuration aren't being installed.

**Solutions:**

1. Check Galaxy server access:

```bash
ansible-galaxy collection list
```

2. Use explicit Galaxy configuration:

```hcl
provisioner "ansible-navigator" {
  ansible_env_vars = [
    "ANSIBLE_GALAXY_SERVER=https://galaxy.ansible.com",
    "ANSIBLE_GALAXY_TOKEN=${var.galaxy_token}"
  ]
}
```

3. Pre-install collections:

```bash
ansible-galaxy collection install -r requirements.yml
```

4. Use offline mode with cached collections:

```hcl
provisioner "ansible-navigator" {
  collections_offline = true
  collections_cache_dir = "/path/to/cached/collections"
}
```

### Requirements File Not Found

**Problem**: `requirements_file` specified but file doesn't exist.

**Solution**: Verify file path is relative to Packer working directory:

```hcl
provisioner "ansible-navigator" {
  # Path relative to where packer is run
  requirements_file = "./ansible/requirements.yml"
}
```

### Invalid Configuration Values

**Problem**: Configuration validation fails.

**Solution**: Check value types and options:

```hcl
provisioner "ansible-navigator" {
  # navigator_mode must be one of: stdout, interactive, json
  navigator_mode = "json"  # ✅
  # navigator_mode = "xml"  # ❌ Invalid
  
  # Boolean values
  keep_going = true  # ✅
  # keep_going = "true"  # ❌ Wrong type
  
  # Repeatable blocks (no array syntax for play)
  play {
    name   = "Play 1"
    target = "namespace.collection.play_one"
  }

  play {
    name   = "Play 2"
    target = "namespace.collection.play_two"
  }
}
```

## Execution Environment Issues

### Container Runtime Not Found

**Problem**: Docker or Podman not available.

**Solutions:**

1. Install Docker:

```bash
# Ubuntu/Debian
sudo apt-get install docker.io

# macOS
brew install docker

# Start Docker
sudo systemctl start docker
```

2. Use Podman instead:

```bash
export ANSIBLE_NAVIGATOR_CONTAINER_ENGINE=podman
```

3. Use ansible-playbook without containerization:

```hcl
provisioner "ansible" {  # Use regular ansible provisioner
  playbook_file = "site.yml"
}
```

### Execution Environment Image Issues

**Problem**: Custom execution environment not working.

**Solutions:**

1. Build custom EE:

```dockerfile
FROM quay.io/ansible/creator-ee:latest
RUN pip install custom-package
```

2. Push to registry:

```bash
docker build -t myregistry/my-ee:latest .
docker push myregistry/my-ee:latest
```

3. Use in configuration:

```hcl
provisioner "ansible-navigator" {
  execution_environment = "myregistry/my-ee:latest"
}
```

### Permission Denied in Container

**Problem**: Ansible can't write files in container.

**Solution**: Set proper work directory:

```hcl
provisioner "ansible-navigator" {
  work_dir = "/tmp/ansible-work"
  
  ansible_env_vars = [
    "ANSIBLE_LOCAL_TEMP=/tmp/ansible-tmp",
    "ANSIBLE_REMOTE_TEMP=/tmp/ansible-remote"
  ]
}
```

## Connection Problems

### SSH Connection Failed

**Symptoms:**

- "Host key verification failed"
- "Permission denied (publickey)"
- "Connection refused"

**Solutions:**

1. Disable host key checking:

```hcl
provisioner "ansible-navigator" {
  ansible_env_vars = [
    "ANSIBLE_HOST_KEY_CHECKING=False"
  ]
}
```

2. Specify SSH key:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = [
    "--private-key", "/path/to/key.pem"
  ]
}
```

3. Debug SSH connection:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = ["-vvvv"]
  
  ansible_env_vars = [
    "ANSIBLE_SSH_ARGS='-o StrictHostKeyChecking=no -v'"
  ]
}
```

### WinRM Connection Failed

**Problem**: Windows connection issues.

**Solutions:**

1. Configure WinRM properly:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = [
    "--extra-vars", "ansible_winrm_server_cert_validation=ignore",
    "--extra-vars", "ansible_winrm_transport=ntlm"
  ]
}
```

2. Enable basic auth (for testing only):

```powershell
# On Windows target
winrm set winrm/config/service/auth '@{Basic="true"}'
winrm set winrm/config/service '@{AllowUnencrypted="true"}'
```

### Proxy Issues

**Problem**: Can't connect through corporate proxy.

**Solution:**

```hcl
provisioner "ansible-navigator" {
  use_proxy = true
  
  ansible_env_vars = [
    "HTTP_PROXY=${var.proxy_url}",
    "HTTPS_PROXY=${var.proxy_url}",
    "NO_PROXY=localhost,127.0.0.1,internal.domain"
  ]
}
```

## Performance Issues

### Slow Collection Installation

**Problem**: Collections take too long to download.

**Solutions:**

1. Use collection cache:

```hcl
provisioner "ansible-navigator" {
  collections_cache_dir = "~/.ansible/collections_cache"
}
```

2. Pre-download collections:

```bash
ansible-galaxy collection download -r requirements.yml -p ./collections
```

3. Use private Galaxy server:

```hcl
provisioner "ansible-navigator" {
  ansible_env_vars = [
    "ANSIBLE_GALAXY_SERVER=https://galaxy.internal.com"
  ]
}
```

### Playbook Execution Slow

**Problem**: Playbooks run slower than expected.

**Solutions:**

1. Increase parallelism:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = [
    "--forks", "50"
  ]
}
```

2. Enable pipelining:

```hcl
provisioner "ansible-navigator" {
  ansible_env_vars = [
    "ANSIBLE_PIPELINING=True"
  ]
}
```

3. Use fact caching:

```hcl
provisioner "ansible-navigator" {
  ansible_env_vars = [
    "ANSIBLE_GATHERING=smart",
    "ANSIBLE_CACHE_PLUGIN=jsonfile",
    "ANSIBLE_CACHE_PLUGIN_CONNECTION=/tmp/facts"
  ]
}
```

## Debugging Techniques

### Enable Maximum Verbosity

```hcl
provisioner "ansible-navigator" {
  extra_arguments = ["-vvvv"]
  
  # Enable debug environment variables
  ansible_env_vars = [
    "ANSIBLE_DEBUG=True",
    "ANSIBLE_VERBOSITY=4"
  ]
}
```

### Use Interactive Mode

```hcl
provisioner "ansible-navigator" {
  navigator_mode = "interactive"
  # This will pause and show the TUI
}
```

### Capture JSON Events

```hcl
provisioner "ansible-navigator" {
  navigator_mode = "json"
  structured_logging = true
  log_output_path = "./debug-output.json"
  verbose_task_output = true
}
```

### Test with Docker Locally

```bash
# Test your playbook in the same EE
docker run -it --rm \
  -v $(pwd):/workspace \
  quay.io/ansible/creator-ee:latest \
  ansible-navigator run site.yml
```

### Check Plugin Logs

```bash
# Set Packer log level
export PACKER_LOG=1
export PACKER_LOG_PATH=packer.log

# Run Packer
packer build template.pkr.hcl

# Check the log
cat packer.log | grep ansible-navigator
```

## FAQ

### Q: Can I use this plugin without Docker/Podman?

**A:** The plugin requires ansible-navigator, which typically uses container runtimes. However, you can:

1. Use the standard `ansible` provisioner instead
2. Configure ansible-navigator to work without containers (advanced)

### Q: How do I pass vault passwords?

**A:** Use vault password files:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = [
    "--vault-password-file", ".vault-pass"
  ]
  
  # Or via environment
  ansible_env_vars = [
    "ANSIBLE_VAULT_PASSWORD_FILE=.vault-pass"
  ]
}
```

### Q: Can I use local collections under development?

**A:** Yes, use local paths:

```hcl
provisioner "ansible-navigator" {
  collections = [
    "mycompany.mycollection@../path/to/collection"
  ]
  collections_force_update = true
}
```

### Q: How do I use private container registries?

**A:** Login before running Packer:

```bash
# Docker
docker login myregistry.com

# Podman
podman login myregistry.com

# Then use in config
provisioner "ansible-navigator" {
  execution_environment = "myregistry.com/my-ee:latest"
}
```

### Q: Why is my playbook not found?

**A:** Check these:

1. Path is relative to Packer working directory
2. File exists and is readable
3. YAML syntax is valid
4. If using collections, the play name is correct

### Q: How do I debug connection issues?

**A:** Enable SSH debugging:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = ["-vvvv"]
  
  ansible_env_vars = [
    "ANSIBLE_SSH_ARGS='-vvv -o StrictHostKeyChecking=no'",
    "ANSIBLE_SSH_COMMON_ARGS='-o StrictHostKeyChecking=no'",
    "ANSIBLE_CONNECTION=ssh"
  ]
}
```

### Q: Can I use custom Python interpreters?

**A:** Yes, specify in extra variables:

```hcl
provisioner "ansible-navigator" {
  extra_arguments = [
    "--extra-vars", "ansible_python_interpreter=/usr/bin/python3"
  ]
}
```

## Getting Additional Help

If your issue isn't covered here:

1. **Check the logs** with `PACKER_LOG=1`
2. **Search existing issues**: [GitHub Issues](https://github.com/solomonhd/packer-plugin-ansible-navigator/issues)
3. **Ask for help**: [GitHub Discussions](https://github.com/solomonhd/packer-plugin-ansible-navigator/discussions)
4. **Report a bug** with:
   - Packer version
   - Plugin version
   - Minimal reproduction template
   - Full error output
   - Debug logs

---

[← Examples Gallery](EXAMPLES.md) | [Back to README →](../README.md)
