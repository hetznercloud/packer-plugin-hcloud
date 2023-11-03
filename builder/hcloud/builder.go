// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/packer-plugin-hcloud/version"
)

// The unique id for the builder
const BuilderId = "hcloud.builder"

type Builder struct {
	config       Config
	runner       multistep.Runner
	hcloudClient *hcloud.Client
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	opts := []hcloud.ClientOption{
		hcloud.WithToken(b.config.HCloudToken),
		hcloud.WithEndpoint(b.config.Endpoint),
		hcloud.WithBackoffFunc(hcloud.ConstantBackoff(b.config.PollInterval)),
		hcloud.WithApplication("hcloud-packer", version.PluginVersion.String()),
	}
	b.hcloudClient = hcloud.NewClient(opts...)
	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hcloudClient", b.hcloudClient)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepPreValidate{
			Force:        b.config.PackerForce,
			SnapshotName: b.config.SnapshotName,
		},
		&communicator.StepSSHKeyGen{
			CommConf:            &b.config.Comm,
			SSHTemporaryKeyPair: b.config.Comm.SSH.SSHTemporaryKeyPair,
		},
		multistep.If(b.config.PackerDebug && b.config.Comm.SSHPrivateKeyFile == "",
			&communicator.StepDumpSSHKey{
				Path: fmt.Sprintf("ssh_key_%s.pem", b.config.PackerBuildName),
				SSH:  &b.config.Comm.SSH,
			},
		),
		&stepCreateSSHKey{},
		&stepCreateServer{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      getServerIP,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepShutdownServer{},
		&stepCreateSnapshot{},
	}
	// Run the steps
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)
	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("snapshot_name"); !ok {
		return nil, nil
	}

	artifact := &Artifact{
		snapshotName: state.Get("snapshot_name").(string),
		snapshotId:   state.Get("snapshot_id").(int64),
		hcloudClient: b.hcloudClient,
		StateData:    map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}
