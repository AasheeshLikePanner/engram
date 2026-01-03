package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBPath    string
	GRPCPort  int
	OllamaURL string
	RedisURL  string
	StoreType string // "badger" or "redis"
	ToolsPath string
}

func Load() *Config {
	return &Config{
		DBPath:    getEnv("DB_PATH", "./data"),
		GRPCPort:  getEnvAsInt("GRPC_PORT", 50051),
		OllamaURL: getEnv("OLLAMA_URL", "http://localhost:11434"),
		RedisURL:  getEnv("REDIS_URL", "localhost:6379"),
		StoreType: getEnv("STORE_TYPE", "badger"),
		ToolsPath: getEnv("TOOLS_PATH", "./tools"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
