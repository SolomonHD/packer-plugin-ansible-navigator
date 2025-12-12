# Example demonstrating flat (string-value) navigator_config
# This tests backward compatibility after the fix for HCL2 type specification

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
    play {
      name   = "Test playbook"
      target = "./playbook.yml"
    }

    # Flat navigator_config - simple string values (backward compatibility)
    navigator_config = {
      mode = "stdout"
    }
  }
}
