package server

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"sync"
	"syscall"
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
	app        *fiber.App
	hub        *ws.Hub
	db         *sql.DB
	port       string
	running    bool
	mu         sync.Mutex
	connInfo   appport.ConnectionInfo
	frontendFS fs.FS
}

func New(frontendFS fs.FS) *Server {
	return &Server{frontendFS: frontendFS}
}

func (s *Server) Start(port, dbPath, jwtSecret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	log.Printf("server start: port=%s dbPath=%s", port, dbPath)

	db, err := database.OpenSQLite(dbPath)
	if err != nil {
		log.Printf("server start: sqlite open failed: %v", err)
		return fmt.Errorf("failed to open SQLite: %w", err)
	}
	s.db = db
	log.Printf("server start: sqlite ready")

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
		ErrorHandler:          handler.ErrorHandler,
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Output: log.Writer(),
	}))
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

	frontend.RegisterSPA(app, s.frontendFS)

	s.app = app
	s.port = port

	localIPs := network.GetLocalIPs()
	localIP := ""
	if len(localIPs) > 0 {
		localIP = localIPs[0]
	}
	publicIP := network.GetPublicIP()
	portNum := portToInt(port)

	s.connInfo = appport.ConnectionInfo{
		LocalIP:   localIP,
		PublicIP:  publicIP,
		Port:      portNum,
		UPnPOK:    false,
		LocalURL:  buildURL(localIP, portNum),
		PublicURL: buildURL(publicIP, portNum),
	}

	ln, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Printf("server listen error: %v", err)
		s.db.Close()
		s.db = nil
		if isAddrInUse(err) {
			return fmt.Errorf("port %s is already in use — close the other program or choose a different port", port)
		}
		return fmt.Errorf("failed to bind port %s: %w", port, err)
	}

	go func() {
		if serveErr := app.Listener(ln); serveErr != nil {
			log.Printf("fiber serve error: %v", serveErr)
		}
	}()

	s.running = true
	log.Printf("server start: listening on %s", ln.Addr().String())
	return nil
}

func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
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

func buildURL(ip string, port int) string {
	if ip == "" || port == 0 {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", ip, port)
}
