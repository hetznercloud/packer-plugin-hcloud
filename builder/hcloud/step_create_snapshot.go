// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type stepCreateSnapshot struct{}

//nolint:gosimple,goimports
func (s *stepCreateSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c, ui, client := UnpackState(state)

	serverID := state.Get(StateServerID).(int64)

	// Skip snapshot creation if skip_create_snapshot is set to true.
	if c.SkipCreateSnapshot {
		ui.Say("Skipping snapshot creation...")
		return multistep.ActionContinue
	}

	ui.Say("Creating snapshot...")
	ui.Say("This can take some time")
	result, _, err := client.Server.CreateImage(ctx, &hcloud.Server{ID: serverID}, &hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Labels:      c.SnapshotLabels,
		Description: hcloud.Ptr(c.SnapshotName),
	})
	if err != nil {
		return errorHandler(state, ui, "Could not create snapshot", err)
	}
	state.Put(StateSnapshotID, result.Image.ID)
	state.Put(StateSnapshotName, c.SnapshotName)

	if err := client.Action.WaitFor(ctx, result.Action); err != nil {
		return errorHandler(state, ui, "Could not create snapshot", err)
	}

	oldSnap, found := state.GetOk(StateSnapshotIDOld)
	if !found {
		return multistep.ActionContinue
	}
	oldSnapID := oldSnap.(int64)

	// The pre validate step has been invoked with -force AND found an existing
	// snapshot with the same name.
	// Now that we safely saved the new snapshot, let's delete the old one,
	// thus implementing an overwrite semantics.
	ui.Say(fmt.Sprintf("Deleting old snapshot with ID: %d", oldSnapID))
	image := &hcloud.Image{ID: oldSnapID}
	_, err = client.Image.Delete(ctx, image)
	if err != nil {
		return errorHandler(state, ui, fmt.Sprintf("Could not delete old snapshot id=%d", oldSnapID), err)
	}
	return multistep.ActionContinue
}

func (s *stepCreateSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}
