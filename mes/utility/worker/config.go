package worker

import (
	"context"
	"fmt"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
)

const GroupDefault = "default"

type serverConfig struct {
	Queues      map[string]int
	Concurrency int
}

type groupConfig struct {
	Redis   asynq.RedisClientOpt
	Servers []serverConfig
}

func getVal(fallback string, groups ...string) string {
	if len(groups) > 0 && groups[0] != "" {
		return groups[0]
	}
	return fallback
}

// RedisConfig 统一读取指定组的 Redis，环境变量优先于配置文件。
func RedisConfig(ctx context.Context, groups ...string) asynq.RedisClientOpt {
	prefix := "worker.groups." + getVal(GroupDefault, groups...) + ".redis."
	return asynq.RedisClientOpt{
		Addr:     configValue(ctx, "REDIS_ADDR", prefix+"addr", "127.0.0.1:6379"),
		Password: configValue(ctx, "REDIS_PASSWORD", prefix+"password", ""),
		DB:       g.Cfg().MustGet(ctx, prefix+"db", 0).Int(),
	}
}

func configValue(ctx context.Context, env, key, fallback string) string {
	if value, exists := os.LookupEnv(env); exists {
		return value
	}
	return g.Cfg().MustGet(ctx, key, fallback).String()
}

// getConfig 只维护 worker.groups 一套配置，启动前统一校验。
func getConfig(ctx context.Context, group string) (groupConfig, error) {
	var cfg groupConfig
	key := "worker.groups." + group
	value, err := g.Cfg().Get(ctx, key)
	if err != nil {
		return cfg, err
	}
	if value == nil || value.IsNil() {
		return cfg, fmt.Errorf("缺少 %s 配置", key)
	}
	if err := value.Scan(&cfg); err != nil {
		return cfg, err
	}
	cfg.Redis = RedisConfig(ctx, group)
	if len(cfg.Servers) == 0 {
		return cfg, fmt.Errorf("%s.servers 不能为空", key)
	}
	for _, server := range cfg.Servers {
		if server.Concurrency <= 0 || len(server.Queues) == 0 {
			return cfg, fmt.Errorf("%s 的并发数必须大于 0，队列不能为空", key)
		}
		for queue, weight := range server.Queues {
			if queue == "" || weight <= 0 {
				return cfg, fmt.Errorf("%s 的队列名称不能为空，权重必须大于 0", key)
			}
		}
	}
	return cfg, nil
}
