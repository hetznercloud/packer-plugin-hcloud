package hcloud

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestBuilderAcc_basic(t *testing.T) {
	if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
		t.Skip("HCLOUD_TOKEN must be set for acceptance tests")
	}

	testCase := &acctest.PluginTestCase{
		Name:     "hcloud_basic_test",
		Template: testBuilderAccBasic,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"location": "nbg1",
		"server_type": "cx11",
		"image": "ubuntu-22.04",
		"user_data": "",
		"user_data_file": "",
		"ssh_username": "root"
	}]
}
`
