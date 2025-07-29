// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
type stepPreValidate struct {
	Force        bool
	SnapshotName string
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c, ui, client := UnpackState(state)

	ui.Say(fmt.Sprintf("Validating server types: %s", c.ServerType))
	serverType, _, err := client.ServerType.Get(ctx, c.ServerType)
	if err != nil {
		return errorHandler(state, ui, fmt.Sprintf("Could not fetch server type '%s'", c.ServerType), err)
	}
	if serverType == nil {
		return errorHandler(state, ui, "", fmt.Errorf("Could not find server type '%s'", c.ServerType))
	}
	state.Put(StateServerType, serverType)

	if c.UpgradeServerType != "" {
		ui.Say(fmt.Sprintf("Validating upgrade server types: %s", c.UpgradeServerType))
		upgradeServerType, _, err := client.ServerType.Get(ctx, c.UpgradeServerType)
		if err != nil {
			return errorHandler(state, ui, fmt.Sprintf("Could not fetch upgrade server type '%s'", c.UpgradeServerType), err)
		}
		if upgradeServerType == nil {
			return errorHandler(state, ui, "", fmt.Errorf("Could not find upgrade server type '%s'", c.UpgradeServerType))
		}

		if serverType.Architecture != upgradeServerType.Architecture {
			// This is also validated by API, but if we validate it here, its faster and we never have to create
			// a server in the first place. Saving users to first hour of billing.
			return errorHandler(state, ui, "", fmt.Errorf("server_type and upgrade_server_type have incompatible architectures"))
		}
	}

	// Skip snapshot name validation if skip_create_snapshot is set to true.
	if c.SkipCreateSnapshot {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Validating snapshot name: %s", s.SnapshotName))

	// We would like to ask only for snapshots with a certain name using
	// ImageListOpts{Name: s.SnapshotName}, but snapshots do not have name, they
	// only have description. Thus we are obliged to ask for _all_ the snapshots.
	opts := hcloud.ImageListOpts{
		Type:         []hcloud.ImageType{hcloud.ImageTypeSnapshot},
		Architecture: []hcloud.Architecture{serverType.Architecture},
	}
	snapshots, err := client.Image.AllWithOpts(ctx, opts)
	if err != nil {
		return errorHandler(state, ui, "Could not fetch snapshots", err)
	}

	for _, snap := range snapshots {
		if snap.Description == s.SnapshotName {
			msg := fmt.Sprintf(
				"Found existing snapshot (id=%d, arch=%s) with name '%s'",
				snap.ID,
				serverType.Architecture,
				s.SnapshotName,
			)
			if s.Force {
				ui.Say(msg + ". Force flag specified, will safely overwrite this snapshot")
				state.Put(StateSnapshotIDOld, snap.ID)
				return multistep.ActionContinue
			}
			return errorHandler(state, ui, "", errors.New(msg))
		}
	}

	// no snapshot with the same name found
	return multistep.ActionContinue
}

// No-op
func (s *stepPreValidate) Cleanup(multistep.StateBag) {
}
