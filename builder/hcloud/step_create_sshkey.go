// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/uuid"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type stepCreateSSHKey struct {
	keyId int64
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c, ui, client := UnpackState(state)

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
	state.Put(StateSSHKeyID, key.ID)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	// If no key id is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	_, ui, client := UnpackState(state)

	ui.Say("Deleting temporary SSH key...")
	_, err := client.SSHKey.Delete(context.TODO(), &hcloud.SSHKey{ID: s.keyId})
	if err != nil {
		errorHandler(state, ui, "Could not cleanup temporary SSH key", err)
	}
}
