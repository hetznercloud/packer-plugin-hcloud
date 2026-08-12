// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPrepareSSHInterface(t *testing.T) {
	testCases := []struct {
		name         string
		sshInterface string
		wantErr      string
	}{
		{
			name:         "public_ipv4",
			sshInterface: "public_ipv4",
		},
		{
			name:         "public_ipv6",
			sshInterface: "public_ipv6",
		},
		{
			name:         "private_ipv4",
			sshInterface: "private_ipv4",
		},
		{
			name:         "invalid",
			sshInterface: "public",
			wantErr:      "ssh_interface must be one of public_ipv4, public_ipv6, or private_ipv4",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := &Config{}
			_, err := c.Prepare(map[string]interface{}{
				"token":         "dummy-token",
				"image":         "debian-12",
				"location":      "nbg1",
				"server_type":   "cpx22",
				"ssh_username":  "root",
				"ssh_interface": testCase.sshInterface,
			})

			if testCase.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
