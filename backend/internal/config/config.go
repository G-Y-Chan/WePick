package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Port         string
	Redis        RedisConfig
	GooglePlaces GooglePlacesConfig
	AllowedOrigins []string
	Env          string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Protocol int
}

type GooglePlacesConfig struct {
	APIKey string
}

func LoadEnv() Config {
	port := os.Getenv("PORT")

	addr := os.Getenv("REDIS_ADDR")

	password := os.Getenv("REDIS_PASSWORD")

	db := 0
	if s := os.Getenv("REDIS_DB"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			db = v
		}
	}

	protocol := 2
	if s := os.Getenv("REDIS_PROTOCOL"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			protocol = v
		}
	}

	apiKey := os.Getenv("GOOGLE_PLACES_API_KEY")

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	var allowedOrigins []string
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		// comma-separated list
		for _, o := range splitAndTrim(origins, ",") {
			allowedOrigins = append(allowedOrigins, o)
		}
	} else {
		// dev defaults
		allowedOrigins = []string{
			"http://localhost:8081",
			"http://localhost:19006",
		}
	}

	slog.Info("config loaded", "port", port, "env", env, "allowed_origins", allowedOrigins)

	return Config{
		Port: port,
		Redis: RedisConfig{
			Addr:     addr,
			Password: password,
			DB:       db,
			Protocol: protocol,
		},
		GooglePlaces: GooglePlacesConfig{
			APIKey: apiKey,
		},
		AllowedOrigins: allowedOrigins,
		Env:            env,
	}
}

func splitAndTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}