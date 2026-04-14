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

// RegisterSPA serves the frontend SPA with a fallback to index.html for
// client-side routing. If customFS is non-nil, it is used as the source of
// files — the Wails build passes its own fresh embed.FS of frontend/dist
// this way. Otherwise the stale-prone internal/frontend/dist copy is used,
// which is only populated by `make embed-frontend` for headless builds.
func RegisterSPA(app *fiber.App, customFS fs.FS) {
	sub, ok := resolveSPAFS(customFS)
	if !ok {
		return
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(sub),
		Browse:       false,
		Index:        "index.html",
		NotFoundFile: "index.html",
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws")
		},
	}))
}

func resolveSPAFS(customFS fs.FS) (fs.FS, bool) {
	if customFS != nil {
		if hasBuild(customFS) {
			return customFS, true
		}
		return nil, false
	}

	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, false
	}
	if !hasBuild(sub) {
		return nil, false
	}
	return sub, true
}

func hasBuild(f fs.FS) bool {
	entries, err := fs.ReadDir(f, ".")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.Name() != ".gitkeep" {
			return true
		}
	}
	return false
}
