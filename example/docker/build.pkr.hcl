# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    hcloud = {
      source  = "github.com/hetznercloud/hcloud"
      version = ">=1.5.1"
    }
  }
}

variable "hcloud_token" {
  type      = string
  sensitive = true
  default   = "${env("HCLOUD_TOKEN")}"
}

source "hcloud" "docker" {
  token = var.hcloud_token

  location    = "hel1"
  image       = "ubuntu-24.04"
  server_type = "cpx22"
  server_name = "docker-{{ timestamp }}"

  user_data = <<-EOF
    #cloud-config
    growpart:
      mode: "off"
    resize_rootfs: false
  EOF

  ssh_username = "root"

  snapshot_name = "docker-{{ timestamp }}"
  snapshot_labels = {
    app = "docker"
  }
}

build {
  sources = ["source.hcloud.docker"]

  provisioner "shell" {
    inline           = ["cloud-init status --wait --long"]
    valid_exit_codes = [0, 2]
  }

  provisioner "shell" {
    scripts = [
      "install.sh",
      "upgrade.sh",
      "cleanup.sh",
    ]
  }
}
