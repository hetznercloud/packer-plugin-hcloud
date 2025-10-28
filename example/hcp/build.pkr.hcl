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

source "hcloud" "example" {
  token = var.hcloud_token

  location    = "hel1"
  image       = "ubuntu-24.04"
  server_type = "cpx22"
  server_name = "example-{{ timestamp }}"

  ssh_username = "root"

  snapshot_name = "example-{{ timestamp }}"
  snapshot_labels = {
    app = "example"
  }
}

build {
  hcp_packer_registry {
    description = "A nice test description"

    bucket_name = "hcloud-hcp-test"
    bucket_labels = {
      "packer version" = packer.version
    }
  }

  sources = ["source.hcloud.example"]

  provisioner "shell" {
    inline           = ["cloud-init status --wait --long"]
    valid_exit_codes = [0, 2]
  }

  provisioner "shell" {
    inline = ["echo 'Hello World!' > /var/log/packer.log"]
  }
}
