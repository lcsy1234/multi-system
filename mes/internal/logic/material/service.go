package material

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"multi-system/mes/internal/asynctask"
	"multi-system/mes/internal/model"
	"multi-system/mes/utility/worker"
)

var decimal = regexp.MustCompile(`^(0|[1-9][0-9]{0,8})(\.[0-9]{1,4})?$`)

type Service struct {
	DB     *sql.DB
	Worker *worker.Producer
}

func Validate(s model.Submission) error {
	if _, err := uuid.Parse(s.ClientRequestID); err != nil {
		return errors.New("请求标识必须为 UUID")
	}
	found := false
	for _, p := range model.Processes {
		if p.ID == s.ProcessID && p.Status != "已完工" {
			found = true
		}
	}
	if !found {
		return errors.New("产程不存在或已完工")
	}
	found = false
	for _, w := range model.Warehouses {
		if w.ID == s.WarehouseID {
			found = true
		}
	}
	if !found {
		return errors.New("请选择有效仓库")
	}
	if _, err := time.Parse("2006-01-02", s.Date); err != nil {
		return errors.New("日期格式应为 YYYY-MM-DD")
	}
	if len(s.Remark) > 1000 || len(s.Items) == 0 || len(s.Items) > 50 {
		return errors.New("请填写 1–50 条物料，备注不超过 1000 字节")
	}
	seen := map[string]bool{}
	for _, line := range s.Items {
		found = false
		for _, m := range model.Materials {
			if m.Code == line.MaterialCode {
				found = true
			}
		}
		if !found || seen[line.MaterialCode] {
			return errors.New("物料不存在或重复")
		}
		seen[line.MaterialCode] = true
		amount, ok := new(big.Rat).SetString(line.Quantity)
		if !decimal.MatchString(line.Quantity) || !ok || amount.Sign() <= 0 {
			return errors.New("申请数量必须大于零，最多 9 位整数和 4 位小数")
		}
		if len(line.Remark) > 500 {
			return errors.New("物料备注过长")
		}
	}
	return nil
}
func Now() string { return time.Now().UTC().Format(time.RFC3339) }
func AddEvent(r *model.Record, message string) {
	r.UpdatedAt = Now()
	r.Events = append(r.Events, model.Event{Time: r.UpdatedAt, Message: message})
	if len(r.Events) > 30 {
		r.Events = r.Events[len(r.Events)-30:]
	}
}
func digest(s model.Submission) string {
	b, _ := json.Marshal(s)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Create 校验并保存领料申请，再将申请 ID 加入异步队列。
// MySQL 保存和 Redis 入队是两个独立操作，本演示未实现事务消息。
func (s *Service) Create(ctx context.Context, input model.Submission) (*model.Record, error) {
	if err := Validate(input); err != nil {
		return nil, err
	}

	application := prepareRequisition(input)
	if err := s.CreateRecord(ctx, application); err != nil {
		return nil, err
	}

	// 重复提交时，读取数据库中已有的状态，已消费的申请不再入队。
	savedApplication, err := s.Get(ctx, application.ID)
	if err != nil {
		return nil, err
	}
	if savedApplication.Status == "consumed" {
		return savedApplication, nil
	}

	payload := model.MaterialRequisitionConsumeInfo{ApplicationID: application.ID}
	if enqueueErr := s.Worker.EnqueueTask(ctx, asynctask.EventMaterialRequisitionConsume, application.ID, payload); enqueueErr != nil {
		message := "申请已保存，但任务入队失败：" + enqueueErr.Error()
		if statusErr := s.SetStatus(ctx, application.ID, "enqueue_failed", message); statusErr != nil {
			return nil, statusErr
		}
	}

	// 消费者可能已经处理任务，返回数据库中的最新状态。
	return s.Get(ctx, application.ID)
}

// prepareRequisition 接收已通过 Validate 的数据，统一数量精度并生成申请信息。
func prepareRequisition(input model.Submission) *model.Record {
	for i := range input.Items {
		amount, _ := new(big.Rat).SetString(input.Items[i].Quantity)
		input.Items[i].Quantity = amount.FloatString(4)
	}

	date := time.Now().Format("20060102")
	requestID := strings.ReplaceAll(input.ClientRequestID, "-", "")
	application := &model.Record{
		Submission: input,
		ID:         input.ClientRequestID,
		Number:     "LLSQ-" + date + "-" + requestID,
		Status:     "queued",
		CreatedAt:  Now(),
	}
	AddEvent(application, "MES 已保存申请，准备直接加入异步队列")
	return application
}
