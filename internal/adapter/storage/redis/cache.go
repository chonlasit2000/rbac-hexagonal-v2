package redis

import (
	"context"
	"fmt"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	// Ping check connection
	// ใช้ context.Background() สำหรับการ Ping ตอนเริ่มโปรแกรม
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		// อย่า panic ใน library code, return error ให้ main ตัดสินใจดีกว่า
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
