// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

func TestStepPreValidate(t *testing.T) {
	fakeSnapNames := []string{"snapshot-old"}

	testCases := []struct {
		name string
		// zero value: assert that state OldSnapshotID is NOT present
		// non-zero value: assert that state OldSnapshotID is present AND has this value
		wantOldSnapID int
		step          stepPreValidate
		wantAction    multistep.StepAction
	}{
		{
			name:       "snapshot name new, success",
			step:       stepPreValidate{SnapshotName: "snapshot-new"},
			wantAction: multistep.ActionContinue,
		},
		{
			name:       "snapshot name old, failure",
			step:       stepPreValidate{SnapshotName: "snapshot-old"},
			wantAction: multistep.ActionHalt,
		},
		{
			name:          "snapshot name old, force flag, success",
			wantOldSnapID: 1000,
			step:          stepPreValidate{SnapshotName: "snapshot-old", Force: true},
			wantAction:    multistep.ActionContinue,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := make(chan error, 1)
			state, teardown := setupStepPreValidate(errors, fakeSnapNames)
			defer teardown()

			if testing.Verbose() {
				state.Put("ui", packersdk.TestUi(t))
			} else {
				// do not output to stdout or console
				state.Put("ui", &packersdk.MockUi{})
			}

			if action := tc.step.Run(context.Background(), state); action != tc.wantAction {
				t.Errorf("step.Run: want: %v; got: %v", tc.wantAction, action)
			}

			oldSnap, found := state.GetOk(OldSnapshotID)
			if found {
				oldSnapID := oldSnap.(int)
				if tc.wantOldSnapID == 0 {
					t.Errorf("OldSnapshotID: got: present with value %d; want: not present", oldSnapID)
				} else if oldSnapID != tc.wantOldSnapID {
					t.Errorf("OldSnapshotID: got: %d; want: %d", oldSnapID, tc.wantOldSnapID)
				}
			} else if tc.wantOldSnapID != 0 {
				t.Errorf("OldSnapshotID: got: not present; want: present, with value %d",
					tc.wantOldSnapID)
			}

			select {
			case err := <-errors:
				t.Errorf("server: got: %s", err)
			default:
			}
		})
	}
}

// Configure a httptest server to reply to the requests done by stepPrevalidate.
// Report errors on the errors channel (cannot use testing.T, it runs on a different goroutine).
// Return a tuple (state, teardown) where:
// - state (containing the client) is ready to be passed to the step.Run() method.
// - teardown is a function meant to be deferred from the test.
func setupStepPreValidate(errors chan<- error, fakeSnapNames []string) (*multistep.BasicStateBag, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			errors <- fmt.Errorf("fake server: reading request: %s", err)
			return
		}
		reqDump := fmt.Sprintf("fake server: request:\n    %s %s\n    body: %s",
			r.Method, r.URL.Path, string(buf))
		if testing.Verbose() {
			fmt.Println(reqDump)
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		var response interface{}

		if r.Method == http.MethodGet && r.URL.Path == "/images" {
			w.WriteHeader(http.StatusOK)
			images := make([]schema.Image, 0, len(fakeSnapNames))
			for i, fakeDesc := range fakeSnapNames {
				img := schema.Image{
					ID:          1000 + i,
					Type:        string(hcloud.ImageTypeSnapshot),
					Description: fakeDesc,
				}
				images = append(images, img)
			}
			response = &schema.ImageListResponse{Images: images}
		}

		if r.Method == http.MethodGet && r.URL.Path == "/server_types" && r.URL.RawQuery == "name=cx11" {
			w.WriteHeader(http.StatusOK)
			serverTypes := []schema.ServerType{{
				Name:         "cx11",
				Architecture: "x86",
			}}
			response = &schema.ServerTypeListResponse{ServerTypes: serverTypes}
		}

		if response != nil {
			if err := enc.Encode(response); err != nil {
				errors <- fmt.Errorf("fake server: encoding reply: %s", err)
			}
			return
		}

		// no match: report error
		w.WriteHeader(http.StatusBadRequest)
		errors <- fmt.Errorf(reqDump)
	}))

	state := multistep.BasicStateBag{}

	client := hcloud.NewClient(hcloud.WithEndpoint(ts.URL), hcloud.WithDebugWriter(os.Stderr))
	state.Put("hcloudClient", client)

	config := &Config{
		ServerType: "cx11",
	}
	state.Put("config", config)

	teardown := func() {
		ts.Close()
	}
	return &state, teardown
}
