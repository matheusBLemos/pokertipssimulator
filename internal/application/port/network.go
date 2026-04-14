package port

type ConnectionInfo struct {
	LocalIP   string `json:"local_ip"`
	PublicIP  string `json:"public_ip"`
	Port      int    `json:"port"`
	UPnPOK    bool   `json:"upnp_ok"`
	LocalURL  string `json:"local_url"`
	PublicURL string `json:"public_url"`
}

type NetworkManager interface {
	StartServer(port int) (ConnectionInfo, error)
	StopServer() error
	GetConnectionInfo() ConnectionInfo
	IsRunning() bool
}
