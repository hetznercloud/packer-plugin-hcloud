// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
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
				c.UpgradeServerType = "cx22"
			},
			WantRequests: []Request{
				{"GET", "/server_types?name=cx22", nil,
					200, `{
						"server_types": [{ "id": 9, "name": "cx22", "architecture": "x86"}]
					}`,
				},
				{"GET", "/server_types?name=cx22", nil,
					200, `{
						"server_types": [{ "id": 10, "name": "cx22", "architecture": "x86"}]
					}`,
				},
				{"GET", "/images?architecture=x86&page=1&type=snapshot", nil,
					200, `{
						"images": []
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverType, ok := state.Get(StateServerType).(*hcloud.ServerType)
				assert.True(t, ok)
				assert.Equal(t, hcloud.ServerType{ID: 9, Name: "cx22", Architecture: "x86"}, *serverType)

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
				c.UpgradeServerType = "cx22"
			},
			WantRequests: []Request{
				{"GET", "/server_types?name=cx22", nil,
					200, `{
						"server_types": [{ "id": 9, "name": "cx22", "architecture": "x86"}]
					}`,
				},
				{"GET", "/server_types?name=cx22", nil,
					200, `{
						"server_types": [{ "id": 10, "name": "cx22", "architecture": "x86"}]
					}`,
				},
				{"GET", "/images?architecture=x86&page=1&type=snapshot", nil,
					200, `{
						"images": [{ "id": 1, "description": "dummy-snapshot"}]
					}`,
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverType, ok := state.Get(StateServerType).(*hcloud.ServerType)
				assert.True(t, ok)
				assert.Equal(t, hcloud.ServerType{ID: 9, Name: "cx22", Architecture: "x86"}, *serverType)

				_, ok = state.Get(StateSnapshotIDOld).(int64)
				assert.False(t, ok)

				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
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
				c.UpgradeServerType = "cx22"
			},
			WantRequests: []Request{
				{"GET", "/server_types?name=cx22", nil,
					200, `{
						"server_types": [{ "id": 9, "name": "cx22", "architecture": "x86"}]
					}`,
				},
				{"GET", "/server_types?name=cx22", nil,
					200, `{
						"server_types": [{ "id": 10, "name": "cx22", "architecture": "x86"}]
					}`,
				},
				{"GET", "/images?architecture=x86&page=1&type=snapshot", nil,
					200, `{
						"images": [{ "id": 1, "description": "dummy-snapshot"}]
					}`,
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				serverType, ok := state.Get(StateServerType).(*hcloud.ServerType)
				assert.True(t, ok)
				assert.Equal(t, hcloud.ServerType{ID: 9, Name: "cx22", Architecture: "x86"}, *serverType)

				snapshotIDOld, ok := state.Get(StateSnapshotIDOld).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(1), snapshotIDOld)
			},
		},
	})
}
