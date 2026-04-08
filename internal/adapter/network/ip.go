package network

import (
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

type IPCandidate struct {
	IP    string
	Score int
}

func GetLocalIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	candidates := make([]IPCandidate, 0)
	seen := make(map[string]struct{})

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if shouldIgnoreInterface(iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}

			ipStr := ip.String()
			if _, ok := seen[ipStr]; ok {
				continue
			}
			seen[ipStr] = struct{}{}

			score := 0
			if isPrivateIPv4(ip) {
				score += 100
			}
			if iface.Flags&net.FlagPointToPoint != 0 {
				score -= 50
			}

			lowerName := strings.ToLower(iface.Name)
			switch {
			case strings.HasPrefix(lowerName, "en0"):
				score += 60
			case strings.HasPrefix(lowerName, "en"),
				strings.HasPrefix(lowerName, "eth"),
				strings.HasPrefix(lowerName, "wlan"),
				strings.HasPrefix(lowerName, "wl"):
				score += 40
			}

			if strings.HasPrefix(ipStr, "169.254.") {
				score -= 100
			}

			candidates = append(candidates, IPCandidate{IP: ipStr, Score: score})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].IP < candidates[j].IP
		}
		return candidates[i].Score > candidates[j].Score
	})

	localIPs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		localIPs = append(localIPs, c.IP)
	}
	return localIPs
}

func GetPublicIP() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}

func shouldIgnoreInterface(name string) bool {
	lower := strings.ToLower(name)
	ignoredPrefixes := []string{
		"lo", "utun", "bridge", "docker", "br-", "veth",
		"awdl", "llw", "anpi", "tap", "tun", "wg",
		"tailscale", "vboxnet", "vmnet",
	}
	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

func isPrivateIPv4(ip net.IP) bool {
	v4 := ip.To4()
	if v4 == nil {
		return false
	}
	return v4[0] == 10 ||
		(v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31) ||
		(v4[0] == 192 && v4[1] == 168)
}
