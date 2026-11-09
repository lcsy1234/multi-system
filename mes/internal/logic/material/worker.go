package material

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
)

// TaskPayload carries a reference; the application and its lines stay in MySQL.
type TaskPayload struct {
	ApplicationID string `json:"applicationId"`
}

func (s *Service) HandleTask(ctx context.Context, task *asynq.Task) error {
	var payload TaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if payload.ApplicationID == "" {
		return fmt.Errorf("%w: applicationId is required", asynq.SkipRetry)
	}
	application, err := s.Get(ctx, payload.ApplicationID)
	if err != nil {
		return err
	}
	if application.Status == "consumed" {
		return nil
	}
	// Stop here for lesson one: query the application and print it. No SCM call.
	log.Printf("[领料任务消费] applicationId=%s number=%s process=%s items=%d", application.ID, application.Number, application.ProcessID, len(application.Items))
	return s.SetStatus(ctx, application.ID, "consumed", "worker 已读取申请并打印消费日志")
}
