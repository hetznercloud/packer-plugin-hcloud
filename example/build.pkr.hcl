# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    hcloud = {
      version = ">=1.1.0"
      source  = "github.com/hetznercloud/hcloud"
    }
  }
}

variable "hcloud_token" {
  type      = string
  default   = "${env("HCLOUD_TOKEN")}"
  sensitive = true
}

source "hcloud" "example" {
  image       = "ubuntu-22.04"
  location    = "hel1"
  server_name = "hcloud-example"
  server_type = "cx11"
  snapshot_labels = {
    app = "hcloud-example"
  }
  snapshot_name = "hcloud-example"
  ssh_username  = "root"
  token         = var.hcloud_token
}

build {
  sources = ["source.hcloud.example"]

  provisioner "shell" {
    inline = ["cloud-init status --wait"]
  }

  provisioner "shell" {
    inline = ["echo \"Hello World!\" > /var/log/packer.log"]
  }
}
