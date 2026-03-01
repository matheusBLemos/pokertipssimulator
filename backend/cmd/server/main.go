package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"pokertipssimulator/config"
	"pokertipssimulator/internal/handler"
	"pokertipssimulator/internal/repository"
	"pokertipssimulator/internal/routes"
	"pokertipssimulator/internal/usecase"
	"pokertipssimulator/internal/ws"
	"pokertipssimulator/pkg/mongodb"
)

func main() {
	cfg := config.Load()

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

	roomRepo := repository.NewRoomRepository(db)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Port)

	<-quit
	log.Println("Shutting down server...")

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
