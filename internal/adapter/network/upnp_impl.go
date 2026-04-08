package network

import (
	"fmt"
	"log"

	"github.com/huin/goupnp/dcps/internetgateway2"
)

type portMapper interface {
	AddPortMapping(
		NewRemoteHost string,
		NewExternalPort uint16,
		NewProtocol string,
		NewInternalPort uint16,
		NewInternalClient string,
		NewEnabled bool,
		NewPortMappingDescription string,
		NewLeaseDuration uint32,
	) error
	DeletePortMapping(
		NewRemoteHost string,
		NewExternalPort uint16,
		NewProtocol string,
	) error
}

type gateway struct {
	client portMapper
}

func discoverGateway() (*gateway, error) {
	// Try WANIPConnection2 first (newer routers)
	clients2, _, _ := internetgateway2.NewWANIPConnection2Clients()
	if len(clients2) > 0 {
		return &gateway{client: clients2[0]}, nil
	}

	// Fall back to WANIPConnection1
	clients1, _, _ := internetgateway2.NewWANIPConnection1Clients()
	if len(clients1) > 0 {
		return &gateway{client: clients1[0]}, nil
	}

	return nil, fmt.Errorf("no UPnP gateway found")
}

func (g *gateway) addPortMapping(port int) bool {
	if g.client == nil {
		return false
	}

	localIP := getPreferredLocalIP()
	if localIP == "" {
		return false
	}

	err := g.client.AddPortMapping(
		"",
		uint16(port),
		"TCP",
		uint16(port),
		localIP,
		true,
		"PokerApp",
		0,
	)
	if err != nil {
		log.Printf("UPnP AddPortMapping error: %v", err)
		return false
	}
	return true
}

func (g *gateway) removePortMapping(port int) {
	if g.client == nil {
		return
	}
	_ = g.client.DeletePortMapping("", uint16(port), "TCP")
}

func getPreferredLocalIP() string {
	ips := GetLocalIPs()
	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}
