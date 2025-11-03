# 🎨 Examples Gallery

Real-world examples and use cases for the Packer Plugin Ansible Navigator.

## Table of Contents

- [Basic Examples](#basic-examples)
- [Cloud Provider Examples](#cloud-provider-examples)
- [Container Examples](#container-examples)
- [Security and Compliance](#security-and-compliance)
- [CI/CD Integration](#cicd-integration)
- [Multi-Stage Deployments](#multi-stage-deployments)
- [Development Workflows](#development-workflows)
- [Production Patterns](#production-patterns)

## Basic Examples

### Hello World

The simplest possible configuration:

```hcl
source "null" "example" {
  communicator = "none"
}

build {
  sources = ["source.null.example"]
  
  provisioner "ansible-navigator" {
    playbook_file = "hello.yml"
  }
}
```

`hello.yml`:
```yaml
---
- name: Hello World
  hosts: all
  tasks:
    - name: Print message
      debug:
        msg: "Hello from Ansible Navigator!"
```

### Local Testing with Docker

```hcl
source "docker" "ubuntu" {
  image  = "ubuntu:22.04"
  commit = true
}

build {
  sources = ["source.docker.ubuntu"]
  
  provisioner "ansible-navigator" {
    playbook_file = "configure.yml"
    groups = ["docker", "test"]
    ansible_env_vars = [
      "ANSIBLE_HOST_KEY_CHECKING=False"
    ]
  }
}
```

## Cloud Provider Examples

### AWS EC2 AMI Building

```hcl
variable "aws_region" {
  default = "us-east-1"
}

variable "instance_type" {
  default = "t3.micro"
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "custom-ubuntu-{{timestamp}}"
  instance_type = var.instance_type
  region        = var.aws_region
  
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"] # Canonical
  }
  
  ssh_username = "ubuntu"
}

build {
  sources = ["source.amazon-ebs.ubuntu"]
  
  provisioner "ansible-navigator" {
    plays = [
      "aws.infrastructure.configure_base",
      "aws.infrastructure.install_cloudwatch",
      "aws.infrastructure.harden_ami"
    ]
    
    collections = [
      "amazon.aws:6.5.0",
      "community.aws:6.4.0"
    ]
    
    extra_arguments = [
      "--extra-vars", "aws_region=${var.aws_region}",
      "--extra-vars", "environment=production"
    ]
    
    # Use AWS-optimized execution environment
    execution_environment = "quay.io/ansible/creator-ee:latest"
  }
}
```

### Azure VM Image

```hcl
source "azure-arm" "windows" {
  use_azure_cli_auth = true
  
  managed_image_resource_group_name = "packer-images"
  managed_image_name                = "windows-server-2022-{{timestamp}}"
  
  os_type         = "Windows"
  image_publisher = "MicrosoftWindowsServer"
  image_offer     = "WindowsServer"
  image_sku       = "2022-datacenter"
  
  location = "East US"
  vm_size  = "Standard_D2s_v3"
  
  communicator   = "winrm"
  winrm_username = "packer"
  winrm_insecure = true
  winrm_use_ssl  = true
}

build {
  sources = ["source.azure-arm.windows"]
  
  provisioner "ansible-navigator" {
    playbook_file = "windows-setup.yml"
    
    collections = [
      "ansible.windows:2.3.0",
      "chocolatey.chocolatey:1.5.1"
    ]
    
    extra_arguments = [
      "--extra-vars", "ansible_winrm_server_cert_validation=ignore"
    ]
    
    ansible_env_vars = [
      "ANSIBLE_HOST_KEY_CHECKING=False"
    ]
  }
}
```

### Google Cloud Platform

```hcl
source "googlecompute" "centos" {
  project_id   = "my-project"
  source_image = "centos-stream-9-v20231115"
  zone         = "us-central1-a"
  
  image_name        = "custom-centos-{{timestamp}}"
  image_description = "CentOS Stream 9 with custom configuration"
  
  ssh_username = "packer"
  
  metadata = {
    enable-oslogin = "FALSE"
  }
}

build {
  sources = ["source.googlecompute.centos"]
  
  provisioner "ansible-navigator" {
    plays = [
      "gcp.compute.install_stackdriver",
      "gcp.compute.configure_networking",
      "baseline.linux.harden"
    ]
    
    collections = [
      "google.cloud:1.2.0",
      "ansible.posix:1.5.4"
    ]
  }
}
```

## Container Examples

### Building Docker Images

```hcl
source "docker" "app" {
  image = "python:3.11-slim"
  commit = true
  
  changes = [
    "EXPOSE 8000",
    "WORKDIR /app",
    "CMD [\"python\", \"app.py\"]"
  ]
}

build {
  sources = ["source.docker.app"]
  
  provisioner "ansible-navigator" {
    plays = [
      "containers.python.install_dependencies",
      "containers.python.configure_app"
    ]
    
    collections = [
      "community.docker:3.4.0",
      "community.general:7.5.0"
    ]
    
    extra_arguments = [
      "--extra-vars", "app_version=${var.version}",
      "--extra-vars", "pip_requirements=requirements.txt"
    ]
  }
  
  post-processor "docker-tag" {
    repository = "myregistry.io/myapp"
    tags       = ["${var.version}", "latest"]
  }
}
```

### Kubernetes-Ready Images

```hcl
source "docker" "k8s_app" {
  image = "registry.access.redhat.com/ubi9/ubi-minimal:latest"
  commit = true
}

build {
  sources = ["source.docker.k8s_app"]
  
  provisioner "ansible-navigator" {
    plays = [
      "kubernetes.apps.prepare_base",
      "kubernetes.apps.install_app",
      "kubernetes.apps.configure_healthchecks"
    ]
    
    collections = [
      "kubernetes.core:2.4.0",
      "community.general:7.5.0"
    ]
    
    # Use Kubernetes-focused execution environment
    execution_environment = "quay.io/ansible/creator-ee:latest"
    
    extra_arguments = [
      "--extra-vars", "k8s_namespace=production",
      "--extra-vars", "enable_metrics=true"
    ]
  }
}
```

## Security and Compliance

### CIS Hardened Image

```hcl
source "amazon-ebs" "hardened" {
  ami_name      = "cis-hardened-ubuntu-{{timestamp}}"
  instance_type = "t3.medium"
  region        = "us-east-1"
  
  source_ami_filter {
    filters = {
      name = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  
  ssh_username = "ubuntu"
}

build {
  sources = ["source.amazon-ebs.hardened"]
  
  provisioner "ansible-navigator" {
    plays = [
      "security.cis.ubuntu_level1",
      "security.cis.ubuntu_level2",
      "security.audit.configure_auditd"
    ]
    
    collections = [
      "community.general:7.5.0",
      "ansible.posix:1.5.4"
    ]
    
    # Enable structured logging for compliance reporting
    navigator_mode = "json"
    structured_logging = true
    log_output_path = "./compliance/cis-hardening-report.json"
    
    extra_arguments = [
      "--extra-vars", "cis_level=2",
      "--extra-vars", "enable_aide=true",
      "--extra-vars", "enable_ossec=true"
    ]
  }
  
  provisioner "shell" {
    inline = [
      "echo 'Running compliance scan...'",
      "sudo lynis audit system --quick"
    ]
  }
}
```

### HIPAA Compliant Infrastructure

```hcl
build {
  sources = ["source.amazon-ebs.base"]
  
  provisioner "ansible-navigator" {
    plays = [
      "compliance.hipaa.configure_encryption",
      "compliance.hipaa.setup_logging",
      "compliance.hipaa.access_controls",
      "compliance.hipaa.audit_configuration"
    ]
    
    requirements_file = "./requirements-hipaa.yml"
    
    extra_arguments = [
      "--extra-vars", "encryption_at_rest=true",
      "--extra-vars", "encryption_in_transit=true",
      "--extra-vars", "log_retention_days=2555",  # 7 years
      "--vault-password-file", ".vault-pass"
    ]
    
    ansible_env_vars = [
      "ANSIBLE_VAULT_PASSWORD_FILE=.vault-pass"
    ]
  }
}
```

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/build-image.yml
name: Build Image

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Packer
        uses: hashicorp/setup-packer@main
        with:
          version: latest
      
      - name: Initialize Packer
        run: packer init config.pkr.hcl
      
      - name: Validate Template
        run: packer validate config.pkr.hcl
      
      - name: Build Image
        run: packer build config.pkr.hcl
        env:
          PKR_VAR_version: ${{ github.sha }}
```

Packer configuration for CI/CD:

```hcl
variable "version" {
  type = string
}

source "docker" "ci" {
  image = "ubuntu:22.04"
  commit = true
}

build {
  sources = ["source.docker.ci"]
  
  provisioner "ansible-navigator" {
    plays = ["ci.build.prepare"]
    
    # Use consistent execution environment for CI
    execution_environment = "quay.io/ansible/creator-ee:v0.21.0"
    
    # Enable JSON logging for CI parsing
    navigator_mode = "json"
    structured_logging = true
    log_output_path = "build-${var.version}.json"
    
    timeout = "30m"
    max_retries = 2
  }
}
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - build
  - test

validate:
  stage: validate
  image: hashicorp/packer:latest
  script:
    - packer init .
    - packer validate .

build:
  stage: build
  image: hashicorp/packer:latest
  script:
    - packer init .
    - packer build -var "version=$CI_COMMIT_SHA" .
  artifacts:
    paths:
      - logs/
    expire_in: 1 week
```

## Multi-Stage Deployments

### Progressive Application Deployment

```hcl
locals {
  timestamp = regex_replace(timestamp(), "[- TZ:]", "")
}

source "amazon-ebs" "app" {
  ami_name      = "app-${local.timestamp}"
  instance_type = "t3.large"
  region        = "us-east-1"
  
  source_ami_filter {
    filters = {
      name = "amzn2-ami-hvm-*-x86_64-gp2"
    }
    most_recent = true
    owners      = ["amazon"]
  }
  
  ssh_username = "ec2-user"
}

build {
  sources = ["source.amazon-ebs.app"]
  
  # Stage 1: Base OS Configuration
  provisioner "ansible-navigator" {
    plays = ["infrastructure.base.configure_os"]
    collections = ["ansible.posix:1.5.4"]
    pause_before = "5s"
  }
  
  # Stage 2: Security Hardening
  provisioner "ansible-navigator" {
    plays = [
      "security.firewall.configure",
      "security.selinux.enforce"
    ]
    collections = ["community.general:7.5.0"]
  }
  
  # Stage 3: Install Dependencies
  provisioner "ansible-navigator" {
    plays = [
      "dependencies.runtime.install_java",
      "dependencies.runtime.install_nodejs",
      "dependencies.database.install_postgres_client"
    ]
    keep_going = false  # Stop if dependencies fail
  }
  
  # Stage 4: Deploy Application
  provisioner "ansible-navigator" {
    plays = ["app.backend.deploy"]
    
    extra_arguments = [
      "--extra-vars", "app_version=${var.app_version}",
      "--extra-vars", "environment=production"
    ]
    
    # Verify deployment
    navigator_mode = "json"
    structured_logging = true
    log_output_path = "./deployment-report.json"
  }
  
  # Stage 5: Post-deployment validation
  provisioner "shell" {
    inline = [
      "curl -f http://localhost:8080/health || exit 1",
      "systemctl is-active app.service || exit 1"
    ]
  }
}
```

## Development Workflows

### Local Development with Hot Reload

```hcl
source "docker" "dev" {
  image = "ubuntu:22.04"
  commit = false  # Don't commit during development
  
  # Mount local code
  volumes = {
    "./app" = "/workspace"
  }
}

build {
  sources = ["source.docker.dev"]
  
  provisioner "ansible-navigator" {
    plays = ["dev.environment.setup"]
    
    # Use local collection under development
    collections = [
      "mycompany.myapp@../ansible-collections/mycompany-myapp"
    ]
    
    # Force update to get latest changes
    collections_force_update = true
    
    extra_arguments = [
      "--extra-vars", "debug=true",
      "--extra-vars", "development_mode=true",
      "-vvv"  # Verbose output for debugging
    ]
    
    # Continue on errors for development
    keep_going = true
  }
}
```

### Testing Ansible Collections

```hcl
variable "collection_path" {
  type = string
  default = "../my-collection"
}

source "docker" "test" {
  image = "quay.io/ansible/creator-ee:latest"
  commit = false
  run_command = ["/bin/bash", "-c", "sleep infinity"]
}

build {
  sources = ["source.docker.test"]
  
  provisioner "ansible-navigator" {
    plays = [
      "test.collection.unit_tests",
      "test.collection.integration_tests"
    ]
    
    collections = [
      "mycollection@${var.collection_path}"
    ]
    
    collections_force_update = true
    
    ansible_env_vars = [
      "ANSIBLE_COLLECTIONS_PATH=/workspace/collections"
    ]
    
    navigator_mode = "json"
    structured_logging = true
    verbose_task_output = true
  }
}
```

## Production Patterns

### Blue-Green Deployment

```hcl
variable "deployment_color" {
  type = string
  validation {
    condition     = contains(["blue", "green"], var.deployment_color)
    error_message = "Deployment color must be 'blue' or 'green'."
  }
}

source "amazon-ebs" "production" {
  ami_name = "production-${var.deployment_color}-{{timestamp}}"
  # ... other configuration
}

build {
  sources = ["source.amazon-ebs.production"]
  
  provisioner "ansible-navigator" {
    plays = [
      "deploy.bluegreen.prepare_${var.deployment_color}",
      "deploy.bluegreen.install_app",
      "deploy.bluegreen.configure_routing"
    ]
    
    requirements_file = "./requirements.yml"
    
    extra_arguments = [
      "--extra-vars", "deployment_color=${var.deployment_color}",
      "--extra-vars", "target_group_arn=${var.target_group_arn}"
    ]
    
    # Production safeguards
    timeout = "45m"
    max_retries = 1
    
    # Detailed logging for production deployments
    navigator_mode = "json"
    structured_logging = true
    log_output_path = "./logs/deploy-${var.deployment_color}-{{timestamp}}.json"
  }
}
```

### Immutable Infrastructure

```hcl
locals {
  build_number = env("BUILD_NUMBER")
  git_commit   = env("GIT_COMMIT")
}

source "amazon-ebs" "immutable" {
  ami_name = "app-${local.build_number}-${substr(local.git_commit, 0, 7)}"
  
  tags = {
    Name         = "Immutable App Image"
    BuildNumber  = local.build_number
    GitCommit    = local.git_commit
    BuildDate    = timestamp()
    Immutable    = "true"
  }
  
  # ... other configuration
}

build {
  sources = ["source.amazon-ebs.immutable"]
  
  provisioner "ansible-navigator" {
    plays = [
      "immutable.build.install_everything",
      "immutable.build.configure_readonly",
      "immutable.build.seal_image"
    ]
    
    collections = [
      "company.immutable:1.0.0"
    ]
    
    extra_arguments = [
      "--extra-vars", "build_number=${local.build_number}",
      "--extra-vars", "git_commit=${local.git_commit}",
      "--extra-vars", "make_readonly=true"
    ]
    
    # No retries for immutable builds
    max_retries = 0
  }
  
  # Final step: make filesystem read-only
  provisioner "shell" {
    inline = [
      "sudo rm -rf /tmp/*",
      "sudo rm -rf /var/tmp/*",
      "history -c"
    ]
  }
}
```

### Disaster Recovery Setup

```hcl
source "amazon-ebs" "dr" {
  ami_name      = "dr-backup-{{timestamp}}"
  instance_type = "t3.xlarge"
  region        = var.dr_region
  
  # ... other configuration
}

build {
  sources = ["source.amazon-ebs.dr"]
  
  provisioner "ansible-navigator" {
    plays = [
      "dr.backup.install_tools",
      "dr.backup.configure_replication",
      "dr.backup.setup_monitoring",
      "dr.backup.test_recovery"
    ]
    
    collections = [
      "company.disaster_recovery:2.1.0",
      "community.postgresql:3.2.0",
      "community.mysql:3.7.2"
    ]
    
    extra_arguments = [
      "--extra-vars", "primary_region=${var.primary_region}",
      "--extra-vars", "dr_region=${var.dr_region}",
      "--extra-vars", "rpo_minutes=15",
      "--extra-vars", "rto_minutes=60"
    ]
    
    # Ensure all DR steps complete
    keep_going = false
    
    # Extended timeout for DR testing
    timeout = "2h"
  }
}
```

## Tips and Best Practices

1. **Use Execution Environments**: Always specify a pinned version for reproducibility
2. **Enable JSON Logging**: Essential for CI/CD and troubleshooting
3. **Version Everything**: Pin collection versions, execution environment tags
4. **Test Locally First**: Use Docker source for quick iteration
5. **Implement Health Checks**: Verify services after provisioning
6. **Use Vault for Secrets**: Never hardcode sensitive data
7. **Keep Plays Atomic**: Each play should do one thing well
8. **Document Requirements**: Include requirements.yml in version control

---

[← Configuration Reference](CONFIGURATION.md) | [Troubleshooting Guide →](TROUBLESHOOTING.md)