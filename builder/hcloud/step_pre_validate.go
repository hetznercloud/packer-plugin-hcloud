package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type stepPreValidate struct {
	Force        bool
	SnapshotName string
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)

	if s.Force {
		ui.Say("Force flag found, skipping prevalidating snapshot name")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Prevalidating snapshot name: %s", s.SnapshotName))

	// We would like to ask only for snapshots with a certain name using
	// ImageListOpts{Name: s.SnapshotName}, but snapshots do not have name, they
	// only have description. Thus we are obliged to ask for _all_ the snapshots.
	opts := hcloud.ImageListOpts{Type: []hcloud.ImageType{hcloud.ImageTypeSnapshot}}
	snapshots, err := client.Image.AllWithOpts(ctx, opts)
	if err != nil {
		err := fmt.Errorf("Error: getting snapshot list: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, snap := range snapshots {
		if snap.Description == s.SnapshotName {
			err := fmt.Errorf("Error: snapshot name: '%s' is used by existing snapshot with ID %d",
				s.SnapshotName, snap.ID)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// No-op
func (s *stepPreValidate) Cleanup(multistep.StateBag) {
}
