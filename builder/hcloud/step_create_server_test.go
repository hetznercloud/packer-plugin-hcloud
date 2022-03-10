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
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type Checker func(requestBody string, path string) error

func TestStepCreateServer(t *testing.T) {
	const snapName = "dummy-snap"
	const imageName = "dummy-image"
	const name = "dummy-name"
	const location = "nbg1"
	const serverType = "cpx11"
	networks := []int64{1}

	testCases := []struct {
		name       string
		config     Config
		check      Checker
		wantAction multistep.StepAction
	}{
		{
			name:       "happy path",
			wantAction: multistep.ActionContinue,
			check: func(r string, path string) error {
				if path == "/servers" {
					payload := schema.ServerCreateRequest{}
					err := json.Unmarshal([]byte(r), &payload)
					if err != nil {
						t.Errorf("server request not a json: got: (%s)", err)
					}

					if payload.Name != name {
						t.Errorf("Incorrect name in request, expected '%s' found '%s'", name, payload.Name)
					}

					if payload.Image != imageName {
						t.Errorf("Incorrect image in request, expected '%s' found '%s'", imageName, payload.Image)
					}

					if payload.Location != location {
						t.Errorf("Incorrect location in request, expected '%s' found '%s'", location, payload.Location)
					}

					if payload.ServerType != serverType {
						t.Errorf("Incorrect serverType in request, expected '%s' found '%s'", serverType, payload.ServerType)
					}
					if payload.Networks != nil {
						t.Error("Networks should not be specified")
					}
				}
				return nil
			},
		},
		{
			name:       "with netowork",
			wantAction: multistep.ActionContinue,
			config: Config{
				Networks: networks,
			},
			check: func(r string, path string) error {
				if path == "/servers" {
					payload := schema.ServerCreateRequest{}
					err := json.Unmarshal([]byte(r), &payload)
					if err != nil {
						t.Errorf("server request not a json: (%s)", err)
					}
					if payload.Networks[0] != networks[0] {
						t.Errorf("network not set")
					}
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := make(chan error, 1)
			state, teardown := setupStepCreateServer(errors, tc.check)
			defer teardown()

			step := &stepCreateServer{}

			baseConfig := Config{
				ServerName:   name,
				Image:        imageName,
				SnapshotName: snapName,
				ServerType:   serverType,
				Location:     location,
				SSHKeys:      []string{"1"},
			}

			config := baseConfig
			config.Networks = tc.config.Networks

			if testing.Verbose() {
				state.Put("ui", packersdk.TestUi(t))
			} else {
				// do not output to stdout or console
				state.Put("ui", &packersdk.MockUi{})
			}
			state.Put("config", &config)
			state.Put("ssh_key_id", int64(1))

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

// Configure a httptest server to reply to the requests done by stepCreateSnapshot.
// React with the appropriate failCause.
// Report errors on the errors channel (cannot use testing.T, it runs on a different goroutine).
// Return a tuple (state, teardown) where:
// - state (containing the client) is ready to be passed to the step.Run() method.
// - teardown is a function meant to be deferred from the test.
func setupStepCreateServer(
	errors chan<- error,
	checker Checker,
) (*multistep.BasicStateBag, func()) {
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

		enc := json.NewEncoder(w)
		var response interface{}
		action := schema.Action{
			ID:     1,
			Status: "success",
		}

		if r.Method == http.MethodPost && r.URL.Path == "/servers" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response = schema.ServerCreateResponse{Action: action}
		}

		if r.Method == http.MethodGet && r.URL.Path == "/actions/1" {
			w.Header().Set("Content-Type", "application/json")
			response = schema.ActionGetResponse{Action: action}
		}

		if r.Method == http.MethodGet && r.URL.Path == "/ssh_keys/1" {
			w.Header().Set("Content-Type", "application/json")
			response = schema.SSHKeyGetResponse{
				SSHKey: schema.SSHKey{ID: 1},
			}
		}

		if err := checker(string(buf), r.URL.Path); err != nil {
			errors <- fmt.Errorf("Error in checker")
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
