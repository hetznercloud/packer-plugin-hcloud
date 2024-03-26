# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    hcloud = {
      version = ">=v1.0.5"
      source  = "github.com/hashicorp/hcloud"
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
  hcp_packer_registry {
    bucket_name = "hcloud-hcp-test"
    description = "A nice test description"
    bucket_labels = {
      "foo" = "bar"
    }
  }

  sources = ["source.hcloud.example"]

  provisioner "shell" {
    inline = ["cloud-init status --wait"]
  }

  provisioner "shell" {
    inline = ["echo \"Hello World!\" > /var/log/packer.log"]
  }
}
