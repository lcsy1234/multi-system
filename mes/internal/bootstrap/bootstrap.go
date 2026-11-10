// Package bootstrap 组装服务依赖，不在包初始化时连接外部服务。
package bootstrap

import (
	"context"
	"multi-system/mes/internal/asynctask"
	"multi-system/mes/internal/logic/material"
	"multi-system/mes/internal/storage"
	"multi-system/mes/utility/worker"
)

func New(ctx context.Context) (*material.Service, error) {
	db, err := storage.Open()
	if err != nil {
		return nil, err
	}
	return &material.Service{
		DB: db,
		Worker: worker.NewProducer(
			worker.RedisConfig(ctx),
			asynctask.MaterialQueue,
		),
	}, nil
}
func Close(service *material.Service) {
	service.DB.Close()
	service.Worker.Close()
}
