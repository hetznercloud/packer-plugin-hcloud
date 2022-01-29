package hcloud

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type stepAttachNetwork struct {
}

func (s *stepAttachNetwork) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	serverid := state.Get("server_id").(int)
	srv := &hcloud.Server{ID: serverid}
	network_id, err := strconv.Atoi(c.Network)
	if err != nil {
		err = fmt.Errorf("Error conversiong network_id: %s", err)
	}
	nw := &hcloud.Network{ID: network_id}
	var ip net.IP
	var aliasIPs []net.IP

	// Attach Network to server based on configuration
	ui.Say("Attaching Network to Server...")

	if c.IP != "" {
		ip = net.ParseIP(c.IP)
	}

	// Check if Alias IP is set and convert string to net.IP
	if len(c.AliasIPs) != 0 {
		for _, a := range c.AliasIPs {
			aliasip := net.ParseIP(a)
			aliasIPs = append(aliasIPs, aliasip)
		}
	}

	// get network
	// networks_new0, _, err := client.Network.Get(ctx, "1130670")
	// ui.Say(fmt.Sprintf("network: ", networks_new0))

	// syntax for create server and attach to network directly
	// net_id := []*hcloud.Network{{ID: network_id}}

	_, _, err = client.Server.AttachToNetwork(ctx, srv, hcloud.ServerAttachToNetworkOpts{
		Network:  nw,
		IP:       ip,
		AliasIPs: aliasIPs,
	})
	if err != nil {
		err := fmt.Errorf("Error attaching Network: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// override ssh host - set connect_with_private_ip to use your set ip_address to connect
	if c.ConnectWithPrivateIP == true {
		state.Put("server_ip", c.IP)
	}

	return multistep.ActionContinue
}

func (s *stepAttachNetwork) Cleanup(state multistep.StateBag) {
	// cleanup by stepCreateServer
}
