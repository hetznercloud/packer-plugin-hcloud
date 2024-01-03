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

const OldSnapshotID = "old_snapshot_id"

type stepCreateSnapshot struct{}

//nolint:gosimple,goimports
func (s *stepCreateSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	serverID := state.Get("server_id").(int64)

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
	state.Put("snapshot_id", result.Image.ID)
	state.Put("snapshot_name", c.SnapshotName)
	_, errCh := client.Action.WatchProgress(ctx, result.Action)

	err1 := <-errCh
	if err1 != nil {
		return errorHandler(state, ui, "Could not create snapshot", err)
	}

	oldSnap, found := state.GetOk(OldSnapshotID)
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
