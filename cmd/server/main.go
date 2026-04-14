package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pokertipssimulator/internal/adapter/network"
	appport "pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/infrastructure/config"
	"pokertipssimulator/internal/server"
	"pokertipssimulator/pkg/applog"
	"pokertipssimulator/pkg/envloader"
)

func main() {
	if logPath, closeLog, err := applog.Init(); err != nil {
		log.Printf("log init failed: %v", err)
	} else {
		defer closeLog()
		log.Printf("log file: %s", logPath)
	}

	envloader.Load(".env")
	cfg := config.Load()

	srv := server.New(nil)
	if err := srv.Start(cfg.Port, cfg.DBPath, cfg.JWTSecret); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	info := srv.GetConnectionInfo()
	upnpOK := network.MapPort(info.Port)
	srv.SetUPnPStatus(upnpOK)

	log.Printf("Server started on port %s", cfg.Port)
	printAccessInfo(srv.GetConnectionInfo(), upnpOK)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	network.UnmapPort(info.Port)
	if err := srv.Stop(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

func printAccessInfo(info appport.ConnectionInfo, upnpOK bool) {
	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────────────┐")
	fmt.Println("  │            Poker Tips Simulator              │")
	fmt.Println("  ├──────────────────────────────────────────────┤")
	fmt.Printf("  │  Local:   http://localhost:%-18d│\n", info.Port)
	if info.LocalURL != "" {
		fmt.Printf("  │  LAN:     %-35s│\n", info.LocalURL)
	}
	if info.PublicURL != "" {
		fmt.Printf("  │  Public:  %-35s│\n", info.PublicURL)
	} else {
		fmt.Println("  │  Public:  (could not determine)              │")
	}
	if upnpOK {
		fmt.Println("  │  UPnP:    port mapped automatically          │")
	} else {
		fmt.Println("  │  UPnP:    failed — manual port forward needed│")
	}
	fmt.Println("  └──────────────────────────────────────────────┘")
	fmt.Println()

	otherIPs := network.GetLocalIPs()
	if len(otherIPs) > 1 {
		fmt.Println("  Other LAN addresses:")
		for _, ip := range otherIPs[1:] {
			fmt.Printf("    - http://%s:%d\n", ip, info.Port)
		}
		fmt.Println()
	}
}
