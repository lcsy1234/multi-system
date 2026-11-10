package asynctask

import (
	"context"

	"github.com/hibiken/asynq"
	"multi-system/mes/internal/listener"
)

const (
	MaterialQueue = "mes-material"

	// 领料申请消费；保留已有任务名，以兼容 Redis 中已排队的任务。
	EventMaterialRequisitionConsume = "mes:material:consume"
)

// Processors 定义任务名与 handler 回调函数的对应关系。
func Processors() (handlers map[string]func(context.Context, *asynq.Task) error) {
	handlers = make(map[string]func(context.Context, *asynq.Task) error)
	handlers[EventMaterialRequisitionConsume] = listener.MaterialRequisitionConsumeProcessor
	return
}
