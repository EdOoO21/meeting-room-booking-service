package settings

import (
	"os"
	"strconv"
	"strings"
)

const defaultHTTPPort = 8080

type Config struct {
	HTTP HTTPConfig
}

type HTTPConfig struct {
	Port int
}

func NewConfig() Config {
	return Config{
		HTTP: HTTPConfig{
			Port: getInt("APP_HTTP_PORT", defaultHTTPPort),
		},
	}
}

func getInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
