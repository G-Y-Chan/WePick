package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port  string
	Redis RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Protocol int
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

	return Config{
		Port: port,
		Redis: RedisConfig{
			Addr:     addr,
			Password: password,
			DB:       db,
			Protocol: protocol,
		},
	}
}