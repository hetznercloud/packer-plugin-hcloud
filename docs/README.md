The `hcloud` Packer plugin is able to create new images for use with [Hetzner
Cloud](https://www.hetzner.cloud).

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    hcloud = {
      source  = "github.com/hetznercloud/hcloud"
      version = "~> 1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/hetznercloud/hcloud
```

#### Migrate from `github.com/hashicorp/hcloud`

On the 7 November 2023, the Packer plugin source moved from `github.com/hashicorp/hcloud` to `github.com/hetznercloud/hcloud`, make sure to update your Packer configuration with the new source:

```patch
 packer {
   required_plugins {
     hcloud = {
-      source  = "github.com/hashicorp/hcloud"
+      source  = "github.com/hetznercloud/hcloud"
       version = ">= 1.1.1"
     }
   }
 }
```

### Components

#### Builders

- [hcloud](/packer/integrations/hetznercloud/hcloud/latest/components/builder/hcloud) - The hcloud builder
  lets you create custom images on Hetzner Cloud by launching an instance, provisioning it, then
  export it as an image for later reuse.
