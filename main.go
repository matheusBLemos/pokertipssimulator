package main

import (
	"embed"
	"io/fs"
	"log"
	"runtime/debug"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"pokertipssimulator/pkg/applog"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logPath, closeLog, logErr := applog.Init()
	if logErr != nil {
		log.Printf("log init failed: %v", logErr)
	} else {
		defer closeLog()
		log.Printf("log file: %s", logPath)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC: %v\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	spaFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatalf("failed to resolve embedded frontend: %v", err)
	}

	app := NewApp(spaFS)

	err = wails.Run(&options.App{
		Title:  "Poker Application",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
