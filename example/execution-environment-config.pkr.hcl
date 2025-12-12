# Example demonstrating execution environment configuration with ansible-navigator
# This shows how to configure container execution environments for Ansible playbooks

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
      name   = "Test playbook with execution environment"
      target = "./playbook.yml"
    }

    # Configure ansible-navigator with execution environment
    navigator_config {
      mode = "stdout"
      
      # Execution environment configuration
      execution_environment {
        enabled      = true
        image        = "quay.io/ansible/creator-ee:latest"
        pull_policy  = "missing"
        
        # Pass environment variables to the execution environment
        environment_variables {
          # List of environment variables to pass through from host
          pass = [
            "SSH_AUTH_SOCK",
            "AWS_ACCESS_KEY_ID",
            "AWS_SECRET_ACCESS_KEY",
            "ANSIBLE_VAULT_PASSWORD"
          ]
          # Set specific environment variables in the EE
          set = {
            ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
            ANSIBLE_ROLES_PATH = "/opt/ansible/roles"
          }
        }
      }
      
      # Ansible configuration
      ansible_config {
        config = "/etc/ansible/ansible.cfg"
        
        defaults {
          host_key_checking = false
          remote_tmp        = "/tmp/.ansible/tmp"
        }
        
        ssh_connection {
          ssh_timeout = 30
          pipelining  = true
        }
      }
      
      # Logging configuration
      logging {
        level  = "debug"
        file   = "/tmp/ansible-navigator.log"
        append = true
      }
      
      # Playbook artifact settings (useful for debugging)
      playbook_artifact {
        enable  = true
        save_as = "/tmp/playbook-artifact.json"
      }
    }
  }
}
