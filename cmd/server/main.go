package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pokertipssimulator/internal/adapter/network"
	"pokertipssimulator/internal/infrastructure/config"
	"pokertipssimulator/internal/server"
	"pokertipssimulator/pkg/envloader"
)

func main() {
	envloader.Load(".env")
	cfg := config.Load()

	srv := server.New()
	if err := srv.Start(cfg.Port, cfg.DBPath, cfg.JWTSecret); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	log.Printf("Server started on port %s", cfg.Port)
	printAccessInfo(cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := srv.Stop(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

func printAccessInfo(port string) {
	localIPs := network.GetLocalIPs()
	localIP := ""
	if len(localIPs) > 0 {
		localIP = localIPs[0]
	}
	publicIP := network.GetPublicIP()

	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────────────┐")
	fmt.Println("  │            Poker Tips Simulator               │")
	fmt.Println("  ├──────────────────────────────────────────────┤")
	fmt.Printf("  │  Local:   http://localhost:%-19s │\n", port)
	if localIP != "" {
		fmt.Printf("  │  LAN:     http://%-15s:%-8s │\n", localIP, port)
	}
	if publicIP != "" {
		fmt.Printf("  │  Public:  http://%-15s:%-8s │\n", publicIP, port)
	} else {
		fmt.Println("  │  Public:  (could not determine)              │")
	}
	fmt.Println("  └──────────────────────────────────────────────┘")
	fmt.Println()

	if len(localIPs) > 1 {
		fmt.Println("  Other LAN addresses:")
		for _, ip := range localIPs[1:] {
			fmt.Printf("    - http://%s:%s\n", ip, port)
		}
		fmt.Println()
	}
}
