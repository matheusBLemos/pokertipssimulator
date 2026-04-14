package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"

	"pokertipssimulator/internal/adapter/network"
	appport "pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/infrastructure/config"
	"pokertipssimulator/internal/server"
	"pokertipssimulator/pkg/envloader"
)

type App struct {
	ctx context.Context
	srv *server.Server
	cfg *config.Config
}

func NewApp(frontendFS fs.FS) *App {
	envloader.Load(".env")
	cfg := config.Load()
	log.Printf("config loaded: port=%s dbPath=%s", cfg.Port, cfg.DBPath)
	return &App{
		srv: server.New(frontendFS),
		cfg: cfg,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("wails startup complete")
}

func (a *App) shutdown(ctx context.Context) {
	info := a.srv.GetConnectionInfo()
	if info.Port > 0 {
		network.UnmapPort(info.Port)
	}
	_ = a.srv.Stop()
}

func (a *App) StartServer(port int) (appport.ConnectionInfo, error) {
	p := fmt.Sprintf("%d", port)
	if port <= 0 {
		p = a.cfg.Port
	}

	log.Printf("StartServer requested: port=%s", p)
	if err := a.srv.Start(p, a.cfg.DBPath, a.cfg.JWTSecret); err != nil {
		log.Printf("StartServer failed: %v", err)
		return appport.ConnectionInfo{}, err
	}

	info := a.srv.GetConnectionInfo()
	upnpOK := network.MapPort(info.Port)
	a.srv.SetUPnPStatus(upnpOK)
	log.Printf("StartServer ready: local=%s public=%s upnp=%v",
		info.LocalURL, info.PublicURL, upnpOK)

	return a.srv.GetConnectionInfo(), nil
}

func (a *App) StopServer() error {
	info := a.srv.GetConnectionInfo()
	network.UnmapPort(info.Port)
	return a.srv.Stop()
}

func (a *App) GetConnectionInfo() appport.ConnectionInfo {
	return a.srv.GetConnectionInfo()
}

func (a *App) IsServerRunning() bool {
	return a.srv.IsRunning()
}
