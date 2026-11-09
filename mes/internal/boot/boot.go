// Package boot wires runtime dependencies explicitly; it has no init side effects.
package boot

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
	"multi-system/mes/internal/logic/material"
	"multi-system/mes/internal/storage"
	"multi-system/mes/internal/taskqueue"
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
		Publisher: taskqueue.NewPublisher(
			asynq.RedisClientOpt{Addr: address, Password: password},
			material.Queue,
		),
	}, nil
}
func Close(service *material.Service) {
	service.DB.Close()
	service.Publisher.Close()
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
	server := asynq.NewServer(asynq.RedisClientOpt{Addr: Value(ctx, "REDIS_ADDR", "demo.redisAddress", "127.0.0.1:6379"), Password: Value(ctx, "REDIS_PASSWORD", "demo.redisPassword", "")}, asynq.Config{Concurrency: 1, Queues: map[string]int{material.Queue: 1}})
	mux := asynq.NewServeMux()
	mux.HandleFunc(material.TaskType, service.HandleTask)
	return server.Run(mux)
}
