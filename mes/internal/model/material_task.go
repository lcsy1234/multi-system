package model

// MaterialRequisitionConsumeInfo 只传递申请 ID，完整单据由消费者从 MySQL 读取。
type MaterialRequisitionConsumeInfo struct {
	ApplicationID string `json:"applicationId"`
}
