package cmd

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/google/uuid"
	"multi-system/mes/internal/asynctask"
	"multi-system/mes/utility/worker"
)

var AsyncDemo = gcmd.Command{
	Name:  "async-demo",
	Usage: "mes async-demo",
	Brief: "投递一个打印消费成功的异步 Demo 任务",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		producer := worker.NewProducer(worker.RedisConfig(ctx), asynctask.MaterialQueue)
		defer producer.Close()

		taskID := uuid.NewString()
		if err := producer.EnqueueTask(ctx, asynctask.EventDemoConsume, taskID, struct{}{}); err != nil {
			return err
		}
		fmt.Printf("Demo 任务已入队：event=%s taskID=%s\n", asynctask.EventDemoConsume, taskID)
		return nil
	},
}
