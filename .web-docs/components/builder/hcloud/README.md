Type: `hcloud`
Artifact BuilderId: `hcloud.builder`

The `hcloud` Packer builder is able to create new images for use with [Hetzner
Cloud](https://www.hetzner.com/cloud/). The builder takes a source image, runs any
provisioning necessary on the image after launching it, then snapshots it into
a reusable image. This reusable image can then be used as the foundation of new
servers that are launched within the Hetzner Cloud.

The builder does _not_ manage images. Once it creates an image, it is up to you
to use it or delete it.

## Connection to the server

The builder will connect to the server using the first available IP, in the following order:

- `public_ipv4`: If enabled, the public IPv4 will be used,
- `public_ipv6`: If enabled, the public IPv6 will be used,
- `private_ipv4`: If the server is attached to private networks, the private IPv4 of the
  first private network will be used.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/packer/docs/templates/legacy_json_templates/communicator) can be configured for this
builder.

### Required Builder Configuration options:

- `token` (string) - The client TOKEN to use to access your account. It can
  also be specified via environment variable `HCLOUD_TOKEN`, if set.

- `image` (string) - ID or name of image to launch server from. Alternatively
  you can use `image_filter`.

- `location` (string) - The name of the location to launch the server in.

- `server_type` (string) - ID or name of the server type this server should
  be created with.

### Optional:

- `endpoint` (string) - Non standard api endpoint URL. Set this if you are
  using a Hetzner Cloud API compatible service. It can also be specified via
  environment variable `HCLOUD_ENDPOINT`.

- `image_filter` (object) - Filters used to populate the `filter`
  field. Example:

  ```hcl
  image_filter {
    most_recent   = true
    with_selector = ["name==my-image"]
  }
  ```

  This selects the most recent image with the label `name==my-image`. NOTE:
  This will fail unless _exactly_ one AMI is returned. In the above example,
  `most_recent` will cause this to succeed by selecting the newest image.

  - `with_selector` (list of strings) - label selectors used to select an
    `image`. NOTE: This will fail unless _exactly_ one image is returned.
    Check the official hcloud docs on
    [Label Selectors](https://docs.hetzner.cloud/reference/cloud#label-selector)
    for more info.

  - `most_recent` (boolean) - Selects the newest created image when true.
    This is most useful if you base your image on another Packer build image.

  You may set this in place of `image`, but not both.

- `server_name` (string) - The name assigned to the server. The Hetzner Cloud
  sets the hostname of the machine to this value.

- `server_labels` (map of key/value strings) - Key/value pair labels to
  apply to the created server.

- `skip_create_snapshot` (boolean) - Skips snapshot creation.

- `snapshot_name` (string) - The name of the resulting snapshot that will
  appear in your account as image description. Defaults to `packer-{{timestamp}}` (see
  [configuration templates](/packer/docs/templates/legacy_json_templates/engine) for more info).
  The snapshot_name must be unique per architecture.
  If you want to reference the image as a sample in your terraform configuration please use the image id or the `snapshot_labels`.

- `snapshot_labels` (map of key/value strings) - Key/value pair labels to
  apply to the created image.

- `poll_interval` (string) - Configures the interval in which actions are
  polled by the client. Default `500ms`. Increase this interval if you run
  into rate limiting errors.

- `user_data` (string) - User data to launch with the server. Packer will not
  automatically wait for a user script to finish before shutting down the
  instance this must be handled in a provisioner.

- `user_data_file` (string) - Path to a file that will be used for the user
  data when launching the server.

- `ssh_keys_labels` (map of key/value strings) - Key/value pair labels to
  apply to the created ssh keys.

- `ssh_keys` (array of strings) - List of SSH keys by name or id to be added
  to image on launch.

<!-- Code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `temporary_key_pair_type` (string) - `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )
  
  Specifies the type of key to create. The possible values are 'dsa',
  'ecdsa', 'ed25519', or 'rsa'.
  
  NOTE: DSA is deprecated and no longer recognized as secure, please
  consider other alternatives like RSA or ED25519.

- `temporary_key_pair_bits` (int) - Specifies the number of bits in the key to create. For RSA keys, the
  minimum size is 1024 bits and the default is 4096 bits. Generally, 3072
  bits is considered sufficient. DSA keys must be exactly 1024 bits as
  specified by FIPS 186-2. For ECDSA keys, bits determines the key length
  by selecting from one of three elliptic curve sizes: 256, 384 or 521
  bits. Attempting to use bit lengths other than these three values for
  ECDSA keys will fail. Ed25519 keys have a fixed length and bits will be
  ignored.
  
  NOTE: DSA is deprecated and no longer recognized as secure as specified
  by FIPS 186-5, please consider other alternatives like RSA or ED25519.

<!-- End of code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; -->


- `rescue` (string) - Enable and boot in to the specified rescue system. This
  enables simple installation of custom operating systems. `linux64` or `linux32`

- `upgrade_server_type` (string) - ID or name of the server type this server should
  be upgraded to, without changing the disk size. Improves building performance.
  The resulting snapshot is compatible with smaller server types and disk sizes.

- `networks` (array of integers) - List of Network IDs which should be
  attached to the server private network interface at creation time.

- `public_ipv4` (string) - ID, name or IP address of a pre-allocated Hetzner
  Primary IPv4 address to use for the created server.

- `public_ipv4_disabled` (bool) - Disable the public ipv4 for the created server.

- `public_ipv6` (string) - ID, name or IP address of a pre-allocated Hetzner
  Primary IPv6 address to use for the created server.

- `public_ipv6_disabled` (bool) - Disable the public ipv6 for the created server.

- `firewalls` (array of strings) - List of Firewall by name or id to be attached
  to the created server.

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

```hcl
source "hcloud" "basic_example" {
  token = "YOUR API TOKEN"
  image = "ubuntu-24.04"
  location = "hel1"
  server_type = "cx23"
  ssh_username = "root"
}

build {
  sources  = ["source.hcloud.basic_example"]
}
```

See the [examples](https://github.com/hetznercloud/packer-plugin-hcloud/tree/main/example) folder in the Packer project for more examples.

## Tips

### Keeping the images size small

To reduce the size of your images, we recommend cleaning up any temporary files that
were added during the build process and discard the unused blocks from the disk.

Below are a few commands that might useful:

- Clean and reset could-init files:

  ```bash
  cloud-init clean --logs --machine-id --seed --configs all

  rm -rf /run/cloud-init/*
  rm -rf /var/lib/cloud/*
  ```

- Clean apt files:

  ```bash
  export DEBIAN_FRONTEND=noninteractive

  apt-get -y autopurge
  apt-get -y clean

  rm -rf /var/lib/apt/lists/*
  ```

- Clean logs:

  ```bash
  journalctl --flush
  journalctl --rotate --vacuum-time=0

  find /var/log -type f -exec truncate --size 0 {} \; # truncate system logs
  find /var/log -type f -name '*.[1-9]' -delete # remove archived logs
  find /var/log -type f -name '*.gz' -delete # remove compressed archived logs
  ```

- Reset host ssh keys:

  ```bash
  rm -f /etc/ssh/ssh_host_*_key /etc/ssh/ssh_host_*_key.pub
  ```

Once the cleanup is complete, you must discard the now unused blocks from the disk. This can be done with the following:

```bash
fstrim --all || true
sync
```

To speed up the process above, you may configure cloud-init to not grow the system partition to the server disk size during the boot process, leaving you with a 4GB system partition:

```yaml
#cloud-config
growpart:
  mode: "off"
resize_rootfs: false
```

See the [docker example](https://github.com/hetznercloud/packer-plugin-hcloud/tree/main/example/docker) in the Packer project for more details.
