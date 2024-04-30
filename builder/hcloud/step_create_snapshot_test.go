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

func TestStepCreateSnapshot(t *testing.T) {
	RunStepTestCases(t, []StepTestCase{
		{
			Name: "happy",
			Step: &stepCreateSnapshot{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateServerID, int64(8))
			},
			WantRequests: []Request{
				{"POST", "/servers/8/actions/create_image",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerActionCreateImageRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					201, `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
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
				snapshotID, ok := state.Get(StateSnapshotID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(16), snapshotID)

				snapshotName, ok := state.Get(StateSnapshotName).(string)
				assert.True(t, ok)
				assert.Equal(t, "dummy-snapshot", snapshotName)
			},
		},
		{
			Name: "fail create image",
			Step: &stepCreateSnapshot{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateServerID, int64(8))
			},
			WantRequests: []Request{
				{"POST", "/servers/8/actions/create_image",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerActionCreateImageRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					400, "",
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
				assert.Regexp(t, "Could not create snapshot: .*", err.Error())
			},
		},
		{
			Name: "fail action",
			Step: &stepCreateSnapshot{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateServerID, int64(8))
			},
			WantRequests: []Request{
				{"POST", "/servers/8/actions/create_image",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerActionCreateImageRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					201, `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{"GET", "/actions/3", nil,
					200, `{
						"action": {
							"id": 3,
							"status": "error",
							"error": {
								"code": "action_failed", 
								"message": "Action failed"
							}
						}
					}`,
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
				assert.Regexp(t, "Could not create snapshot: .*", err.Error())
			},
		},
		{
			Name: "happy with old snapshot",
			Step: &stepCreateSnapshot{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateServerID, int64(8))
				state.Put(StateSnapshotIDOld, int64(20))
			},
			WantRequests: []Request{
				{"POST", "/servers/8/actions/create_image",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerActionCreateImageRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					201, `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{"GET", "/actions/3", nil,
					200, `{
						"action": { "id": 3, "status": "success" }
					}`,
				},
				{"DELETE", "/images/20", nil,
					204, "",
				},
			},
			WantStepAction: multistep.ActionContinue,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				snapshotID, ok := state.Get(StateSnapshotID).(int64)
				assert.True(t, ok)
				assert.Equal(t, int64(16), snapshotID)

				snapshotName, ok := state.Get(StateSnapshotName).(string)
				assert.True(t, ok)
				assert.Equal(t, "dummy-snapshot", snapshotName)
			},
		},
		{
			Name: "fail with old snapshot",
			Step: &stepCreateSnapshot{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateServerID, int64(8))
				state.Put(StateSnapshotIDOld, int64(20))
			},
			WantRequests: []Request{
				{"POST", "/servers/8/actions/create_image",
					func(t *testing.T, r *http.Request, body []byte) {
						payload := schema.ServerActionCreateImageRequest{}
						assert.NoError(t, json.Unmarshal(body, &payload))
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					201, `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{"GET", "/actions/3", nil,
					200, `{
						"action": { "id": 3, "status": "success" }
					}`,
				},
				{"DELETE", "/images/20", nil,
					400, "",
				},
			},
			WantStepAction: multistep.ActionHalt,
			WantStateFunc: func(t *testing.T, state multistep.StateBag) {
				err, ok := state.Get(StateError).(error)
				assert.True(t, ok)
				assert.NotNil(t, err)
				assert.Regexp(t, "Could not delete old snapshot id=20: .*", err.Error())
			},
		},
	})
}
