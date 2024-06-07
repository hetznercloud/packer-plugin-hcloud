# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    hcloud = {
      source  = "github.com/hetznercloud/hcloud"
      version = ">=1.1.0"
    }
  }
}

variable "hcloud_token" {
  type      = string
  sensitive = true
  default   = "${env("HCLOUD_TOKEN")}"
}

source "hcloud" "example" {
  token = var.hcloud_token

  location    = "hel1"
  image       = "ubuntu-24.04"
  server_type = "cx22"
  server_name = "hcloud-example"

  ssh_username = "root"

  snapshot_name = "hcloud-example"
  snapshot_labels = {
    app = "hcloud-example"
  }
}

build {
  sources = ["source.hcloud.example"]

  provisioner "shell" {
    inline = ["cloud-init status --wait || test $? -eq 2"]
  }

  provisioner "shell" {
    inline = ["echo 'Hello World!' > /var/log/packer.log"]
  }
}
