// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestBuilderAcc_basic(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "hcloud_basic_test",
		Setup: func() error {
			if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
				return fmt.Errorf("HCLOUD_TOKEN must be set for acceptance tests")
			}
			return nil
		},
		Template: testBuilderAccBasic,
		Check: func(buildCommand *exec.Cmd, logFile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() == 0 {
					return nil
				}

				logs, err := os.ReadFile(logFile)
				if err != nil {
					return err
				}
				return fmt.Errorf("invalid exit code: %d\n%s",
					buildCommand.ProcessState.ExitCode(),
					logs,
				)
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "hcloud",
		"location": "nbg1",
		"server_type": "cx11",
		"image": "ubuntu-22.04",
		"user_data": "",
		"user_data_file": "",
		"ssh_username": "root"
	}]
}
`
