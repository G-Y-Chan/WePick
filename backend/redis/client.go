package redis

package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func Init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})
}

// Wrapper: Set key with optional expiration
func SetKey(ctx context.Context, key string, value string, expiration time.Duration) error {
    return RDB.Set(ctx, key, value, expiration).Err()
}

// Wrapper: Get key
func GetKey(ctx context.Context, key string) (string, error) {
    return RDB.Get(ctx, key).Result()
}



