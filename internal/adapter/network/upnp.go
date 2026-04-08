package network

import (
	"log"
	"sync"
)

var (
	mappedPort int
	mappedMu   sync.Mutex
)

// MapPort attempts to create a UPnP port mapping on the local gateway router.
// Returns true if the mapping was successful.
func MapPort(port int) bool {
	mappedMu.Lock()
	defer mappedMu.Unlock()

	ok := tryUPnPMapping(port)
	if ok {
		mappedPort = port
		log.Printf("UPnP: port %d mapped successfully", port)
	} else {
		log.Printf("UPnP: port mapping failed or unavailable — manual port forwarding required")
	}
	return ok
}

// UnmapPort removes a previously created UPnP port mapping.
func UnmapPort(port int) {
	mappedMu.Lock()
	defer mappedMu.Unlock()

	if mappedPort == 0 {
		return
	}

	removeUPnPMapping(port)
	mappedPort = 0
	log.Printf("UPnP: port %d unmapped", port)
}

func tryUPnPMapping(port int) bool {
	gw, err := discoverGateway()
	if err != nil {
		return false
	}
	return gw.addPortMapping(port)
}

func removeUPnPMapping(port int) {
	gw, err := discoverGateway()
	if err != nil {
		return
	}
	gw.removePortMapping(port)
}
