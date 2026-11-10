// Package listener 注册事件与 Processor 的对应关系。
package listener

import (
	"github.com/hibiken/asynq"
	"multi-system/mes/internal/asynctask"
	"multi-system/mes/internal/logic/material"
)

func NewMux(service *material.Service) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Handle(asynctask.EventMaterialRequisitionConsume, &MaterialRequisitionConsumeProcessor{
		service: service,
	})
	return mux
}
