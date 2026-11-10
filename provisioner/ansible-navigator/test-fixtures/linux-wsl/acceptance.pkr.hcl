packer {
  required_version = ">= 1.14.2"
  required_plugins {
    docker = {
      source  = "github.com/hashicorp/docker"
      version = "= 1.1.2"
    }
    ansible-navigator = {
      source  = "github.com/solomonhd/ansible-navigator"
      version = ">= 0.0.0"
    }
  }
}

variable "navigator_command" {
  type = string
}
variable "navigator_path" {
  type    = list(string)
  default = []
}
variable "target_image" {
  type = string
}
variable "ee_image" {
  type = string
}
variable "run_id" {
  type = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9_-]+$", var.run_id))
    error_message = "Run ID must be a harmless label value."
  }
}
variable "marker_nonce" {
  type = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9_-]+$", var.marker_nonce))
    error_message = "Marker nonce must contain only letters, digits, underscores and hyphens."
  }
}
variable "expect_failure" {
  type    = bool
  default = false
}

source "docker" "acceptance" {
  image   = var.target_image
  pull    = false
  discard = true
  run_command = [
    "-d", "-i", "-t", "--entrypoint=/bin/sh",
    "--label=io.packer.navigator.acceptance=${var.run_id}",
    "--", "{{.Image}}"
  ]
}

build {
  sources = ["source.docker.acceptance"]

  # These execute through the Docker builder communicator on the target.
  provisioner "shell" {
    inline = [
      "test ! -e /var/tmp/packer-navigator-acceptance-marker",
      "echo ACCEPTANCE_ABSENT:${var.marker_nonce}"
    ]
  }

  provisioner "ansible-navigator" {
    command                = var.navigator_command
    ansible_navigator_path = var.navigator_path
    skip_version_check     = false
    use_proxy              = true
    user                   = "root"
    play {
      name   = "Write the independent target marker"
      target = "${path.root}/marker.yml"
      extra_vars = {
        marker_nonce   = var.marker_nonce
        expect_failure = var.expect_failure ? "true" : "false"
      }
    }
    navigator_config {
      mode = "stdout"
      execution_environment {
        enabled          = true
        container_engine = "docker"
        image            = var.ee_image
        pull_policy      = "never"
        container_options = [
          "--label=io.packer.navigator.acceptance=${var.run_id}"
        ]
      }
      playbook_artifact {
        enable = false
      }
    }
  }

  provisioner "shell" {
    inline = [
      "python3 -c 'import pathlib,sys; assert pathlib.Path(\"/var/tmp/packer-navigator-acceptance-marker\").read_bytes() == sys.argv[1].encode()' '${var.marker_nonce}'",
      "echo ACCEPTANCE_VERIFIED:${var.marker_nonce}"
    ]
  }
}
