// Package boot wires runtime dependencies explicitly; it has no init side effects.
package boot

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
	"multi-system/mes/internal/asynctask"
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
	return &material.Service{
		DB: db,
		Worker: worker.NewProducer(
			RedisConfig(ctx),
			asynctask.MaterialQueue,
		),
	}, nil
}
func Close(service *material.Service) {
	service.DB.Close()
	service.Worker.Close()
}
func RedisConfig(ctx context.Context) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     Value(ctx, "REDIS_ADDR", "demo.redisAddress", "127.0.0.1:6379"),
		Password: Value(ctx, "REDIS_PASSWORD", "demo.redisPassword", ""),
	}
}

// ConsumerConfig 从 config.yaml 读取消费配置，在启动 listener 前校验。
func ConsumerConfig(ctx context.Context) (asynq.Config, error) {
	var config asynq.Config
	value, err := g.Cfg().Get(ctx, "worker.consumer")
	if err != nil {
		return config, err
	}
	if value == nil || value.IsNil() {
		return config, fmt.Errorf("缺少 worker.consumer 配置")
	}
	if err := value.Scan(&config); err != nil {
		return config, err
	}
	if config.Concurrency <= 0 {
		return config, fmt.Errorf("worker.consumer.concurrency 必须大于 0")
	}
	if len(config.Queues) == 0 {
		return config, fmt.Errorf("worker.consumer.queues 至少配置一个队列")
	}
	for queue, weight := range config.Queues {
		if queue == "" || weight <= 0 {
			return config, fmt.Errorf("worker.consumer.queues 的队列名称不能为空，权重必须大于 0")
		}
	}
	return config, nil
}
