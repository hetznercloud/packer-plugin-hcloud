// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"net/http"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/mockutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

func TestStepCreateSSHKey(t *testing.T) {
	RunStepTestCases(t, []StepTestCase{
		{
			Name: "happy",
			Step: &stepCreateSSHKey{},
			SetupConfigFunc: func(c *Config) {
				c.Comm.SSHPublicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILBN85MgkHac/Q+iyPS8+88eBDn2SEGnU4/uLvj6lbT0")
			},
			WantRequests: []mockutil.Request{
				{Method: "POST", Path: "/ssh_keys",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.SSHKeyCreateRequest{})
						assert.Regexp(t, "packer([a-z0-9-]+)$", payload.Name)
						assert.Equal(t, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILBN85MgkHac/Q+iyPS8+88eBDn2SEGnU4/uLvj6lbT0", payload.PublicKey)
					},
					Status: 201,
					JSONRaw: `{
						"ssh_key": {
							"id": 8,
							"name": "packer-659596d1-93df-3868-8170-42139065172e",
							"public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILBN85MgkHac/Q+iyPS8+88eBDn2SEGnU4/uLvj6lbT0"
						}
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				sshKeyID, ok := state.Get(StateSSHKeyID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(8), sshKeyID)
			},
		},
	})
}

func TestStepCleanupSSHKey(t *testing.T) {
	RunStepTestCases(t, []StepTestCase{
		{
			Name:         "happy",
			Step:         &stepCreateSSHKey{keyId: 1},
			StepFuncName: "cleanup",
			WantRequests: []mockutil.Request{
				{Method: "DELETE", Path: "/ssh_keys/1",
					Status: 204,
				},
			},
		},
	})
}
