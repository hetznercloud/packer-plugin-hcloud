// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type stepCreateServer struct {
	serverId int64
}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c, ui, client := UnpackState(state)

	sshKeyId := state.Get(StateSSHKeyID).(int64)

	// Create the server based on configuration
	ui.Say("Creating server...")

	userData := c.UserData
	if c.UserDataFile != "" {
		contents, err := os.ReadFile(c.UserDataFile)
		if err != nil {
			return errorHandler(state, ui, "Could not read user data file", err)
		}

		userData = string(contents)
	}

	sshKeys := []*hcloud.SSHKey{{ID: sshKeyId}}
	for _, k := range c.SSHKeys {
		sshKey, _, err := client.SSHKey.Get(ctx, k)
		if err != nil {
			return errorHandler(state, ui, fmt.Sprintf("Could not fetch SSH key '%s'", k), err)
		}
		if sshKey == nil {
			return errorHandler(state, ui, fmt.Sprintf("Could not find SSH key '%s'", k), err)
		}
		sshKeys = append(sshKeys, sshKey)
	}

	var image *hcloud.Image
	if c.Image != "" {
		image = &hcloud.Image{Name: c.Image}
	} else {
		serverType := state.Get(StateServerType).(*hcloud.ServerType)
		var err error
		image, err = getImageWithSelectors(ctx, client, c, serverType)
		if err != nil {
			return errorHandler(state, ui, "Could not find image", err)
		}
		ui.Message(fmt.Sprintf("Using image %s with ID %d", image.Description, image.ID))
	}

	var networks []*hcloud.Network
	for _, k := range c.Networks {
		networks = append(networks, &hcloud.Network{ID: k})
	}

	serverCreateOpts := hcloud.ServerCreateOpts{
		Name:       c.ServerName,
		ServerType: &hcloud.ServerType{Name: c.ServerType},
		Image:      image,
		SSHKeys:    sshKeys,
		Location:   &hcloud.Location{Name: c.Location},
		UserData:   userData,
		Networks:   networks,
		Labels:     c.ServerLabels,
	}

	if c.UpgradeServerType != "" {
		serverCreateOpts.StartAfterCreate = hcloud.Ptr(false)
	}

	serverCreateResult, _, err := client.Server.Create(ctx, serverCreateOpts)
	if err != nil {
		return errorHandler(state, ui, "Could not create server", err)
	}
	state.Put(StateServerIP, serverCreateResult.Server.PublicNet.IPv4.IP.String())
	// We use this in cleanup
	s.serverId = serverCreateResult.Server.ID

	// Store the server id for later
	state.Put(StateServerID, serverCreateResult.Server.ID)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put(StateInstanceID, serverCreateResult.Server.ID)

	if err := waitForAction(ctx, client, serverCreateResult.Action); err != nil {
		return errorHandler(state, ui, "Could not create server", err)
	}
	for _, nextAction := range serverCreateResult.NextActions {
		if err := waitForAction(ctx, client, nextAction); err != nil {
			return errorHandler(state, ui, "Could not create server", err)
		}
	}

	if c.UpgradeServerType != "" {
		ui.Say("Upgrading server type...")
		serverChangeTypeAction, _, err := client.Server.ChangeType(ctx, serverCreateResult.Server, hcloud.ServerChangeTypeOpts{
			ServerType:  &hcloud.ServerType{Name: c.UpgradeServerType},
			UpgradeDisk: false,
		})
		if err != nil {
			return errorHandler(state, ui, "Could not upgrade server type", err)
		}

		if err := waitForAction(ctx, client, serverChangeTypeAction); err != nil {
			return errorHandler(state, ui, "Could not upgrade server type", err)
		}

		ui.Say("Starting server...")
		serverPoweronAction, _, err := client.Server.Poweron(ctx, serverCreateResult.Server)
		if err != nil {
			return errorHandler(state, ui, "Could not start server", err)
		}

		if err := waitForAction(ctx, client, serverPoweronAction); err != nil {
			return errorHandler(state, ui, "Could not start server", err)
		}
	}

	if c.RescueMode != "" {
		ui.Say("Enabling Rescue Mode...")
		_, err := setRescue(ctx, client, serverCreateResult.Server, c.RescueMode, sshKeys)
		if err != nil {
			return errorHandler(state, ui, "Could not enable rescue mode", err)
		}
		ui.Say("Rebooting server...")
		action, _, err := client.Server.Reset(ctx, serverCreateResult.Server)
		if err != nil {
			return errorHandler(state, ui, "Could not reboot server", err)
		}
		if err := waitForAction(ctx, client, action); err != nil {
			return errorHandler(state, ui, "Could not reboot server", err)
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	// If the serverID isn't there, we probably never created it
	if s.serverId == 0 {
		return
	}

	_, ui, client := UnpackState(state)

	// Destroy the server we just created
	ui.Say("Destroying server...")
	_, _, err := client.Server.DeleteWithResult(context.TODO(), &hcloud.Server{ID: s.serverId})
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying server. Please destroy it manually: %s", err))
	}
}

func setRescue(ctx context.Context, client *hcloud.Client, server *hcloud.Server, rescue string, sshKeys []*hcloud.SSHKey) (string, error) {
	rescueChanged := false
	if server.RescueEnabled {
		rescueChanged = true
		action, _, err := client.Server.DisableRescue(ctx, server)
		if err != nil {
			return "", err
		}
		if err := waitForAction(ctx, client, action); err != nil {
			return "", err
		}
	}

	if rescue != "" {
		res, _, err := client.Server.EnableRescue(ctx, server, hcloud.ServerEnableRescueOpts{
			Type:    hcloud.ServerRescueType(rescue),
			SSHKeys: sshKeys,
		})
		if err != nil {
			return "", err
		}
		if err := waitForAction(ctx, client, res.Action); err != nil {
			return "", err
		}
		return res.RootPassword, nil
	}

	if rescueChanged {
		action, _, err := client.Server.Reset(ctx, server)
		if err != nil {
			return "", err
		}
		if err := waitForAction(ctx, client, action); err != nil {
			return "", err
		}
	}
	return "", nil
}

func waitForAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action) error {
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

func getImageWithSelectors(ctx context.Context, client *hcloud.Client, c *Config, serverType *hcloud.ServerType) (*hcloud.Image, error) {
	var allImages []*hcloud.Image

	selector := strings.Join(c.ImageFilter.WithSelector, ",")
	opts := hcloud.ImageListOpts{
		ListOpts:     hcloud.ListOpts{LabelSelector: selector},
		Status:       []hcloud.ImageStatus{hcloud.ImageStatusAvailable},
		Architecture: []hcloud.Architecture{serverType.Architecture},
	}

	allImages, err := client.Image.AllWithOpts(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(allImages) == 0 {
		return nil, fmt.Errorf("no image found for selector %q", selector)
	}
	if len(allImages) > 1 {
		if !c.ImageFilter.MostRecent {
			return nil, fmt.Errorf("more than one image found for selector %q", selector)
		}

		sort.Slice(allImages, func(i, j int) bool {
			return allImages[i].Created.After(allImages[j].Created)
		})
	}

	return allImages[0], nil
}
