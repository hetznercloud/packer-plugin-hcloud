// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,imageFilter

package hcloud

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// Configures the client token used for authentication. It can also be specified
	// with the `HCLOUD_TOKEN` environment variable.
	HCloudToken string `mapstructure:"token" required:"true"`
	// Configures the client endpoint. It can also be specified with the
	// `HCLOUD_ENDPOINT` environment variable.
	Endpoint string `mapstructure:"endpoint"`
	// Configures the interval at which the API is polled by the client. Default
	// `500ms`. Increase this interval if you run into rate limiting errors.
	PollInterval time.Duration `mapstructure:"poll_interval"`

	// Name assigned to the server. Hetzner Cloud sets the hostname of the server to
	// this value.
	ServerName string `mapstructure:"server_name"`
	// Name of the location where to create the server.
	Location string `mapstructure:"location" required:"true"`
	// ID or name of the server type used to create the server.
	ServerType string `mapstructure:"server_type" required:"true"`
	// ID or name of the server type the server should be upgraded to, without changing
	// the disk size. This improves building performance and the resulting snapshot is
	// compatible with smaller server types and disk sizes.
	UpgradeServerType string `mapstructure:"upgrade_server_type"`

	// ID or name of image to launch server from. Alternatively you can use
	// `image_filter`.
	Image string `mapstructure:"image" required:"true"`
	// Filters used to populate the `image` field. You may set this in place of `image`,
	// but not both.
	//
	// This selects the most recent image with the label `name==my-image`:
	//
	// ```hcl
	// image_filter {
	//   most_recent   = true
	//   with_selector = ["name==my-image"]
	// }
	// ```
	//
	// NOTE: This will fail unless _exactly_ one image is returned. In the above
	// example, `most_recent` will cause this to succeed by selecting the newest image.
	//
	// @include 'builder/hcloud/imageFilter-not-required.mdx'
	ImageFilter *imageFilter `mapstructure:"image_filter"`

	// Name of the resulting snapshot that will appear in your project as image
	// description. Defaults to `packer-{{timestamp}}` (see [configuration
	// templates](/packer/docs/templates/legacy_json_templates/engine) for more info).
	// The `snapshot_name` must be unique per architecture. If you want to reference the
	// image as a sample in your terraform configuration please use the image id or the
	// `snapshot_labels`.
	SnapshotName string `mapstructure:"snapshot_name"`
	// Key/value pair labels to apply to the created image.
	SnapshotLabels map[string]string `mapstructure:"snapshot_labels"`
	// User data to launch the server with. Packer will not automatically wait for a
	// user script to finish before shutting down the instance this must be handled in a
	// provisioner.
	UserData string `mapstructure:"user_data"`
	// Path to a file that will be used for the user data when launching the server. See
	// the `user_data` field.
	UserDataFile string `mapstructure:"user_data_file"`
	// List of SSH keys names or IDs to be added to image on launch.
	//
	// @include 'packer-plugin-sdk/communicator/SSHTemporaryKeyPair-not-required.mdx'
	SSHKeys []string `mapstructure:"ssh_keys"`
	// List of Network IDs to attach to the server private network interface at creation
	// time.
	Networks []int64 `mapstructure:"networks"`
	// Enable and boot in to the specified rescue system. This enables simple
	// installation of custom operating systems. `linux64` or `linux32`
	RescueMode string `mapstructure:"rescue"`

	ctx interpolate.Context
}

type imageFilter struct {
	// Label selectors used to select an `image`. See the [Label Selectors
	// docs](https://docs.hetzner.cloud/#label-selector) for more info.
	//
	// NOTE: This will fail unless _exactly_ one image is returned.
	WithSelector []string `mapstructure:"with_selector"`
	// Selects the newest created image when true. This is useful if you base your image
	// on another Packer build image.
	MostRecent bool `mapstructure:"most_recent"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Defaults
	if c.HCloudToken == "" {
		c.HCloudToken = os.Getenv("HCLOUD_TOKEN")
	}
	if c.Endpoint == "" {
		if os.Getenv("HCLOUD_ENDPOINT") != "" {
			c.Endpoint = os.Getenv("HCLOUD_ENDPOINT")
		} else {
			c.Endpoint = hcloud.Endpoint
		}
	}
	if c.PollInterval == 0 {
		c.PollInterval = 500 * time.Millisecond
	}

	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}
		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.ServerName == "" {
		// Default to packer-[time-ordered-uuid]
		c.ServerName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	var errs *packersdk.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}
	if c.HCloudToken == "" {
		// Required configurations that will display errors if not set
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("token for auth must be specified"))
	}

	if c.Location == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("location is required"))
	}

	if c.ServerType == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("server type is required"))
	}

	if c.Image == "" && c.ImageFilter == nil {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("image or image_filter is required"))
	}
	if c.ImageFilter != nil {
		if len(c.ImageFilter.WithSelector) == 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("image_filter.with_selector is required when specifying filter"))
		} else if c.Image != "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("only one of image or image_filter can be specified"))
		}
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("only one of user_data or user_data_file can be specified"))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packersdk.LogSecretFilter.Set(c.HCloudToken)
	return nil, nil
}

func getServerIP(state multistep.StateBag) (string, error) {
	return state.Get("server_ip").(string), nil
}
