// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
type stepPreValidate struct {
	Force        bool
	SnapshotName string
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	ui.Say("Prevalidating server types")
	serverType, _, err := client.ServerType.Get(ctx, c.ServerType)
	if err != nil {
		err = fmt.Errorf("Error: getting server type: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if serverType == nil {
		err = fmt.Errorf("Error: server type '%s' not found", c.ServerType)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("serverType", serverType)

	if c.UpgradeServerType != "" {
		upgradeServerType, _, err := client.ServerType.Get(ctx, c.UpgradeServerType)
		if err != nil {
			err = fmt.Errorf("Error: getting upgrade server type: %w", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if serverType == nil {
			err = fmt.Errorf("Error: upgrade server type '%s' not found", c.UpgradeServerType)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if serverType.Architecture != upgradeServerType.Architecture {
			// This is also validated by API, but if we validate it here, its faster and we never have to create
			// a server in the first place. Saving users to first hour of billing.
			err = fmt.Errorf("Error: server_type and upgrade_server_type have incompatible architectures")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say(fmt.Sprintf("Prevalidating snapshot name: %s", s.SnapshotName))

	// We would like to ask only for snapshots with a certain name using
	// ImageListOpts{Name: s.SnapshotName}, but snapshots do not have name, they
	// only have description. Thus we are obliged to ask for _all_ the snapshots.
	opts := hcloud.ImageListOpts{
		Type:         []hcloud.ImageType{hcloud.ImageTypeSnapshot},
		Architecture: []hcloud.Architecture{serverType.Architecture},
	}
	snapshots, err := client.Image.AllWithOpts(ctx, opts)
	if err != nil {
		err := fmt.Errorf("Error: getting snapshot list: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, snap := range snapshots {
		if snap.Description == s.SnapshotName {
			snapMsg := fmt.Sprintf("snapshot name: '%s' is used by existing snapshot with ID %d (arch=%s)",
				s.SnapshotName, snap.ID, serverType.Architecture)
			if s.Force {
				ui.Say(snapMsg + ". Force flag specified, will safely overwrite this snapshot")
				state.Put(OldSnapshotID, snap.ID)
				return multistep.ActionContinue
			}
			err := fmt.Errorf("Error: " + snapMsg)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// no snapshot with the same name found
	return multistep.ActionContinue
}

// No-op
func (s *stepPreValidate) Cleanup(multistep.StateBag) {
}
