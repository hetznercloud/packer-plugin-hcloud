// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type stepCreateSSHKey struct {
	keyId int64
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	ui.Say("Uploading temporary SSH key for instance...")

	if c.Comm.SSHPublicKey == nil {
		return errorHandler(state, ui, "", fmt.Errorf("missing SSH public key in communicator"))
	}

	// The name of the public key on the Hetzner Cloud
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	key, _, err := client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      name,
		PublicKey: string(c.Comm.SSHPublicKey),
		Labels:    c.SSHKeysLabels,
	})
	if err != nil {
		return errorHandler(state, ui, "Could not upload temporary SSH key", err)
	}

	// We use this to check cleanup
	s.keyId = key.ID

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state.Put("ssh_key_id", key.ID)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	// If no key id is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting temporary SSH key...")
	_, err := client.SSHKey.Delete(context.TODO(), &hcloud.SSHKey{ID: s.keyId})
	if err != nil {
		errorHandler(state, ui, "Could not cleanup temporary SSH key", err)
	}
}
