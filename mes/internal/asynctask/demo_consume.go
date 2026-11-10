package asynctask

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

// DemoConsumeProcessor 是 Demo:Consume 任务被消费时执行的回调函数。
func DemoConsumeProcessor(ctx context.Context, task *asynq.Task) error {
	fmt.Println("异步任务消费成功")
	return nil
}
