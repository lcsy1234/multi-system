// Package boot wires runtime dependencies explicitly; it has no init side effects.
package boot

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
	"multi-system/mes/internal/asynctask"
	"multi-system/mes/internal/listener"
	"multi-system/mes/internal/logic/material"
	"multi-system/mes/internal/storage"
	"multi-system/mes/utility/worker"
	"os"
)

func Value(ctx context.Context, env, key, fallback string) string {
	if value := os.Getenv(env); value != "" {
		return value
	}
	return g.Cfg().MustGet(ctx, key, fallback).String()
}
func New(ctx context.Context) (*material.Service, error) {
	db, err := storage.Open()
	if err != nil {
		return nil, err
	}
	address := Value(ctx, "REDIS_ADDR", "demo.redisAddress", "127.0.0.1:6379")
	password := Value(ctx, "REDIS_PASSWORD", "demo.redisPassword", "")
	return &material.Service{
		DB: db,
		Worker: worker.NewProducer(
			asynq.RedisClientOpt{Addr: address, Password: password},
			asynctask.MaterialQueue,
		),
	}, nil
}
func Close(service *material.Service) {
	service.DB.Close()
	service.Worker.Close()
}
func Worker(ctx context.Context) error {
	service, err := New(ctx)
	if err != nil {
		return err
	}
	defer Close(service)
	if err := service.DB.PingContext(ctx); err != nil {
		return err
	}
	server := asynq.NewServer(asynq.RedisClientOpt{
		Addr:     Value(ctx, "REDIS_ADDR", "demo.redisAddress", "127.0.0.1:6379"),
		Password: Value(ctx, "REDIS_PASSWORD", "demo.redisPassword", ""),
	}, asynq.Config{
		Concurrency: 1,
		Queues:      map[string]int{asynctask.MaterialQueue: 1},
	})
	return server.Run(listener.NewMux(service))
}
