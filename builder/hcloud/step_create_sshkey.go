package hcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type stepCreateSSHKey struct {
	Comm  *communicator.Config
	keyId int
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

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
		return multistep.ActionContinue
	}

	// Create the key!
	key, _, err := client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      name,
		PublicKey: string(c.Comm.SSHPublicKey),
	})
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyId = key.ID

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state.Put("ssh_key_id", s.keyId)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	c := state.Get("config").(*Config)
	// If no key id or a SSHKey Pair is set, then we never created it, so just return
	if s.keyId == 0 || c.Comm.SSHKeyPairName != "" {

		return
	}

	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting temporary ssh key...")
	_, err := client.SSHKey.Delete(context.TODO(), &hcloud.SSHKey{ID: s.keyId})
	if err != nil {
		log.Printf("Error cleaning up ssh key: %s", err)
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %s", err))
	}
}
