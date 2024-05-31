// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packersdk.Artifact = (*Artifact)(nil)
}

func TestArtifactId(t *testing.T) {
	generatedData := make(map[string]interface{})
	a := &Artifact{"packer-foobar", 42, nil, generatedData}
	expected := "42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	generatedData := make(map[string]interface{})
	a := &Artifact{"packer-foobar", 42, nil, generatedData}
	expected := "A snapshot was created: 'packer-foobar' (ID: 42)"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}

func TestArtifactState_StateData(t *testing.T) {
	expectedData := "this is the data"
	artifact := &Artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := artifact.State("state_data")
	if result != expectedData {
		t.Fatalf("Bad: State data was %s instead of %s", result, expectedData)
	}

	// Invalid state
	result = artifact.State("invalid_key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for invalid state data name")
	}

	// Nil StateData should not fail and should return nil
	artifact = &Artifact{}
	result = artifact.State("key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for nil StateData")
	}
}

func TestArtifactState_hcpPackerRegistryMetadata(t *testing.T) {
	artifact := &Artifact{
		snapshotId:   167438588,
		snapshotName: "test-image",
		StateData: map[string]interface{}{
			"source_image":    "ubuntu-24.04",
			"source_image_id": int64(161547269),
			"server_type":     "cpx11",
		},
	}

	result := artifact.State(registryimage.ArtifactStateURI)
	require.NotNil(t, result)

	var image registryimage.Image
	if err := mapstructure.Decode(result, &image); err != nil {
		t.Errorf("unexpected error when trying to decode state into registryimage.Image %v", err)
	}

	assert.Equal(t, registryimage.Image{
		ImageID:       "167438588",
		ProviderName:  "hetznercloud",
		SourceImageID: "161547269",
		Labels: map[string]string{
			"source_image": "ubuntu-24.04",
			"server_type":  "cpx11",
		},
	}, image)
}
