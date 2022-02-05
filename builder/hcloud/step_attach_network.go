package hcloud

import (
	"context"
	"fmt"
	"net"

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
	srv_ip := state.Get("srv_ip").(string)
	srv := &hcloud.Server{ID: serverid}
	network_id := state.Get("network_id").(int)
	nw := &hcloud.Network{ID: network_id}
	var ip net.IP
	var aliasIPs []net.IP

	// Attach Network to server based on configuration
	ui.Say("Attaching Network to Server...")

	if c.IP != "" {
		ip = net.ParseIP(c.IP)
	}

	if c.Subnet != "" && c.IP == "" {
		ip = net.ParseIP(srv_ip)
	}

	// Check if Alias IP is set and convert string to net.IP
	if len(c.AliasIPs) != 0 {
		for _, a := range c.AliasIPs {
			aliasip := net.ParseIP(a)
			aliasIPs = append(aliasIPs, aliasip)
		}
	}

	_, _, err := client.Server.AttachToNetwork(ctx, srv, hcloud.ServerAttachToNetworkOpts{
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
	// override ssh host
	if c.ConnectWithPrivateIP == true {
		server, _, err := client.Server.GetByID(ctx, serverid)
		if err != nil {
			state.Put("error", fmt.Errorf("Could not Server to retrieve private IP: %s", err))
			return multistep.ActionHalt
		}
		if err == nil {
			n := server.PrivateNet
			for _, x := range n {
				state.Put("server_ip", x.IP.String())
			}
		}
	}
	return multistep.ActionContinue
}

func (s *stepAttachNetwork) Cleanup(state multistep.StateBag) {
	// cleanup by stepCreateServer
}
