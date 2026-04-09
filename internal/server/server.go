package server

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"pokertipssimulator/internal/adapter/handler"
	"pokertipssimulator/internal/adapter/network"
	"pokertipssimulator/internal/adapter/repository"
	"pokertipssimulator/internal/adapter/ws"
	"pokertipssimulator/internal/application"
	appport "pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/frontend"
	"pokertipssimulator/internal/infrastructure/auth"
	"pokertipssimulator/internal/infrastructure/database"
)

type Server struct {
	app     *fiber.App
	hub     *ws.Hub
	db      *sql.DB
	port    string
	running bool
	mu      sync.Mutex
	connInfo appport.ConnectionInfo
}

func New() *Server {
	return &Server{}
}

func (s *Server) Start(port, dbPath, jwtSecret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	db, err := database.OpenSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite: %w", err)
	}
	s.db = db

	roomRepo := repository.NewSQLiteRoomRepository(db)
	jwtService := auth.NewJWTService(jwtSecret)

	roomUC := application.NewRoomUseCase(roomRepo, jwtService)
	gameUC := application.NewGameUseCase(roomRepo)
	actionUC := application.NewActionUseCase(roomRepo)
	tipsUC := application.NewTipsUseCase(roomRepo)

	hub := ws.NewHub()
	go hub.Run()
	s.hub = hub

	roomHandler := handler.NewRoomHandler(roomUC)
	gameHandler := handler.NewGameHandler(gameUC, hub)
	actionHandler := handler.NewActionHandler(actionUC, hub)
	tipsHandler := handler.NewTipsHandler(tipsUC, hub)
	wsHandler := handler.NewWSHandler(hub, roomUC)

	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool { return true },
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))

	handler.SetupRoutes(app, roomHandler, gameHandler, actionHandler, tipsHandler, wsHandler, jwtSecret)

	app.Get("/api/v1/connection-info", func(c *fiber.Ctx) error {
		s.mu.Lock()
		info := s.connInfo
		s.mu.Unlock()
		return c.JSON(info)
	})

	frontend.RegisterSPA(app)

	s.app = app
	s.port = port

	localIPs := network.GetLocalIPs()
	localIP := ""
	if len(localIPs) > 0 {
		localIP = localIPs[0]
	}
	publicIP := network.GetPublicIP()

	s.connInfo = appport.ConnectionInfo{
		LocalIP:  localIP,
		PublicIP: publicIP,
		Port:     portToInt(port),
		UPnPOK:   false,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := app.Listen("0.0.0.0:" + port); err != nil {
			errCh <- err
		}
	}()

	// Give the server a moment to start or fail
	select {
	case err := <-errCh:
		s.db.Close()
		return fmt.Errorf("failed to start server: %w", err)
	case <-time.After(500 * time.Millisecond):
		s.running = true
		return nil
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if err := s.app.ShutdownWithTimeout(5 * time.Second); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	if s.db != nil {
		s.db.Close()
	}

	s.running = false
	return nil
}

func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Server) GetConnectionInfo() appport.ConnectionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connInfo
}

func (s *Server) SetUPnPStatus(ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connInfo.UPnPOK = ok
}

func (s *Server) Port() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

func portToInt(p string) int {
	n := 0
	for _, c := range p {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
