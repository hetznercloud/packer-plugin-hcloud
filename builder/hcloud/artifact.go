// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcloud

import (
	"context"
	"fmt"
	"log"
	"strconv"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type Artifact struct {
	// The name of the snapshot
	snapshotName string

	// The ID of the image
	snapshotId int64

	// The hcloudClient for making API calls
	hcloudClient *hcloud.Client

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return strconv.FormatInt(a.snapshotId, 10)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%v' (ID: %v)", a.snapshotName, a.snapshotId)
}

func (a *Artifact) State(name string) interface{} {
	if name == registryimage.ArtifactStateURI {
		return a.stateHCPPackerRegistryMetadata()
	}
	return a.StateData[name]
}

func (a *Artifact) stateHCPPackerRegistryMetadata() interface{} {
	labels := make(map[string]string)

	// Those labels contains the value the user specified in their template
	sourceImage, ok := a.StateData["source_image"].(string)
	if ok {
		labels["source_image"] = sourceImage
	}
	serverType, ok := a.StateData["server_type"].(string)
	if ok {
		labels["server_type"] = serverType
	}

	img := &registryimage.Image{
		ImageID:      a.Id(),
		ProviderName: "hetznercloud", // Use explicit name over the builder ID
		Labels:       labels,
	}

	sourceImageID, ok := a.StateData["source_image_id"].(int64)
	if ok {
		img.SourceImageID = strconv.FormatInt(sourceImageID, 10)
	}

	return img
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d (%s)", a.snapshotId, a.snapshotName)
	_, err := a.hcloudClient.Image.Delete(context.TODO(), &hcloud.Image{ID: a.snapshotId})
	return err
}
