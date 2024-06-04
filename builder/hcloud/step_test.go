package hcloud

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type StepTestCase struct {
	Name         string
	Step         multistep.Step
	StepFuncName string

	SetupConfigFunc func(*Config)
	SetupStateFunc  func(multistep.StateBag)

	WantRequests []Request

	WantStepAction multistep.StepAction
	WantStateFunc  func(*testing.T, multistep.StateBag)
}

type Request struct {
	Method              string
	Path                string
	WantRequestBodyFunc func(t *testing.T, r *http.Request, body []byte)

	Status int
	Body   string
}

func RunStepTestCases(t *testing.T, testCases []StepTestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := &Config{
				ServerName:   "dummy-server",
				Image:        "debian-12",
				SnapshotName: "dummy-snapshot",
				ServerType:   "cpx11",
				Location:     "nbg1",
				SSHKeys:      []string{"1"},
			}

			if tc.SetupConfigFunc != nil {
				tc.SetupConfigFunc(config)
			}

			server := NewTestServer(t, tc.WantRequests)
			defer server.Close()
			client := hcloud.NewClient(hcloud.WithEndpoint(server.URL))

			state := NewTestState(t)
			state.Put(StateConfig, config)
			state.Put(StateHCloudClient, client)

			if tc.SetupStateFunc != nil {
				tc.SetupStateFunc(state)
			}

			switch strings.ToLower(tc.StepFuncName) {
			case "cleanup":
				tc.Step.Cleanup(state)
			default:
				action := tc.Step.Run(context.Background(), state)
				assert.Equal(t, tc.WantStepAction, action)
			}

			if tc.WantStateFunc != nil {
				tc.WantStateFunc(t, state)
			}
		})
	}
}

func NewTestServer(t *testing.T, requests []Request) *httptest.Server {
	index := 0

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if testing.Verbose() {
			t.Logf("request %d: %s %s\n", index, r.Method, r.RequestURI)
		}

		if index >= len(requests) {
			t.Fatalf("received unknown request %d", index)
		}

		response := requests[index]
		assert.Equal(t, response.Method, r.Method)
		assert.Equal(t, response.Path, r.RequestURI)

		if response.WantRequestBodyFunc != nil {
			buffer, err := io.ReadAll(r.Body)
			defer func() {
				if err := r.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			if err != nil {
				t.Fatal(err)
			}
			response.WantRequestBodyFunc(t, r, buffer)
		}

		w.WriteHeader(response.Status)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(response.Body))
		if err != nil {
			t.Fatal(err)
		}

		index++
	}))
}

func NewTestState(t *testing.T) multistep.StateBag {
	state := &multistep.BasicStateBag{}

	if testing.Verbose() {
		state.Put(StateUI, packersdk.TestUi(t))
	} else {
		state.Put(StateUI, &packersdk.MockUi{})
	}

	return state
}
