package config

import "os"

type Config struct {
	MongoURI  string
	MongoDB   string
	DBPath    string
	Port      string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		MongoURI:  getEnv("MONGO_URI", ""),
		MongoDB:   getEnv("MONGO_DB", "pokertips"),
		DBPath:    getEnv("DB_PATH", "pokertips.db"),
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
