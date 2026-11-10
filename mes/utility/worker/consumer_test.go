package worker

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/hibiken/asynq"
)

func TestMiddlewareLogger(t *testing.T) {
	const traceID = "0123456789abcdef0123456789abcdef"
	businessErr := errors.New("业务执行失败")
	tests := []struct {
		name        string
		payload     string
		wantTraceID string
		handlerErr  error
		invalid     bool
	}{
		{name: "生成链路ID", payload: `{}`},
		{name: "空链路ID生成新值", payload: `{"traceId":""}`},
		{name: "复用消息链路ID", payload: `{"traceId":"` + traceID + `"}`, wantTraceID: traceID},
		{name: "业务错误原样返回", payload: `{}`, handlerErr: businessErr},
		{name: "非法JSON不执行业务", payload: `{`, invalid: true},
		{name: "非对象消息不执行业务", payload: `null`, invalid: true},
		{name: "非字符串链路ID不panic", payload: `{"traceId":123}`, invalid: true},
		{name: "非法链路ID不执行业务", payload: `{"traceId":"bad"}`, invalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
				called = true
				got := gtrace.GetTraceID(ctx)
				if tt.wantTraceID != "" && got != tt.wantTraceID {
					t.Fatalf("traceId = %q，期望 %q", got, tt.wantTraceID)
				}
				if _, err := hex.DecodeString(got); err != nil || len(got) != 32 {
					t.Fatalf("无效 traceId：%q", got)
				}
				return tt.handlerErr
			})
			err := middlewareLogger(next).ProcessTask(context.Background(), asynq.NewTask("Test:Consume", []byte(tt.payload)))
			if tt.invalid {
				if err == nil || called {
					t.Fatal("无效消息应返回错误，且不能执行业务")
				}
				return
			}
			if !called || !errors.Is(err, tt.handlerErr) {
				t.Fatalf("called=%v，err=%v", called, err)
			}
		})
	}
}
