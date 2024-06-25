// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"net/http"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/mockutils"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

func TestStepCreateServer(t *testing.T) {
	RunStepTestCases(t, []StepTestCase{
		{
			Name: "happy",
			Step: &stepCreateServer{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateSSHKeyID, int64(1))
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "POST", Path: "/servers",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerCreateRequest{})
						assert.Equal(t, "dummy-server", payload.Name)
						assert.Equal(t, int64(114690387), int64(payload.Image.(float64)))
						assert.Equal(t, "nbg1", payload.Location)
						assert.Equal(t, "cpx11", payload.ServerType)
						assert.True(t, payload.PublicNet.EnableIPv4)
						assert.True(t, payload.PublicNet.EnableIPv6)
						assert.Nil(t, payload.Networks)
					},
					Status: 201,
					JSONRaw: `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "1.2.3.4" }}},
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{Method: "GET", Path: "/actions?id=3&page=1&sort=status&sort=id",
					Status: 200,
					JSONRaw: `{
						"actions": [
							{ "id": 3, "status": "success" }
						],
						"meta": { "pagination": { "page": 1 }}
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
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "POST", Path: "/servers",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerCreateRequest{})
						assert.Equal(t, "dummy-server", payload.Name)
						assert.Equal(t, int64(114690387), int64(payload.Image.(float64)))
						assert.Equal(t, "nbg1", payload.Location)
						assert.Equal(t, "cpx11", payload.ServerType)
						assert.Equal(t, []int64{12}, payload.Networks)
					},
					Status: 201,
					JSONRaw: `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "1.2.3.4" }}},
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{Method: "GET", Path: "/actions?id=3&page=1&sort=status&sort=id",
					Status: 200,
					JSONRaw: `{
						"actions": [
							{ "id": 3, "status": "success" }
						],
						"meta": { "pagination": { "page": 1 }}
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
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "GET", Path: "/primary_ips?name=permanent-packer-ipv4",
					Status: 200,
					JSONRaw: `{
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
				{Method: "GET", Path: "/primary_ips?name=permanent-packer-ipv6",
					Status: 200,
					JSONRaw: `{
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
				{Method: "POST", Path: "/servers",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerCreateRequest{})
						assert.Equal(t, "dummy-server", payload.Name)
						assert.Equal(t, int64(114690387), int64(payload.Image.(float64)))
						assert.Equal(t, "nbg1", payload.Location)
						assert.Equal(t, "cpx11", payload.ServerType)
						assert.Nil(t, payload.Networks)
						assert.NotNil(t, payload.PublicNet)
						assert.Equal(t, int64(1), payload.PublicNet.IPv4ID)
						assert.Equal(t, int64(2), payload.PublicNet.IPv6ID)
					},
					Status: 201,
					JSONRaw: `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "127.0.0.1" }, "ipv6": { "ip": "::1" }}},
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{Method: "GET", Path: "/actions?id=3&page=1&sort=status&sort=id",
					Status: 200,
					JSONRaw: `{
						"actions": [
							{ "id": 3, "status": "success" }
						],
						"meta": { "pagination": { "page": 1 }}
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
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "GET", Path: "/primary_ips?name=127.0.0.1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
				},
				{Method: "GET", Path: "/primary_ips?ip=127.0.0.1",
					Status: 200,
					JSONRaw: `{
						"primary_ips": [
							{
								"id": 1,
								"ip": "127.0.0.1",
								"type": "ipv4"
							}
						]
					}`,
				},
				{Method: "GET", Path: "/primary_ips?name=%3A%3A1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
				},
				{Method: "GET", Path: "/primary_ips?ip=%3A%3A1",
					Status: 200,
					JSONRaw: `{
						"primary_ips": [
							{
								"id": 2,
								"ip": "::1",
								"type": "ipv6"
							}
						]
					}`,
				},
				{Method: "POST", Path: "/servers",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerCreateRequest{})
						assert.Equal(t, "dummy-server", payload.Name)
						assert.Equal(t, int64(114690387), int64(payload.Image.(float64)))
						assert.Equal(t, "nbg1", payload.Location)
						assert.Equal(t, "cpx11", payload.ServerType)
						assert.Nil(t, payload.Networks)
						assert.NotNil(t, payload.PublicNet)
						assert.Equal(t, int64(1), payload.PublicNet.IPv4ID)
						assert.Equal(t, int64(2), payload.PublicNet.IPv6ID)
					},
					Status: 201,
					JSONRaw: `{
						"server": { "id": 8, "name": "dummy-server", "public_net": { "ipv4": { "ip": "127.0.0.1" }, "ipv6": { "ip": "::1" }}},
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{Method: "GET", Path: "/actions?id=3&page=1&sort=status&sort=id",
					Status: 200,
					JSONRaw: `{
						"actions": [
							{ "id": 3, "status": "success" }
						],
						"meta": { "pagination": { "page": 1 }}
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
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "GET", Path: "/primary_ips?name=127.0.0.1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
				},
				{Method: "GET", Path: "/primary_ips?ip=127.0.0.1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
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
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "GET", Path: "/primary_ips?name=127.0.0.1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
				},
				{Method: "GET", Path: "/primary_ips?ip=127.0.0.1",
					Status: 500,
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
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "GET", Path: "/primary_ips?name=127.0.0.1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
				},
				{Method: "GET", Path: "/primary_ips?ip=127.0.0.1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
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
				state.Put(StateServerType, &hcloud.ServerType{ID: 9, Name: "cpx11", Architecture: "x86"})
			},
			WantRequests: []mockutils.Request{
				{Method: "GET", Path: "/ssh_keys/1",
					Status: 200,
					JSONRaw: `{
						"ssh_key": { "id": 1 }
					}`,
				},
				{Method: "GET", Path: "/images?architecture=x86&include_deprecated=true&name=debian-12",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 114690387, "name": "debian-12", "description": "Debian 12", "architecture": "x86" }]
					}`,
				},
				{Method: "GET", Path: "/primary_ips?name=127.0.0.1",
					Status:  200,
					JSONRaw: `{ "primary_ips": [] }`,
				},
				{Method: "GET", Path: "/primary_ips?ip=127.0.0.1",
					Status: 500,
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

func TestFirstAvailableIP(t *testing.T) {
	testCases := []struct {
		name   string
		server *hcloud.Server
		want   string
	}{
		{
			name:   "empty",
			server: &hcloud.Server{},
			want:   "",
		},
		{
			name: "public_ipv4",
			server: &hcloud.Server{
				PublicNet: hcloud.ServerPublicNetFromSchema(schema.ServerPublicNet{
					IPv4: schema.ServerPublicNetIPv4{ID: 1, IP: "1.2.3.4"},
					IPv6: schema.ServerPublicNetIPv6{ID: 2, IP: "2a01:4f8:1c19:1403::/64"},
				}),
				PrivateNet: []hcloud.ServerPrivateNet{
					hcloud.ServerPrivateNetFromSchema(schema.ServerPrivateNet{Network: 3, IP: "10.0.0.1"}),
				},
			},
			want: "1.2.3.4",
		},
		{
			name: "public_ipv6",
			server: &hcloud.Server{
				PublicNet: hcloud.ServerPublicNetFromSchema(schema.ServerPublicNet{
					IPv6: schema.ServerPublicNetIPv6{ID: 2, IP: "2a01:4f8:1c19:1403::/64"},
				}),
				PrivateNet: []hcloud.ServerPrivateNet{
					hcloud.ServerPrivateNetFromSchema(schema.ServerPrivateNet{Network: 3, IP: "10.0.0.1"}),
				},
			},
			want: "2a01:4f8:1c19:1403::1",
		},
		{
			name: "private_ipv4",
			server: &hcloud.Server{
				PrivateNet: []hcloud.ServerPrivateNet{
					hcloud.ServerPrivateNetFromSchema(schema.ServerPrivateNet{Network: 3, IP: "10.0.0.1"}),
				},
			},
			want: "10.0.0.1",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := firstAvailableIP(testCase.server)
			assert.Equal(t, testCase.want, result)
		})
	}
}
