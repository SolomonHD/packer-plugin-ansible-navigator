# Example demonstrating additional ansible.cfg section blocks under navigator_config.ansible_config

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
      name   = "Run playbook with expanded ansible.cfg sections"
      target = "./playbook.yml"
    }

    navigator_config {
      mode = "stdout"

      ansible_config {
        # Existing sections
        defaults {
          host_key_checking = false
        }

        ssh_connection {
          ssh_timeout = 30
          pipelining  = true
        }

        # New sections (generated into ansible.cfg)
        privilege_escalation {
          become        = true
          become_method = "sudo"
          become_user   = "root"
        }

        persistent_connection {
          connect_timeout       = 30
          connect_retry_timeout = 15
          command_timeout       = 60
        }

        inventory {
          enable_plugins = ["ini", "yaml"]
        }

        galaxy {
          server_list   = ["automation_hub"]
          ignore_certs  = true
        }
      }
    }
  }
}

