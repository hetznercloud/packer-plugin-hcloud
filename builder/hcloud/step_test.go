package hcloud

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/mockutils"
)

type StepTestCase struct {
	Name         string
	Step         multistep.Step
	StepFuncName string

	SetupConfigFunc func(*Config)
	SetupStateFunc  func(multistep.StateBag)

	WantRequests []mockutils.Request

	WantStepAction multistep.StepAction
	WantStateFunc  func(*testing.T, multistep.StateBag)
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

			server := httptest.NewServer(mockutils.Handler(t, tc.WantRequests))
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

func NewTestState(t *testing.T) multistep.StateBag {
	state := &multistep.BasicStateBag{}

	if testing.Verbose() {
		state.Put(StateUI, packersdk.TestUi(t))
	} else {
		state.Put(StateUI, &packersdk.MockUi{})
	}

	return state
}

func decodeJSONBody[T any](t *testing.T, body io.ReadCloser, v *T) *T {
	t.Helper()
	require.NoError(t, json.NewDecoder(body).Decode(v))
	return v
}
