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
						"action": { "id": 3, "status": "running" }
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
						"action": { "id": 3, "status": "running" }
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
			Name: "happy with public ipv4 and ipv6 names",
			Step: &stepCreateServer{},
			SetupConfigFunc: func(c *Config) {
				c.PublicIPv4 = "permanent-packer-ipv4"
				c.PublicIPv6 = "permanent-packer-ipv6"
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
				{"GET", "/primary_ips?name=permanent-packer-ipv4", nil,
					200, `{
						"primary_ips": [
							{
								"name": "permanent-packer-ipv4",
								"id": 1,
								"ip": "127.0.0.1",
								"type": "ipv4"
							}
						]
					}`,
				},
				{"GET", "/primary_ips?name=permanent-packer-ipv6", nil,
					200, `{
						"primary_ips": [
							{
								"name": "permanent-packer-ipv6",
								"id": 2,
								"ip": "::1",
								"type": "ipv6"
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
						assert.Equal(t, int64(2), payload.PublicNet.IPv6ID)
					},
					201, `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "127.0.0.1" }, "ipv6": { "ip": "::1" }}},
						"action": { "id": 3, "status": "running" }
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
		{
			Name: "happy with public ipv4 and ipv6 addresses",
			Step: &stepCreateServer{},
			SetupConfigFunc: func(c *Config) {
				c.PublicIPv4 = "127.0.0.1"
				c.PublicIPv6 = "::1"
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
				{"GET", "/primary_ips?name=127.0.0.1", nil,
					200, `{ "primary_ips": [] }`,
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
				{"GET", "/primary_ips?name=%3A%3A1", nil,
					200, `{ "primary_ips": [] }`,
				},
				{"GET", "/primary_ips?ip=%3A%3A1", nil,
					200, `{
						"primary_ips": [
							{
								"id": 2,
								"ip": "::1",
								"type": "ipv6"
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
						assert.Equal(t, int64(2), payload.PublicNet.IPv6ID)
					},
					201, `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "127.0.0.1" }, "ipv6": { "ip": "::1" }}},
						"action": { "id": 3, "status": "running" }
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
		{
			Name: "fail to get for primary ip by address",
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
				{"GET", "/primary_ips?name=127.0.0.1", nil,
					200, `{ "primary_ips": [] }`,
				},
				{"GET", "/primary_ips?ip=127.0.0.1", nil,
					200, `{ "primary_ips": [] }`,
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
				assert.Regexp(t, "Could not find primary ip .*", err.Error())
			},
		},
		{
			Name: "fail to search for primary ip by address",
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
				{"GET", "/primary_ips?name=127.0.0.1", nil,
					200, `{ "primary_ips": [] }`,
				},
				{"GET", "/primary_ips?ip=127.0.0.1", nil,
					500, `{}`,
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
				assert.Regexp(t, "Could not fetch primary ip .*", err.Error())
			},
		},
		{
			Name: "fail to get for primary ipv4 by address",
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
				{"GET", "/primary_ips?name=127.0.0.1", nil,
					200, `{ "primary_ips": [] }`,
				},
				{"GET", "/primary_ips?ip=127.0.0.1", nil,
					200, `{ "primary_ips": [] }`,
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
				assert.Regexp(t, "Could not find primary ip .*", err.Error())
			},
		},
		{
			Name: "fail to search for primary ipv4 by address",
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
				{"GET", "/primary_ips?name=127.0.0.1", nil,
					200, `{ "primary_ips": [] }`,
				},
				{"GET", "/primary_ips?ip=127.0.0.1", nil,
					500, `{}`,
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
				assert.Regexp(t, "Could not fetch primary ip .*", err.Error())
			},
		},
	})
}
