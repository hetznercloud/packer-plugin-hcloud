// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type stepShutdownServer struct{}

//nolint:gosimple,goimports
func (s *stepShutdownServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	serverID := state.Get("server_id").(int64)

	ui.Say("Shutting down server...")

	action, _, err := client.Server.Shutdown(ctx, &hcloud.Server{ID: serverID})

	if err != nil {
		return errorHandler(state, ui, "Error stopping server", err)
	}

	_, errCh := client.Action.WatchProgress(ctx, action)
	for {
		select {
		case err1 := <-errCh:
			if err1 == nil {
				return multistep.ActionContinue
			} else {
				return errorHandler(state, ui, "Error stopping server", err)
			}
		}
	}
}

func (s *stepShutdownServer) Cleanup(state multistep.StateBag) {
	// no cleanup
}
