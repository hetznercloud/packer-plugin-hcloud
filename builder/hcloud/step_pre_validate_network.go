package hcloud

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type stepPreValidateNetwork struct {
	net_id int
}

func (s *stepPreValidateNetwork) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	var network *hcloud.Network
	var subnetiprange *net.IPNet
	var subnetgateway net.IP
	var server_list []*hcloud.Server
	var hosts []string
	var server_ip_array []net.IP

	ui.Say(fmt.Sprintf("Prevalidating Network: %s", c.Network))

	// check if input of Network matches existing network - output network_id
	if c.Network != "" {
		// check if input can be converted to string
		network_id, err := strconv.Atoi(c.Network)
		// Lookup by Name
		if err != nil {
			network, _, err = client.Network.GetByName(ctx, c.Network)
			if err != nil {
				ui.Error(err.Error())
				state.Put("error", fmt.Errorf("Error fetching Network ID: %s", err))
				return multistep.ActionHalt
			}
			if network == nil {
				state.Put("error", fmt.Errorf("Could not find network: %s", err))
				return multistep.ActionHalt
			}
			server_list = network.Servers
			s.net_id = network.ID
			state.Put("network_id", s.net_id)

		}
		// Lookup by ID
		if network == nil {
			network, _, err = client.Network.GetByID(ctx, network_id)
			if err != nil {
				ui.Error(err.Error())
				state.Put("error", fmt.Errorf("Error fetching Network ID: %s", err))
				return multistep.ActionHalt
			}
			if network == nil {
				state.Put("error", fmt.Errorf("Could not find network: %s", err))
				return multistep.ActionHalt
			}
			server_list = network.Servers
			s.net_id = network.ID
			state.Put("network_id", s.net_id)

		}

		if c.Subnet != "" {
			for _, x := range network.Subnets {
				if x.IPRange.String() == c.Subnet {
					subnetiprange = x.IPRange
					subnetgateway = x.Gateway

				}
			}
			for _, y := range server_list {
				server, _, err := client.Server.GetByID(ctx, y.ID)
				if err != nil {
					state.Put("error", fmt.Errorf("Could not Server to retrieve private IP: %s", err))
					return multistep.ActionHalt
				}
				if err == nil {
					for _, x := range server.PrivateNet {
						server_ip_array = append(server_ip_array, x.IP)
						if len(x.Aliases) != 0 {
							server_ip_array = append(server_ip_array, x.Aliases...)
						}
					}
				}
			}
			mask := binary.BigEndian.Uint32(subnetiprange.Mask)
			start := binary.BigEndian.Uint32(subnetiprange.IP)
			finish := (start & mask) | (mask ^ 0xffffffff)
			for i := start + 1; i <= finish-1; i++ {
				ip := make(net.IP, 4)
				binary.BigEndian.PutUint32(ip, i)
				hosts = append(hosts, ip.String())
			}
			for i := 0; i < len(hosts); i++ {
				host := hosts[i]
				if host == subnetgateway.String() {
					hosts = append(hosts[:i], hosts[i+1:]...)
					i--
				}
				for _, ip_to_remove := range server_ip_array {
					if host == ip_to_remove.String() {
						hosts = append(hosts[:i], hosts[i+1:]...)
						i--
						break
					}
				}
			}
			state.Put("srv_ip", hosts[0])
		}

	}
	if len(hosts) == 0 {
		state.Put("srv_ip", "")
	}
	return multistep.ActionContinue
}

// No-op
func (s *stepPreValidateNetwork) Cleanup(multistep.StateBag) {
}
