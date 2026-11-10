// Package worker 封装异步任务生产端，不包含领料业务逻辑。
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hibiken/asynq"
)

type Producer struct {
	client *asynq.Client
	queue  string
}

func NewProducer(redis asynq.RedisClientOpt, queue string) *Producer {
	return &Producer{
		client: asynq.NewClient(redis),
		queue:  queue,
	}
}

// EnqueueTask 将事件和业务数据封装为任务并推入队列，不等待消费者处理。
// taskID 用于队列内去重；相同 ID 的任务已存在时，按已入队处理。
func (p *Producer) EnqueueTask(ctx context.Context, event, taskID string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(event, data)
	_, err = p.client.EnqueueContext(ctx, task,
		asynq.Queue(p.queue),
		asynq.TaskID(taskID),
		asynq.MaxRetry(0),             // 当前演示不自动重试。
		asynq.Retention(24*time.Hour), // 成功任务保留 24 小时。
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	return err
}

func (p *Producer) Close() error {
	return p.client.Close()
}
