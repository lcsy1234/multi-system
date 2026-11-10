package worker

import "github.com/hibiken/asynq"

// StartConsumer 启动独立消费服务并阻塞，收到退出信号后由 Asynq 优雅关闭。
// handler 由业务层注册事件与 Processor，worker 包只负责运行消费服务。
func StartConsumer(redis asynq.RedisClientOpt, config asynq.Config, handler asynq.Handler) error {
	server := asynq.NewServer(redis, config)
	return server.Run(handler)
}
