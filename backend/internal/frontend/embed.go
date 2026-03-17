package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed dist/*
var distFS embed.FS

// RegisterSPA serves the embedded frontend files with SPA fallback.
// It should be called after all API and WS routes are registered.
func RegisterSPA(app *fiber.App) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("failed to create sub filesystem: " + err.Error())
	}

	// Check if there's an actual frontend build (not just .gitkeep)
	entries, _ := fs.ReadDir(sub, ".")
	hasBuild := false
	for _, e := range entries {
		if e.Name() != ".gitkeep" {
			hasBuild = true
			break
		}
	}
	if !hasBuild {
		return
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(sub),
		Browse:     false,
		Index:      "index.html",
		NotFoundFile: "index.html",
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws")
		},
	}))
}
