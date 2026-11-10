// Package asynctask 定义生产端和消费端共用的事件名称。
package asynctask

const MaterialQueue = "mes-material"

// EventMaterialRequisitionConsume 表示读取并消费领料申请，不表示已经发送到 SCM。
// 保留已有事件值，已在 Redis 中排队的任务可继续被消费。
const EventMaterialRequisitionConsume = "mes:material:consume"
