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
	// create labels map
	labels := make(map[string]string)

	// This label contains the value the user specified in their template
	sourceImage, ok := a.StateData["source_image"].(string)
	if ok {
		labels["source_image"] = sourceImage
	}
	// This is the canonical ID of the source image that was used, useful for ancestry tracking
	sourceImageID, ok := a.StateData["source_image_id"].(int64)
	// get and set region from stateData into labels
	region, ok := a.StateData["region"].(string)
	if ok {
		labels["region"] = region
	}
	// get and set server_type from stateData into labels
	serverType, ok := a.StateData["server_type"].(string)
	if ok {
		labels["server_type"] = serverType
	}

	return &registryimage.Image{
		ProviderName:   "hetznercloud",
		ImageID:        a.Id(),
		ProviderRegion: region,
		Labels:         labels,
		SourceImageID:  strconv.FormatInt(sourceImageID, 10),
	}
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d (%s)", a.snapshotId, a.snapshotName)
	_, err := a.hcloudClient.Image.Delete(context.TODO(), &hcloud.Image{ID: a.snapshotId})
	return err
}
