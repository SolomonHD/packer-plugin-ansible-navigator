# Example demonstrating flat (simple) navigator_config
# This uses typed structs with block syntax (replace-navigator-config-with-typed-structs)

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

    # Simple navigator_config with just mode setting
    # Uses block syntax (no `=`)
    navigator_config {
      mode = "stdout"
    }
  }
}
