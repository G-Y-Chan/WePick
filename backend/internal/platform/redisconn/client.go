// Package redisconn contains pure Redis connection bootstrap. It is separate
// from the redisrepo adapter so the composition root can create a client once
// and hand it to whichever adapters need it.
package redisconn

import (
	"github.com/redis/go-redis/v9"

	"backend/config"
)

// New builds a go-redis client from the supplied Redis configuration.
// It is the renamed, relocated equivalent of backend/infra.NewRedisClient.
func New(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: cfg.Protocol,
	})
}
