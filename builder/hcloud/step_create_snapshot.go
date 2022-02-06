package hcloud

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

const OldSnapshotID = "old_snapshot_id"

type stepCreateSnapshot struct{}

//nolint: gosimple
func (s *stepCreateSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	serverID := state.Get("server_id").(int)

	if c.SkipImageCreation == true {
		ui.Say("Skip creation of snapshot ...")
		return multistep.ActionContinue
	}

	if c.MaxSnapshots != 0 {
		opts := hcloud.ImageListOpts{Type: []hcloud.ImageType{hcloud.ImageTypeSnapshot}}
		snapshots, err := client.Image.AllWithOpts(ctx, opts)
		if err != nil {
			err := fmt.Errorf("Error: getting snapshot list: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		var snap_list []*hcloud.Image
		for _, snap := range snapshots {
			if (reflect.DeepEqual(snap.Labels, c.SnapshotLabels)) == true {
				snap_list = append(snap_list, snap)

			}
		}
		// sort Label from latest to oldest, keep number of snapshots based on value of MaxSnapshots
		// it will delete older snapshots of matching label - so take care
		sort.Slice(snap_list, func(i, j int) bool {
			return snap_list[i].Created.After(snap_list[j].Created)
		})
		for i := c.MaxSnapshots; i < len(snap_list); i++ {
			snap := snap_list[i]

			ui.Say(fmt.Sprintf("Deleting old snapshot with ID: %d", snap.ID))
			image := &hcloud.Image{ID: snap.ID}
			_, err = client.Image.Delete(ctx, image)
			if err != nil {
				err := fmt.Errorf("Error deleting old snapshot with ID: %d: %s", snap.ID, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	ui.Say("Creating snapshot ...")
	ui.Say("This can take some time")
	result, _, err := client.Server.CreateImage(ctx, &hcloud.Server{ID: serverID}, &hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Labels:      c.SnapshotLabels,
		Description: hcloud.String(c.SnapshotName),
	})
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("snapshot_id", result.Image.ID)
	state.Put("snapshot_name", c.SnapshotName)
	_, errCh := client.Action.WatchProgress(ctx, result.Action)

	err1 := <-errCh
	if err1 != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err1)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	oldSnap, found := state.GetOk(OldSnapshotID)
	if !found {
		return multistep.ActionContinue
	}
	oldSnapID := oldSnap.(int)

	// The pre validate step has been invoked with -force AND found an existing
	// snapshot with the same name.
	// Now that we safely saved the new snapshot, let's delete the old one,
	// thus implementing an overwrite semantics.
	ui.Say(fmt.Sprintf("Deleting old snapshot with ID: %d", oldSnapID))
	image := &hcloud.Image{ID: oldSnapID}
	_, err = client.Image.Delete(ctx, image)
	if err != nil {
		err := fmt.Errorf("Error deleting old snapshot with ID: %d: %s", oldSnapID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepCreateSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}
