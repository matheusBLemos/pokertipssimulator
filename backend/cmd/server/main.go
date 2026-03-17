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
		if err := app.Listen(":" + cfg.Port); err != nil {
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
	localIP := getLocalIP()
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
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return ""
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
