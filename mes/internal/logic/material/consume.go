package material

import (
	"context"
	"log"
)

// Consume 是领料任务的业务处理方法，由 listener 的 Processor 调用。
func (s *Service) Consume(ctx context.Context, applicationID string) error {
	application, err := s.Get(ctx, applicationID)
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
