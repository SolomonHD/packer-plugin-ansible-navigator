# Galaxy Authentication Guide

This guide explains how to install Ansible collections and roles from **private sources** using the Packer Ansible Navigator plugin. It covers authentication for private GitHub repositories, Private Automation Hub, Red Hat Automation Hub, and custom Galaxy servers.

## Table of Contents

- [Understanding the Boundary](#understanding-the-boundary)
- [Public Galaxy (Default)](#public-galaxy-default)
- [Private GitHub Repositories (SSH)](#private-github-repositories-ssh)
- [Private GitHub Repositories (HTTPS with Tokens)](#private-github-repositories-https-with-tokens)
- [Private Automation Hub / Red Hat Automation Hub](#private-automation-hub--red-hat-automation-hub)
- [Custom Galaxy Servers](#custom-galaxy-servers)
- [CI/CD Integration Patterns](#cicd-integration-patterns)

## Understanding the Boundary

The Packer Ansible Navigator plugin **delegates** all authentication to external tools:

### What the Plugin Handles

- Invoking `ansible-galaxy install` with your [`requirements_file`](provisioner/ansible-navigator/galaxy.go)
- Passing additional flags via [`galaxy_args`](../openspec/specs/remote-provisioner-capabilities/spec.md#L77)
- Managing install paths and temporary directories

### What Must Be Configured Externally

- **Git credentials**: SSH keys, credential helpers, or environment variables for git authentication
- **Galaxy server configuration**: API keys and server definitions in [`ansible.cfg`](../openspec/specs/remote-provisioner-capabilities/spec.md#L100)
- **Certificate trust**: System CA certificates or `ansible.cfg` SSL verification settings

**Key Principle**: The plugin never intercepts or modifies authentication. If `ansible-galaxy` can authenticate on your system, it will work in the plugin.

---

## Public Galaxy (Default)

Collections from [galaxy.ansible.com](https://galaxy.ansible.com) work **out of the box** with no additional configuration.

### Example: Public Collections

**requirements.yml**:
```yaml
---
collections:
  - name: community.general
    version: ">=8.0.0"
  - name: ansible.posix
```

**Packer HCL**:
```hcl
provisioner "ansible-navigator" {
  playbook_file    = "playbook.yml"
  requirements_file = "requirements.yml"
}
```

No authentication setup required. The plugin will install collections from public Galaxy.

---

## Private GitHub Repositories (SSH)

Use SSH authentication when your collections are hosted in private GitHub or GitLab repositories accessed via SSH URLs.

### Prerequisites

1. **Generate SSH key** (if you don't have one):
   ```bash
   ssh-keygen -t ed25519 -C "your_email@example.com"
   ```

2. **Add public key to GitHub**:
   - Go to GitHub Settings → SSH and GPG keys → New SSH key
   - Paste contents of `~/.ssh/id_ed25519.pub`

3. **Add private key to ssh-agent**:
   ```bash
   eval "$(ssh-agent -s)"
   ssh-add ~/.ssh/id_ed25519
   ```

### Example: Private Repository via SSH

**requirements.yml**:
```yaml
---
collections:
  - name: my_namespace.my_collection
    type: git
    source: git+ssh://git@github.com/myorg/ansible-collection-mycollection.git
    version: main
```

**Packer HCL**:
```hcl
provisioner "ansible-navigator" {
  playbook_file     = "playbook.yml"
  requirements_file = "requirements.yml"
}
```

### Troubleshooting SSH Authentication

If you see **"Permission denied (publickey)"**:

1. Verify SSH key is added to GitHub:
   ```bash
   ssh -T git@github.com
   # Should show: "Hi username! You've successfully authenticated..."
   ```

2. Ensure ssh-agent is running and key is loaded:
   ```bash
   ssh-add -l
   # Should list your key
   ```

3. If key is missing from agent:
   ```bash
   ssh-add ~/.ssh/id_ed25519
   ```

### SSH in CI/CD Environments

See [CI/CD Integration Patterns](#cicd-integration-patterns) for automated ssh-agent setup.

---

## Private GitHub Repositories (HTTPS with Tokens)

Use HTTPS authentication with **Personal Access Tokens (PAT)** or **GitHub App tokens** when SSH is not available or preferred.

### Prerequisites

1. **Create Personal Access Token**:
   - GitHub: Settings → Developer settings → Personal access tokens → Generate new token
   - Required scopes: `repo` (for private repositories)

2. **Configure git credential helper** (choose one):

   **Option A: Credential cache (temporary, recommended for development)**:
   ```bash
   git config --global credential.helper cache
   git config --global credential.helper 'cache --timeout=3600'
   ```

   **Option B: Credential store (persistent, stored in plaintext)**:
   ```bash
   git config --global credential.helper store
   ```

   **Option C: Platform-specific (macOS keychain, Windows Credential Manager)**:
   ```bash
   # macOS
   git config --global credential.helper osxkeychain
   
   # Linux (gnome-keyring)
   git config --global credential.helper /usr/share/git/credential/gnome-keyring/git-credential-gnome-keyring
   ```

3. **Set environment variables** (alternative to credential helper):
   ```bash
   export GIT_USERNAME="your-github-username"
   export GIT_PASSWORD="ghp_YourPersonalAccessToken"
   ```

   Or use `GIT_ASKPASS` script:
   ```bash
   #!/bin/bash
   echo "$GITHUB_TOKEN"
   ```
   ```bash
   chmod +x git-askpass.sh
   export GIT_ASKPASS="$(pwd)/git-askpass.sh"
   export GITHUB_TOKEN="ghp_YourPersonalAccessToken"
   ```

### Example: Private Repository via HTTPS

**requirements.yml**:
```yaml
---
collections:
  - name: my_namespace.my_collection
    type: git
    source: https://github.com/myorg/ansible-collection-mycollection.git
    version: main
```

**Packer HCL**:
```hcl
provisioner "ansible-navigator" {
  playbook_file     = "playbook.yml"
  requirements_file = "requirements.yml"
}
```

### Security Warning

**❌ NEVER commit tokens to version control:**

```hcl
# BAD - Token embedded in template
provisioner "ansible-navigator" {
  # DON'T DO THIS!
  environment_vars = [
    "GIT_PASSWORD=ghp_hardcodedtoken123"
  ]
}
```

**✅ Use environment variables or CI/CD secrets:**

```bash
# Good - Token from environment
export GIT_PASSWORD="${GITHUB_TOKEN}"
packer build template.pkr.hcl
```

See [CI/CD Integration Patterns](#cicd-integration-patterns) for secure token injection.

---

## Private Automation Hub / Red Hat Automation Hub

Private Automation Hub and Red Hat Automation Hub require API key authentication configured in `ansible.cfg`.

### Configuration via ansible.cfg

The plugin supports providing custom `ansible.cfg` content via the [`ansible_cfg`](../openspec/specs/remote-provisioner-capabilities/spec.md#L100) configuration block.

**Packer HCL**:
```hcl
provisioner "ansible-navigator" {
  playbook_file     = "playbook.yml"
  requirements_file = "requirements.yml"
  
  ansible_cfg = <<-EOF
    [galaxy]
    server_list = automation_hub
    
    [galaxy_server.automation_hub]
    url = https://console.redhat.com/api/automation-hub/
    auth_url = https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token
    token = ${var.automation_hub_token}
  EOF
}
```

**requirements.yml**:
```yaml
---
collections:
  - name: redhat.insights
    source: automation_hub
  - name: ansible.posix
    source: automation_hub
```

### Private Automation Hub (Self-Hosted)

For self-hosted Private Automation Hub:

```hcl
variable "pah_token" {
  type      = string
  sensitive = true
}

provisioner "ansible-navigator" {
  playbook_file     = "playbook.yml"
  requirements_file = "requirements.yml"
  
  ansible_cfg = <<-EOF
    [galaxy]
    server_list = company_hub
    
    [galaxy_server.company_hub]
    url = https://automation-hub.company.com/api/galaxy/
    token = ${var.pah_token}
    
    # For self-signed certificates (development only)
    # verify_ssl = false
  EOF
}
```

### Handling Self-Signed Certificates

**Option 1: Add CA certificate to system trust store** (recommended):
```bash
# Ubuntu/Debian
sudo cp company-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# RHEL/CentOS
sudo cp company-ca.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust
```

**Option 2: Disable SSL verification** (development only):
```ini
[galaxy_server.company_hub]
url = https://automation-hub.company.com/api/galaxy/
token = ${var.pah_token}
verify_ssl = false
```

---

## Custom Galaxy Servers

Use [`galaxy_args`](../openspec/specs/remote-provisioner-capabilities/spec.md#L77) to pass server URLs, API keys, and other flags directly to `ansible-galaxy install`.

### Example: Custom Galaxy Server with API Key

**Packer HCL**:
```hcl
variable "galaxy_api_key" {
  type      = string
  sensitive = true
}

provisioner "ansible-navigator" {
  playbook_file     = "playbook.yml"
  requirements_file = "requirements.yml"
  
  galaxy_args = [
    "--server", "https://galaxy.mycompany.com",
    "--api-key", var.galaxy_api_key
  ]
}
```

### Example: Ignoring SSL Certificates (Development)

```hcl
provisioner "ansible-navigator" {
  playbook_file     = "playbook.yml"
  requirements_file = "requirements.yml"
  
  galaxy_args = [
    "--server", "https://galaxy.mycompany.com",
    "--api-key", var.galaxy_api_key,
    "--ignore-certs"
  ]
}
```

### Multiple Galaxy Servers

Combine [`ansible.cfg`](../openspec/specs/remote-provisioner-capabilities/spec.md#L100) with server definitions and use `requirements.yml` `source` field:

**Packer HCL**:
```hcl
provisioner "ansible-navigator" {
  playbook_file     = "playbook.yml"
  requirements_file = "requirements.yml"
  
  ansible_cfg = <<-EOF
    [galaxy]
    server_list = public_galaxy, company_hub
    
    [galaxy_server.public_galaxy]
    url = https://galaxy.ansible.com
    
    [galaxy_server.company_hub]
    url = https://galaxy.mycompany.com
    token = ${var.company_galaxy_token}
  EOF
}
```

**requirements.yml**:
```yaml
---
collections:
  - name: community.general
    source: public_galaxy  # From public Galaxy
  - name: mycompany.internal
    source: company_hub    # From private server
```

---

## CI/CD Integration Patterns

Securely handle authentication credentials in automated build pipelines.

### GitHub Actions

#### SSH Key Setup

```yaml
name: Packer Build
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Configure SSH
        env:
          SSH_PRIVATE_KEY: ${{ secrets.SSH_PRIVATE_KEY }}
        run: |
          mkdir -p ~/.ssh
          echo "$SSH_PRIVATE_KEY" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan github.com >> ~/.ssh/known_hosts
          eval "$(ssh-agent -s)"
          ssh-add ~/.ssh/id_ed25519
      
      - name: Run Packer
        run: packer build template.pkr.hcl
```

Create secret:
```bash
gh secret set SSH_PRIVATE_KEY < ~/.ssh/id_ed25519
```

#### HTTPS Token Authentication

```yaml
name: Packer Build
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Configure Git Credentials
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          git config --global credential.helper store
          echo "https://${GITHUB_TOKEN}@github.com" > ~/.git-credentials
      
      - name: Run Packer
        run: packer build template.pkr.hcl
```

#### Private Automation Hub Token

```yaml
name: Packer Build
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Packer
        env:
          AUTOMATION_HUB_TOKEN: ${{ secrets.AUTOMATION_HUB_TOKEN }}
        run: |
          packer build \
            -var "automation_hub_token=${AUTOMATION_HUB_TOKEN}" \
            template.pkr.hcl
```

### GitLab CI

#### SSH Key Setup

```yaml
packer_build:
  stage: build
  script:
    - eval $(ssh-agent -s)
    - echo "$SSH_PRIVATE_KEY" | tr -d '\r' | ssh-add -
    - mkdir -p ~/.ssh
    - ssh-keyscan github.com >> ~/.ssh/known_hosts
    - packer build template.pkr.hcl
  variables:
    SSH_PRIVATE_KEY: $SSH_PRIVATE_KEY  # Defined in GitLab CI/CD Variables
```

Create variable:
- Go to Settings → CI/CD → Variables
- Add `SSH_PRIVATE_KEY` with your private key content
- Mark as **Protected** and **Masked**

#### HTTPS Token Authentication

```yaml
packer_build:
  stage: build
  before_script:
    - git config --global credential.helper store
    - echo "https://oauth2:${GITLAB_TOKEN}@gitlab.com" > ~/.git-credentials
  script:
    - packer build template.pkr.hcl
  variables:
    GITLAB_TOKEN: $GITLAB_TOKEN  # Protected variable
```

### Jenkins

#### SSH Key with Credentials Binding

```groovy
pipeline {
  agent any
  
  stages {
    stage('Packer Build') {
      steps {
        sshagent(['github-ssh-key']) {  // Credential ID in Jenkins
          sh '''
            eval $(ssh-agent -s)
            packer build template.pkr.hcl
          '''
        }
      }
    }
  }
}
```

Add SSH key:
- Manage Jenkins → Credentials → Add Credentials
- Kind: **SSH Username with private key**
- ID: `github-ssh-key`

#### HTTPS Token with Credentials Binding

```groovy
pipeline {
  agent any
  
  stages {
    stage('Packer Build') {
      steps {
        withCredentials([string(credentialsId: 'github-token', variable: 'GITHUB_TOKEN')]) {
          sh '''
            git config --global credential.helper store
            echo "https://${GITHUB_TOKEN}@github.com" > ~/.git-credentials
            packer build template.pkr.hcl
          '''
        }
      }
    }
  }
}
```

---

## Troubleshooting

For common authentication error messages and solutions, see [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md#authentication-failures).

### Quick Reference

| Error Message | Likely Cause | Solution |
|---------------|--------------|----------|
| "Permission denied (publickey)" | SSH key not configured | [SSH Authentication](#private-github-repositories-ssh) |
| "fatal: Authentication failed" | Git credential helper not set up | [HTTPS Authentication](#private-github-repositories-https-with-tokens) |
| "HTTP Error 401: Unauthorized" | Invalid Galaxy API token | [Private Automation Hub](#private-automation-hub--red-hat-automation-hub) |
| "Could not find/install packages" | Missing authentication or wrong source | Check if collections require auth |

---

## Provisioner Type Coverage

This guide applies to both provisioner types:

- **[`ansible-navigator`](../openspec/specs/remote-provisioner-capabilities/spec.md)**: Remote provisioner (SSH-based)
- **[`ansible-navigator-local`](../openspec/specs/local-provisioner-capabilities/spec.md)**: Local provisioner (chroot/container)

Both provisioners use the same authentication mechanisms since they delegate to `ansible-galaxy`.

---

## Additional Resources

- [ansible-galaxy documentation](https://docs.ansible.com/ansible/latest/cli/ansible-galaxy.html)
- [GitHub Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [GitLab Deploy Tokens](https://docs.gitlab.com/ee/user/project/deploy_tokens/)
- [Private Automation Hub Collection Management](https://access.redhat.com/documentation/en-us/red_hat_ansible_automation_platform/)
