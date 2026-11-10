package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"multi-system/mes/internal/logic/material"
	"multi-system/mes/internal/model"
)

// MaterialRequisitionConsumeProcessor 解析队列消息并调用领料业务方法。
type MaterialRequisitionConsumeProcessor struct {
	service *material.Service
}

func (p *MaterialRequisitionConsumeProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var input model.MaterialRequisitionConsumeInfo
	if err := json.Unmarshal(task.Payload(), &input); err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if input.ApplicationID == "" {
		return fmt.Errorf("%w: applicationId is required", asynq.SkipRetry)
	}
	return p.service.Consume(ctx, input.ApplicationID)
}
