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

func TestStepCreateSnapshot(t *testing.T) {
	RunStepTestCases(t, []StepTestCase{
		{
			Name: "happy",
			Step: &stepCreateSnapshot{},
			SetupStateFunc: func(state multistep.StateBag) {
				state.Put(StateServerID, int64(8))
			},
			WantRequests: []mockutil.Request{
				{Method: "POST", Path: "/servers/8/actions/create_image",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerActionCreateImageRequest{})
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					Status: 201,
					JSONRaw: `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
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
			WantRequests: []mockutil.Request{
				{Method: "POST", Path: "/servers/8/actions/create_image",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerActionCreateImageRequest{})
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					Status: 400,
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
			WantRequests: []mockutil.Request{
				{Method: "POST", Path: "/servers/8/actions/create_image",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerActionCreateImageRequest{})
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					Status: 201,
					JSONRaw: `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
						"action": { "id": 3, "status": "running" }
					}`,
				},
				{Method: "GET", Path: "/actions?id=3&page=1&sort=status&sort=id",
					Status: 200,
					JSONRaw: `{
						"actions": [
							{
								"id": 3,
								"status": "error",
								"error": {
									"code": "action_failed",
									"message": "Action failed"
								}
							}
						],
						"meta": { "pagination": { "page": 1 }}
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
			WantRequests: []mockutil.Request{
				{Method: "POST", Path: "/servers/8/actions/create_image",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerActionCreateImageRequest{})
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					Status: 201,
					JSONRaw: `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
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
				{Method: "DELETE", Path: "/images/20",
					Status: 204,
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
			WantRequests: []mockutil.Request{
				{Method: "POST", Path: "/servers/8/actions/create_image",
					Want: func(t *testing.T, req *http.Request) {
						payload := decodeJSONBody(t, req.Body, &schema.ServerActionCreateImageRequest{})
						assert.Equal(t, "dummy-snapshot", *payload.Description)
						assert.Equal(t, "snapshot", *payload.Type)
					},
					Status: 201,
					JSONRaw: `{
						"image": { "id": 16, "description": "dummy-snapshot", "type": "snapshot" },
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
				{Method: "DELETE", Path: "/images/20",
					Status: 400,
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
