package hcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

type FailCause int

const (
	Pass FailCause = iota
	FailCreateImage
	FailWatchProgress
)

func TestStepCreateSnapshot(t *testing.T) {
	testCases := []struct {
		name       string
		snapName   string
		failCause  FailCause // zero value: pass
		wantAction multistep.StepAction
	}{
		{
			name:       "happy path",
			snapName:   "dummy-snap",
			wantAction: multistep.ActionContinue,
		},
		{
			name:       "fail create image",
			snapName:   "dummy-snap",
			failCause:  FailCreateImage,
			wantAction: multistep.ActionHalt,
		},
		{
			name:       "fail watch progress",
			snapName:   "dummy-snap",
			failCause:  FailWatchProgress,
			wantAction: multistep.ActionHalt,
		},
	}
	const serverID = 42

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := make(chan error, 1)
			state, teardown := setupStepCreateSnapshot(errors, tc.failCause)
			defer teardown()

			step := &stepCreateSnapshot{}
			config := Config{
				SnapshotName: tc.snapName,
			}
			// do not output to stdout or console
			state.Put("ui", &packersdk.MockUi{})
			state.Put("config", &config)
			state.Put("server_id", serverID)

			if action := step.Run(context.Background(), state); action != tc.wantAction {
				t.Errorf("step.Run: want: %v; got: %v", tc.wantAction, action)
			}

			select {
			case err := <-errors:
				t.Errorf("server: got: %s", err)
			default:
			}
		})
	}
}

// Configure a httptest server to reply to the request to create a snapshot.
// React with the appropriate failCause.
// Report errors on the errors channel (cannot use testing.T, it runs on a different goroutine).
// Return a tuple (state, teardown) where:
// - state (containing the client) is ready to be passed to the step.Run() method.
// - teardown is a function meant to be deferred from the test.
func setupStepCreateSnapshot(
	errors chan<- error,
	failCause FailCause,
) (*multistep.BasicStateBag, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			errors <- fmt.Errorf("fake server: reading request: %s", err)
			return
		}
		reqDump := fmt.Sprintf("fake server: request:\n%s %s\nbody: %s", r.Method, r.URL.Path, string(buf))
		if testing.Verbose() {
			fmt.Println(reqDump)
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		var response interface{}
		action := schema.Action{
			ID:       13,
			Progress: 100,
			Status:   "success",
		}

		if r.Method == http.MethodPost && r.URL.Path == "/servers/42/actions/create_image" {
			if failCause == FailCreateImage {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			response = schema.ServerActionCreateImageResponse{Action: action}
		} else if r.Method == http.MethodGet && r.URL.Path == "/actions/13" {
			if failCause == FailWatchProgress {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			response = schema.ActionGetResponse{Action: action}
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
	client := hcloud.NewClient(hcloud.WithEndpoint(ts.URL))
	state.Put("hcloudClient", client)

	teardown := func() {
		ts.Close()
	}
	return &state, teardown
}
