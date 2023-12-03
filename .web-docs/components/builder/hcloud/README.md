Type: `hcloud`
Artifact BuilderId: `hcloud.builder`

The `hcloud` Packer builder is able to create new images for use with [Hetzner
Cloud](https://www.hetzner.cloud). The builder takes a source image, runs any
provisioning necessary on the image after launching it, then snapshots it into
a reusable image. This reusable image can then be used as the foundation of new
servers that are launched within the Hetzner Cloud.

The builder does _not_ manage images. Once it creates an image, it is up to you
to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/packer/docs/templates/legacy_json_templates/communicator) can be configured for this
builder.

### Required Builder Configuration options:

<!-- Code generated from the comments of the Config struct in builder/hcloud/config.go; DO NOT EDIT MANUALLY -->

- `token` (string) - Configures the client token used for authentication. It can also be specified
  with the `HCLOUD_TOKEN` environment variable.

- `location` (string) - Name of the location where to create the server.

- `server_type` (string) - ID or name of the server type used to create the server.

- `image` (string) - ID or name of image to launch server from. Alternatively you can use
  `image_filter`.

<!-- End of code generated from the comments of the Config struct in builder/hcloud/config.go; -->


### Optional:

<!-- Code generated from the comments of the Config struct in builder/hcloud/config.go; DO NOT EDIT MANUALLY -->

- `endpoint` (string) - Configures the client endpoint. It can also be specified with the
  `HCLOUD_ENDPOINT` environment variable.

- `poll_interval` (duration string | ex: "1h5m2s") - Configures the interval at which the API is polled by the client. Default
  `500ms`. Increase this interval if you run into rate limiting errors.

- `server_name` (string) - Name assigned to the server. Hetzner Cloud sets the hostname of the server to
  this value.

- `upgrade_server_type` (string) - ID or name of the server type the server should be upgraded to, without changing
  the disk size. This improves building performance and the resulting snapshot is
  compatible with smaller server types and disk sizes.

- `image_filter` (\*imageFilter) - Filters used to populate the `image` field. You may set this in place of `image`,
  but not both.
  
  This selects the most recent image with the label `name==my-image`:
  
  ```hcl
  image_filter {
    most_recent   = true
    with_selector = ["name==my-image"]
  }
  ```
  
  NOTE: This will fail unless _exactly_ one image is returned. In the above
  example, `most_recent` will cause this to succeed by selecting the newest image.
  
  <!-- Code generated from the comments of the imageFilter struct in builder/hcloud/config.go; DO NOT EDIT MANUALLY -->

- `with_selector` ([]string) - Label selectors used to select an `image`. See the [Label Selectors
  docs](https://docs.hetzner.cloud/#label-selector) for more info.
  
  NOTE: This will fail unless _exactly_ one image is returned.

- `most_recent` (bool) - Selects the newest created image when true. This is useful if you base your image
  on another Packer build image.

<!-- End of code generated from the comments of the imageFilter struct in builder/hcloud/config.go; -->


- `snapshot_name` (string) - Name of the resulting snapshot that will appear in your project as image
  description. Defaults to `packer-{{timestamp}}` (see [configuration
  templates](/packer/docs/templates/legacy_json_templates/engine) for more info).
  The `snapshot_name` must be unique per architecture. If you want to reference the
  image as a sample in your terraform configuration please use the image id or the
  `snapshot_labels`.

- `snapshot_labels` (map[string]string) - Key/value pair labels to apply to the created image.

- `user_data` (string) - User data to launch the server with. Packer will not automatically wait for a
  user script to finish before shutting down the instance this must be handled in a
  provisioner.

- `user_data_file` (string) - Path to a file that will be used for the user data when launching the server. See
  the `user_data` field.

- `ssh_keys` ([]string) - List of SSH keys names or IDs to be added to image on launch.
  
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


- `networks` ([]int64) - List of Network IDs to attach to the server private network interface at creation
  time.

- `rescue` (string) - Enable and boot in to the specified rescue system. This enables simple
  installation of custom operating systems. `linux64` or `linux32`

<!-- End of code generated from the comments of the Config struct in builder/hcloud/config.go; -->


## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

```hcl
source "hcloud" "basic_example" {
  token = "YOUR API TOKEN"
  image = "ubuntu-22.04"
  location = "nbg1"
  server_type = "cx11"
  ssh_username = "root"
}

build {
  sources  = ["source.hcloud.basic_example"]
}
```
