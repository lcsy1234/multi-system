// Package taskqueue 封装异步任务的序列化和 Redis 入队操作。
package taskqueue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hibiken/asynq"
)

type Publisher struct {
	client *asynq.Client
	queue  string
}

func NewPublisher(redis asynq.RedisClientOpt, queue string) *Publisher {
	return &Publisher{
		client: asynq.NewClient(redis),
		queue:  queue,
	}
}

// Push 将业务数据序列化为 JSON 并推入队列，不等待消费者处理。
// taskID 用于队列内去重；相同 ID 的任务已存在时，按已入队处理。
func (p *Publisher) Push(ctx context.Context, taskType, taskID string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, data)
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

func (p *Publisher) Close() error {
	return p.client.Close()
}
