package main

import (
	"context"
	"fmt"

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

func NewApp() *App {
	envloader.Load(".env")
	cfg := config.Load()
	return &App{
		srv: server.New(),
		cfg: cfg,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.srv.Stop()
}

func (a *App) StartServer(port int) (appport.ConnectionInfo, error) {
	p := fmt.Sprintf("%d", port)
	if port <= 0 {
		p = a.cfg.Port
	}

	if err := a.srv.Start(p, a.cfg.DBPath, a.cfg.JWTSecret); err != nil {
		return appport.ConnectionInfo{}, err
	}

	upnpOK := network.MapPort(port)
	a.srv.SetUPnPStatus(upnpOK)

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
