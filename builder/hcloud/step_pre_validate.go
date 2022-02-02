package hcloud

import (
	"context"
	"fmt"
	"strconv"

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
	keyId        int
	net_id       int
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

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
			snapMsg := fmt.Sprintf("snapshot name: '%s' is used by existing snapshot with ID %d",
				s.SnapshotName, snap.ID)
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

	if c.Comm.SSHKeyPairName != "" && c.Comm.SSHPrivateKeyFile != "" {
		sshKey, _, err := client.SSHKey.Get(ctx, c.Comm.SSHKeyPairName)
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", fmt.Errorf("Error fetching SSH key: %s", err))
			return multistep.ActionHalt
		}
		if sshKey == nil {
			state.Put("error", fmt.Errorf("Could not find key: %s", c.Comm.SSHKeyPairName))
			return multistep.ActionHalt
		}
		s.keyId = sshKey.ID
		state.Put("ssh_key_id", s.keyId)
	}

	// check if input of Network matches existing network - output network_id
	if c.Network != "" {
		// check if input can be converted to string
		network_id, err := strconv.Atoi(c.Network)
		// Lookup by Name
		if err != nil {
			network, _, err := client.Network.GetByName(ctx, c.Network)
			if err != nil {
				ui.Error(err.Error())
				state.Put("error", fmt.Errorf("Error fetching Network ID: %s", err))
				return multistep.ActionHalt
			}
			if network == nil {
				state.Put("error", fmt.Errorf("Could not find network: %s", err))
				return multistep.ActionHalt
			}
			s.net_id = network.ID
			state.Put("network_id", s.net_id)
		}
		// Lookup by ID
		if err == nil {
			network, _, err := client.Network.GetByID(ctx, network_id)

			if err != nil {
				ui.Error(err.Error())
				state.Put("error", fmt.Errorf("Error fetching Network ID: %s", err))
				return multistep.ActionHalt
			}
			if network == nil {
				state.Put("error", fmt.Errorf("Could not find network: %s", err))
				return multistep.ActionHalt
			}
			s.net_id = network.ID
			state.Put("network_id", s.net_id)
		}

	}
	return multistep.ActionContinue
}

// No-op
func (s *stepPreValidate) Cleanup(multistep.StateBag) {
}
