package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"pokertipssimulator/config"
	"pokertipssimulator/internal/frontend"
	"pokertipssimulator/internal/handler"
	"pokertipssimulator/internal/repository"
	"pokertipssimulator/internal/routes"
	"pokertipssimulator/internal/usecase"
	"pokertipssimulator/internal/ws"
	"pokertipssimulator/pkg/envloader"
	"pokertipssimulator/pkg/mongodb"
	"pokertipssimulator/pkg/sqlite"
)

func main() {
	envloader.Load(".env")
	cfg := config.Load()

	var roomRepo repository.RoomRepository

	if cfg.MongoURI != "" {
		// MongoDB mode (dev with Docker)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db, err := mongodb.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
		if err != nil {
			log.Fatalf("failed to connect to MongoDB: %v", err)
		}
		defer func() {
			if err := db.Client().Disconnect(context.Background()); err != nil {
				log.Printf("failed to disconnect from MongoDB: %v", err)
			}
		}()

		roomRepo = repository.NewRoomRepository(db)
		log.Println("Using MongoDB storage")
	} else {
		// SQLite mode (local build)
		db, err := sqlite.Connect(cfg.DBPath)
		if err != nil {
			log.Fatalf("failed to open SQLite: %v", err)
		}
		defer db.Close()

		roomRepo = repository.NewSQLiteRoomRepository(db)
		log.Printf("Using SQLite storage (%s)", cfg.DBPath)
	}

	roomUC := usecase.NewRoomUseCase(roomRepo, cfg.JWTSecret)
	gameUC := usecase.NewGameUseCase(roomRepo)
	actionUC := usecase.NewActionUseCase(roomRepo)

	hub := ws.NewHub()
	go hub.Run()

	roomHandler := handler.NewRoomHandler(roomUC)
	gameHandler := handler.NewGameHandler(gameUC, hub)
	actionHandler := handler.NewActionHandler(actionUC, hub)
	wsHandler := handler.NewWSHandler(hub, roomUC)

	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.Setup(app, roomHandler, gameHandler, actionHandler, wsHandler, cfg.JWTSecret)

	// Serve embedded frontend (SPA) after API routes
	frontend.RegisterSPA(app)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen("0.0.0.0:" + cfg.Port); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Port)
	printAccessInfo(cfg.Port)

	<-quit
	log.Println("Shutting down server...")

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}

func printAccessInfo(port string) {
	localIPs := getLocalIPs()
	localIP := ""
	if len(localIPs) > 0 {
		localIP = localIPs[0]
	}
	publicIP := getPublicIP()

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

func getLocalIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	type candidate struct {
		ip    string
		score int
	}

	candidates := make([]candidate, 0)
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

			candidates = append(candidates, candidate{ip: ipStr, score: score})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].ip < candidates[j].ip
		}
		return candidates[i].score > candidates[j].score
	})

	localIPs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		localIPs = append(localIPs, c.ip)
	}
	return localIPs
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

func getPublicIP() string {
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
