package asynctask

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"multi-system/mes/internal/model"
)

// MaterialRequisitionConsumer 是 Processor 所需的业务能力，避免依赖具体 Service。
type MaterialRequisitionConsumer interface {
	Consume(ctx context.Context, applicationID string) error
}

var materialConsumer MaterialRequisitionConsumer

// Init 在启动消费前设置业务依赖，消费运行期间不再修改。
func Init(service MaterialRequisitionConsumer) {
	materialConsumer = service
}

// MaterialRequisitionConsumeProcessor 是领料任务的 handler 回调函数。
func MaterialRequisitionConsumeProcessor(ctx context.Context, task *asynq.Task) error {
	var input model.MaterialRequisitionConsumeInfo
	if err := json.Unmarshal(task.Payload(), &input); err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if input.ApplicationID == "" {
		return fmt.Errorf("%w: applicationId is required", asynq.SkipRetry)
	}
	if materialConsumer == nil {
		return fmt.Errorf("领料任务消费依赖未初始化，请先调用 asynctask.Init")
	}
	return materialConsumer.Consume(ctx, input.ApplicationID)
}
