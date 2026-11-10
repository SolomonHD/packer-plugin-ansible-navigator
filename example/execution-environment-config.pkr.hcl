# Advanced EE settings for both provisioners. Validate this template before use.
# Run from any directory; playbook paths are relative to this template.
#
# host-ee: Navigator, Docker, and /opt/ansible/data must exist on the Packer host.
# target-ee: supply a prebuilt target image with Navigator, Docker, a reachable
# container daemon, and /opt/ansible/data. A plain OS image is insufficient.
# This example does not install a runtime or configure daemon access on the target.
# Select one build with -only='host-ee.*' or -only='target-ee.*'.

packer {
  required_plugins {
    docker = {
      version = ">= 1.1.2"
      source  = "github.com/hashicorp/docker"
    }
    ansible = {
      version = ">= 8.7.0"
      source  = "github.com/solomonhd/ansible-navigator"
    }
  }
}

variable "local_target_image" {
  type        = string
  description = "Prebuilt image with Navigator, Docker, daemon access, and mount sources for target-ee."
}

source "docker" "host_target" {
  image  = "python:3.12-slim"
  commit = true
}

source "docker" "navigator_target" {
  image  = var.local_target_image
  commit = true
}

build {
  name    = "host-ee"
  sources = ["source.docker.host_target"]

  provisioner "ansible-navigator" {
    play {
      name   = "Host-based EE"
      target = "${path.root}/execution-environment-playbook.yml"
    }

    navigator_config {
      mode = "stdout"
      execution_environment {
        enabled           = true
        image             = "quay.io/ansible/creator-ee:latest"
        container_engine  = "docker"
        container_options = ["--label=packer-example=true"]
        pull_policy       = "missing"
        pull_arguments    = ["--quiet"]

        # Source is on the Packer host; repeat this block for more mounts.
        volume_mounts {
          src     = "/opt/ansible/data"
          dest    = "/runner/data"
          options = "ro"
        }
      }
    }
  }
}

build {
  name    = "target-ee"
  sources = ["source.docker.navigator_target"]

  provisioner "ansible-navigator-local" {
    play {
      name   = "On-target EE"
      target = "${path.root}/execution-environment-playbook.yml"
    }

    navigator_config {
      mode = "stdout"
      execution_environment {
        enabled           = true
        image             = "quay.io/ansible/creator-ee:latest"
        container_engine  = "docker"
        container_options = ["--label=packer-example=true", "--tty=false"]
        pull_policy       = "missing"
        pull_arguments    = ["--quiet"]

        # Source is on the target where Navigator and Docker execute.
        volume_mounts {
          src     = "/opt/ansible/data"
          dest    = "/runner/data"
          options = "ro"
        }
      }
    }
  }
}
