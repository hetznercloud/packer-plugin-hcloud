// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

func TestStepCreateServer(t *testing.T) {
	RunStepTestCases(t, []StepTestCase{
		{
			Name: "happy",
			Step: &stepCreateServer{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateSSHKeyID, int64(1))
			},
			WantRequests: []Request{
				{"GET", "/ssh_keys/1", nil,
					200, `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{"POST", "/servers",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerCreateRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-server", payload.Name)
						assert.Equal(t, "debian-12", payload.Image)
						assert.Equal(t, "nbg1", payload.Location)
						assert.Equal(t, "cpx11", payload.ServerType)
						assert.Nil(t, payload.Networks)
						assert.Nil(t, payload.PublicNet)
					},
					201, `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "1.2.3.4" }}},
						"action": { "id": 3, "status": "progress" }
					}`,
				},
				{"GET", "/actions/3", nil,
					200, `{
						"action": { "id": 3, "status": "success" }
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverID, ok := state.Get(StateServerID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(8), serverID)

				instanceID, ok := state.Get(StateInstanceID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(8), instanceID)

				serverIP, ok := state.Get(StateServerIP).(string)
				assert.True(t, ok)
				assert.Equal(t, "1.2.3.4", serverIP)
			},
		},
		{
			Name: "happy with network",
			Step: &stepCreateServer{},
			SetupConfigFunc: func(c *Config) {
				c.Networks = []int64{12}
			},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateSSHKeyID, int64(1))
			},
			WantRequests: []Request{
				{"GET", "/ssh_keys/1", nil,
					200, `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{"POST", "/servers",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerCreateRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-server", payload.Name)
						assert.Equal(t, "debian-12", payload.Image)
						assert.Equal(t, "nbg1", payload.Location)
						assert.Equal(t, "cpx11", payload.ServerType)
						assert.Equal(t, []int64{12}, payload.Networks)
					},
					201, `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "1.2.3.4" }}},
						"action": { "id": 3, "status": "progress" }
					}`,
				},
				{"GET", "/actions/3", nil,
					200, `{
						"action": { "id": 3, "status": "success" }
					}`,
				},
			},

			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverID, ok := state.Get(StateServerID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(8), serverID)

				instanceID, ok := state.Get(StateInstanceID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(8), instanceID)

				serverIP, ok := state.Get(StateServerIP).(string)
				assert.True(t, ok)
				assert.Equal(t, "1.2.3.4", serverIP)
			},
		},
		{
			Name: "happy with public ipv4",
			Step: &stepCreateServer{},
			SetupConfigFunc: func(c *Config) {
				c.PublicIPv4 = "127.0.0.1"
			},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateSSHKeyID, int64(1))
			},
			WantRequests: []Request{
				{"GET", "/ssh_keys/1", nil,
					200, `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{"GET", "/primary_ips?ip=127.0.0.1", nil,
					200, `{
						"primary_ips": [
							{
								"id": 1,
								"ip": "127.0.0.1",
								"type": "ipv4"
							}
						]
					}`,
				},
				{"POST", "/servers",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerCreateRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-server", payload.Name)
						assert.Equal(t, "debian-12", payload.Image)
						assert.Equal(t, "nbg1", payload.Location)
						assert.Equal(t, "cpx11", payload.ServerType)
						assert.Nil(t, payload.Networks)
						assert.NotNil(t, payload.PublicNet)
						assert.Equal(t, int64(1), payload.PublicNet.IPv4ID)
					},
					201, `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "127.0.0.1" }}},
						"action": { "id": 3, "status": "progress" }
					}`,
				},
				{"GET", "/actions/3", nil,
					200, `{
						"action": { "id": 3, "status": "success" }
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverID, ok := state.Get(StateServerID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(8), serverID)

				instanceID, ok := state.Get(StateInstanceID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(8), instanceID)

				serverIP, ok := state.Get(StateServerIP).(string)
				assert.True(t, ok)
				assert.Equal(t, "127.0.0.1", serverIP)
			},
		},
	})
}
