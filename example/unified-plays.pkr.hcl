# Example demonstrating the new unified play execution model with requirements_file

packer {
  required_plugins {
    docker = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/docker"
    }
    ansible = {
      version = ">= 1.0.0"
      source  = "github.com/solomonhd/ansible-navigator"
    }
  }
}

source "docker" "ubuntu" {
  image  = "ubuntu:22.04"
  commit = true
}

build {
  sources = ["source.docker.ubuntu"]

  provisioner "ansible-navigator" {
    # Unified requirements file for both roles and collections
    requirements_file = "./requirements.yml"
    
    # Cache directories for offline use
    collections_cache_dir = "~/.packer.d/ansible_collections_cache"
    roles_cache_dir       = "~/.packer.d/ansible_roles_cache"
    
    # Dependency management options
    offline_mode = false
    force_update = false

    # Array of plays - supports both playbooks and role FQDNs
    plays = [
      {
        name   = "Setup base system"
        target = "geerlingguy.docker"
        extra_vars = {
          docker_install_compose = "true"
          docker_edition         = "ce"
        }
      },
      {
        name       = "Deploy web stack"
        target     = "myorg.webserver.deploy"
        become     = true
        vars_files = ["vars/web.yml"]
        tags       = ["deploy", "web"]
      },
      {
        name   = "Custom playbook test"
        target = "site.yml"
        tags   = ["test"]
      }
    ]
  }
}