package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

// StartConsumer 启动独立消费服务并阻塞，收到退出信号后由 Asynq 优雅关闭。
// processors 由业务层提供，worker 包统一注册事件并运行消费服务。
func StartConsumer(ctx context.Context, processors map[string]func(context.Context, *asynq.Task) error) error {
	config, err := consumerConfig(ctx)
	if err != nil {
		return err
	}
	if len(processors) == 0 {
		return fmt.Errorf("未注册任务 Processor，请先初始化 asynctask.Processors")
	}

	mux := asynq.NewServeMux()
	for event, processor := range processors {
		mux.HandleFunc(event, processor)
	}

	server := asynq.NewServer(RedisConfig(ctx), config)
	return server.Run(mux)
}
