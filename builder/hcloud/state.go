package hcloud

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// State keys to put or get data from the state
const (
	StateConfig = "config"
	StateHook   = "hook"
	StateUI     = "ui"
	StateError  = "error"

	StateHCloudClient = "hcloud_client"

	StateGeneratedData = "generated_data"
	StateInstanceID    = "instance_id"
	StateServerID      = "server_id"
	StateServerIP      = "server_ip"
	StateServerType    = "server_type"
	StateSnapshotID    = "snapshot_id"
	StateSnapshotIDOld = "snapshot_id_old"
	StateSnapshotName  = "snapshot_name"
	StateSSHKeyID      = "ssh_key_id"
)

func UnpackState(state multistep.StateBag) (*Config, packersdk.Ui, *hcloud.Client) {
	config := state.Get(StateConfig).(*Config)
	ui := state.Get(StateUI).(packersdk.Ui)
	client := state.Get(StateHCloudClient).(*hcloud.Client)

	return config, ui, client
}
