// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/mockutil"
)

func TestStepPreValidate(t *testing.T) {
	RunStepTestCases(t, []StepTestCase{
		{
			Name: "happy",
			Step: &stepPreValidate{
				SnapshotName: "dummy-snapshot",
				Force:        false,
			},
			SetupConfigFunc: func(c *Config) {
				c.UpgradeServerType = "cpx32"
			},
			WantRequests: []mockutil.Request{
				{
					Method: "GET", Path: "/server_types?name=cpx22",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 109, "name": "cpx22", "architecture": "x86"}]
					}`,
				},
				{
					Method: "GET", Path: "/server_types?name=cpx32",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 110, "name": "cpx32", "architecture": "x86"}]
					}`,
				},
				{
					Method: "GET", Path: "/images?architecture=x86&page=1&type=snapshot",
					Status: 200,
					JSONRaw: `{
						"images": []
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverType, ok := state.Get(StateServerType).(*hcloud.ServerType)
				assert.True(t, ok)
				assert.Equal(t, hcloud.ServerType{ID: 109, Name: "cpx22", Architecture: "x86"}, *serverType)

				_, ok = state.Get(StateSnapshotIDOld).(int64)
				assert.False(t, ok)
			},
		},
		{
			Name: "fail with existing snapshot",
			Step: &stepPreValidate{
				SnapshotName: "dummy-snapshot",
				Force:        false,
			},
			SetupConfigFunc: func(c *Config) {
				c.UpgradeServerType = "cpx32"
			},
			WantRequests: []mockutil.Request{
				{
					Method: "GET", Path: "/server_types?name=cpx22",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 109, "name": "cpx22", "architecture": "x86"}]
					}`,
				},
				{
					Method: "GET", Path: "/server_types?name=cpx32",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 110, "name": "cpx32", "architecture": "x86"}]
					}`,
				},
				{
					Method: "GET", Path: "/images?architecture=x86&page=1&type=snapshot",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 1, "description": "dummy-snapshot"}]
					}`,
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverType, ok := state.Get(StateServerType).(*hcloud.ServerType)
				assert.True(t, ok)
				assert.Equal(t, hcloud.ServerType{ID: 109, Name: "cpx22", Architecture: "x86"}, *serverType)

				_, ok = state.Get(StateSnapshotIDOld).(int64)
				assert.False(t, ok)

				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.Error(t, err)
				assert.Equal(t, "Found existing snapshot (id=1, arch=x86) with name 'dummy-snapshot'", err.Error())
			},
		},
		{
			Name: "happy with existing snapshot",
			Step: &stepPreValidate{
				SnapshotName: "dummy-snapshot",
				Force:        true,
			},
			SetupConfigFunc: func(c *Config) {
				c.UpgradeServerType = "cpx32"
			},
			WantRequests: []mockutil.Request{
				{
					Method: "GET", Path: "/server_types?name=cpx22",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 109, "name": "cpx22", "architecture": "x86"}]
					}`,
				},
				{
					Method: "GET", Path: "/server_types?name=cpx32",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 110, "name": "cpx32", "architecture": "x86"}]
					}`,
				},
				{
					Method: "GET", Path: "/images?architecture=x86&page=1&type=snapshot",
					Status: 200,
					JSONRaw: `{
						"images": [{ "id": 1, "description": "dummy-snapshot"}]
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverType, ok := state.Get(StateServerType).(*hcloud.ServerType)
				assert.True(t, ok)
				assert.Equal(t, hcloud.ServerType{ID: 109, Name: "cpx22", Architecture: "x86"}, *serverType)

				snapshotIDOld, ok := state.Get(StateSnapshotIDOld).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(1), snapshotIDOld)
			},
		},
		{
			Name: "skip snapshot name validation",
			Step: &stepPreValidate{
				SnapshotName: "dummy-snapshot",
			},
			SetupConfigFunc: func(c *Config) {
				c.UpgradeServerType = "cpx32"
				c.SkipCreateSnapshot = true
			},
			WantRequests: []mockutil.Request{
				{
					Method: "GET", Path: "/server_types?name=cpx22",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 109, "name": "cpx22", "architecture": "x86"}]
					}`,
				},
				{
					Method: "GET", Path: "/server_types?name=cpx32",
					Status: 200,
					JSONRaw: `{
						"server_types": [{ "id": 110, "name": "cpx32", "architecture": "x86"}]
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverType, ok := state.Get(StateServerType).(*hcloud.ServerType)
				assert.True(t, ok)
				assert.Equal(t, hcloud.ServerType{ID: 109, Name: "cpx22", Architecture: "x86"}, *serverType)
			},
		},
	})
}
