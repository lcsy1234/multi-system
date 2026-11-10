SERVICES := erp wms scm mes

.PHONY: help deps fmt test vet build dev smoke demo integration db-init listener socket worker async-demo frontend-dev frontend-build

help:
	@echo "deps  下载依赖 | fmt  格式化 | test  编译检查/运行测试 | vet  静态检查"
	@echo "build 编译四个服务 | dev 启动四个服务 | smoke 启动并验证四个服务"
	@echo "demo 启动 Redis + 四个服务（不启动消费） | integration 验证真实异步链路"
	@echo "listener 启动独立消费进程 | worker 为 listener 的兼容别名"
	@echo "socket 启动独立 WebSocket 演示服务（8005）"
	@echo "async-demo 投递一个打印消费成功的 Demo 任务"

deps:
	@set -e; for service in $(SERVICES); do (cd $$service && GOWORK=off go mod download); done

fmt:
	@gofmt -w erp wms scm mes

test:
	@set -e; for service in $(SERVICES); do (cd $$service && GOWORK=off go test ./...); done

vet:
	@set -e; for service in $(SERVICES); do (cd $$service && GOWORK=off go vet ./...); done

build:
	@mkdir -p .bin
	@set -e; for service in $(SERVICES); do (cd $$service && GOWORK=off go build -o ../.bin/$$service .); done
	cd mes && GOWORK=off go build -o ../.bin/mes-listener ./boot/listener.go
	cd mes && GOWORK=off go build -o ../.bin/mes-socket ./boot/socket.go

dev: build
	python3 scripts/services.py run

smoke: build
	python3 scripts/services.py smoke

db-init: build
	cd mes && ../.bin/mes migrate

demo: frontend-build db-init
	python3 scripts/demo.py

integration: build
	python3 scripts/integration.py

listener: build
	cd mes && REDIS_ADDR=$${REDIS_ADDR:-127.0.0.1:6380} ../.bin/mes-listener

worker: listener

socket: build
	cd mes && ../.bin/mes-socket

async-demo: build
	cd mes && REDIS_ADDR=$${REDIS_ADDR:-127.0.0.1:6380} ../.bin/mes async-demo

frontend-build:
	@test -d mes-web/node_modules || npm --prefix mes-web ci
	npm --prefix mes-web run build

frontend-dev:
	npm --prefix mes-web run dev
