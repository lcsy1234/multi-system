SERVICES := erp wms scm mes

.PHONY: help deps fmt test vet build dev smoke

help:
	@echo "deps  下载依赖 | fmt  格式化 | test  编译检查/运行测试 | vet  静态检查"
	@echo "build 编译四个服务 | dev 启动四个服务 | smoke 启动并验证四个服务"

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

dev: build
	python3 scripts/services.py run

smoke: build
	python3 scripts/services.py smoke
