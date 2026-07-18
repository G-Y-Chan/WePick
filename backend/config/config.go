package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	Redis        RedisConfig
	GooglePlaces GooglePlacesConfig // Added Google Places configuration
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
	}
}
