# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "Hetzner Cloud"
  description = "The hcloud plugin can be used with HashiCorp Packer to create custom images on Hetzner Cloud."
  identifier = "packer/hashicorp/hcloud"
  component {
    type = "builder"
    name = "Hetzner Cloud"
    slug = "hcloud"
  }
}
