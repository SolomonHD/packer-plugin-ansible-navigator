# Example demonstrating nested navigator_config with execution-environment support
# This tests the fix for HCL2 type specification (fix-navigator-config-hcl2-type)

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

    # Nested navigator_config - this is what the fix enables
    # Previously this would fail with: "element 'execution-environment': string required"
    navigator_config = {
      # Execution environment settings (nested object)
      execution-environment = {
        enabled      = true
        image        = "quay.io/ansible/ansible-navigator:latest"
        pull-policy  = "missing"
        environment-variables = {
          pass = ["SSH_AUTH_SOCK", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"]
        }
      }

      # Mode setting
      mode = "stdout"

      # Logging settings (nested object)
      logging = {
        level = "debug"
        file  = "/tmp/ansible-navigator.log"
      }

      # Ansible settings (nested object)
      ansible = {
        config = {
          path = "/etc/ansible/ansible.cfg"
        }
        cmdline = "--forks 10"
      }

      # Playbook artifact settings (nested object)
      playbook-artifact = {
        enable   = false
        save-as  = "/tmp/artifact.json"
        replay   = "/tmp/artifact.json"
      }
    }
  }
}
