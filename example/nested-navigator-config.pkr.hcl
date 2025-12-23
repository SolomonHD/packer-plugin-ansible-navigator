# Example demonstrating nested navigator_config with execution-environment support
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
    # Use play blocks instead of deprecated playbook_file
    play {
      name   = "Test playbook"
      target = "./playbook.yml"
    }

    # Nested navigator_config using typed structs with block syntax
    # Note: Uses underscores instead of hyphens for field names
    navigator_config {
      # Mode setting
      mode = "stdout"

      # Execution environment settings (nested block)
      execution_environment {
        enabled      = true
        image        = "quay.io/ansible/ansible-navigator:latest"
        pull_policy  = "missing"
        
        environment_variables {
          pass = ["SSH_AUTH_SOCK", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"]
          set = {
            ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
            ANSIBLE_LOCAL_TMP  = "/tmp/.ansible-local"
          }
        }
      }

      # Logging settings (nested block)
      logging {
        level = "debug"
        file  = "/tmp/ansible-navigator.log"
        append = true
      }

      # Ansible settings (nested block)
      ansible_config {
        defaults {
          remote_tmp       = "/tmp/.ansible/tmp"
          host_key_checking = false
        }
        
        ssh_connection {
          ssh_timeout = 30
          pipelining  = true
        }
      }

      # Playbook artifact settings (nested block)
      playbook_artifact {
        enable   = false
        save_as  = "/tmp/artifact.json"
        replay   = "/tmp/artifact.json"
      }
    }
  }
}
