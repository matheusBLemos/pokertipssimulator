package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DBPath    string
	Port      string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		DBPath:    getEnv("DB_PATH", defaultDBPath()),
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// defaultDBPath returns a per-user, writable location for the SQLite file.
// When the binary is launched from Finder as a bundled .app, the working
// directory is "/" and a relative path cannot be opened for writing, so we
// resolve the OS user config dir (e.g. ~/Library/Application Support on macOS)
// and fall back to "pokertips.db" only if that lookup fails.
func defaultDBPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "pokertips.db"
	}
	appDir := filepath.Join(dir, "pokertips")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "pokertips.db"
	}
	return filepath.Join(appDir, "pokertips.db")
}
