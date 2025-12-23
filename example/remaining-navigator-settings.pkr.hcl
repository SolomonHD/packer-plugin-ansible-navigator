# Example demonstrating the remaining ansible-navigator top-level settings supported by navigator_config.

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

    navigator_config {
      # New top-level settings (ansible-navigator.yml)
      format = "yaml"
      time_zone = "America/New_York"
      inventory_columns = ["name", "address"]
      collection_doc_cache_path = "/tmp/collection-doc-cache"

      color {
        enable = true
        osc4   = true
      }

      editor {
        command = "vim"
        console = true
      }

      images {
        details = ["everything"]
      }
    }
  }
}

