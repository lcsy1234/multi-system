package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// TraceIdKey 是任务 JSON 中可选的链路标识，值须为 32 位十六进制字符串。
const TraceIdKey = "traceId"

// Consumer 封装单个 Asynq 实例和任务回调路由表。
type Consumer struct {
	*asynq.Server
	mux *asynq.ServeMux
}

// StartConsumer 按组启动消费实例，组名不传时使用 default。
// 此方法在 listener 进程中阻塞；退出或启动失败时统一停止已创建的实例。
func StartConsumer(ctx context.Context, handlers map[string]func(context.Context, *asynq.Task) error, groups ...string) error {
	cfg, err := getConfig(ctx, getVal(GroupDefault, groups...))
	if err != nil {
		return err
	}
	if len(handlers) == 0 {
		return fmt.Errorf("未注册任务 Processor")
	}

	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	for _, s := range cfg.Servers {
		srv := NewConsumer(&cfg.Redis, s.Queues, s.Concurrency)
		defer srv.Stop()
		for pattern, handler := range handlers {
			srv.RegisterHandler(pattern, handler)
		}
		// Asynq.Start 已在内部启动协程，外层统一等待退出信号。
		if err := srv.Start(); err != nil {
			return fmt.Errorf("启动队列 %v 失败：%w", s.Queues, err)
		}
	}
	<-runCtx.Done()
	return nil
}

// NewConsumer 创建实例，并发数和队列权重由配置传入。
func NewConsumer(redisCfg *asynq.RedisClientOpt, queues map[string]int, concurrency int) *Consumer {
	srv := asynq.NewServer(*redisCfg, asynq.Config{
		Concurrency: concurrency,
		Queues:      queues,
	})
	return &Consumer{Server: srv, mux: asynq.NewServeMux()}
}

// RegisterHandler 绑定任务名和回调函数，必须在 Start 前完成。
func (c *Consumer) RegisterHandler(pattern string, handler func(context.Context, *asynq.Task) error) {
	c.mux.HandleFunc(pattern, handler)
}

// Start 安装统一日志中间件，然后在后台启动消费。
func (c *Consumer) Start() error {
	c.mux.Use(middlewareLogger)
	return c.Server.Start(c.mux)
}

// Stop 停止获取新任务，等待执行中的任务结束并释放连接。
func (c *Consumer) Stop() {
	c.Server.Stop()
	c.Server.Shutdown()
}

// middlewareLogger 注入 traceId，记录任务开始、执行结果和耗时。
// 消息须为 JSON 对象；未携带 traceId 时生成新的链路标识。
func middlewareLogger(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		ctx, err := taskContext(ctx, task.Payload())
		if err != nil {
			glog.Error(ctx, "任务消息无效", task.Type(), err)
			return err
		}

		glog.Info(ctx, "开始处理任务", task.Type(), string(task.Payload()))
		start := time.Now()
		err = next.ProcessTask(ctx, task)
		duration := time.Since(start)
		if err != nil {
			glog.Errorf(ctx, "%s, 任务执行失败, 耗时: %v, 错误: %v", task.Type(), duration, err)
		} else {
			glog.Infof(ctx, "%s, 任务执行成功, 耗时: %v", task.Type(), duration)
		}
		return err
	})
}

// taskContext 校验消息格式，复用或生成 traceId 后注入回调 context。
func taskContext(ctx context.Context, payload []byte) (context.Context, error) {
	var body g.Map
	if err := json.Unmarshal(payload, &body); err != nil {
		return ctx, err
	}
	if body == nil {
		return ctx, fmt.Errorf("任务消息必须为 JSON 对象")
	}
	traceID, ok := body[TraceIdKey].(string)
	if _, exists := body[TraceIdKey]; exists && !ok {
		return ctx, fmt.Errorf("%s 必须为字符串", TraceIdKey)
	}
	if traceID == "" {
		traceID = strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	return gtrace.WithTraceID(ctx, traceID)
}
