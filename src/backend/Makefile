.PHONY: help env up down restart logs ps build run test migrate migrate-new install lint

# 跨平台可执行文件后缀
ifeq ($(OS),Windows_NT)
	EXE := .exe
else
	EXE :=
endif

# 默认显示帮助
help:
	@echo "后端可用命令:"
	@echo "  make env          - 生成 .env 文件（从模板）"
	@echo "  make up           - 启动基础设施 (Docker Compose)"
	@echo "  make down         - 停止基础设施"
	@echo "  make restart      - 重启基础设施"
	@echo "  make logs         - 查看基础设施日志"
	@echo "  make ps           - 查看运行状态"
	@echo "  make install      - 下载 Go 依赖"
	@echo "  make build        - 编译后端二进制"
	@echo "  make run          - 运行后端开发服务器"
	@echo "  make test         - 运行测试"
	@echo "  make lint         - 运行代码检查"
	@echo "  make migrate      - 执行数据库迁移"
	@echo "  make migrate-new  - 创建新的迁移文件（用法: make migrate-new NAME=xxx）"

# 生成 .env 文件
env:
	@if [ -f .env ]; then \
		echo ".env 已存在，跳过"; \
	else \
		cp .env.example .env; \
		echo "已生成 .env，请编辑填入你的配置"; \
	fi

# 启动基础设施
up:
	docker compose up -d
	@echo "等待服务就绪..."
	@sleep 5
	@docker compose ps

# 停止基础设施
down:
	docker compose down

# 重启
restart: down up

# 查看日志
logs:
	docker compose logs -f

# 查看状态
ps:
	docker compose ps

# 安装依赖
install:
	go mod download

# 编译后端
build:
	go build -o bin/server$(EXE) ./cmd/server/main.go

# 运行后端
run:
	go run ./cmd/server/main.go

# 运行测试
test:
	go test -v ./...

# 代码检查
lint:
	golangci-lint run ./...

# 数据库迁移（需要安装 golang-migrate）
migrate:
	migrate -path internal/repository/migrations \
		-database "postgres://goai:goai_dev@localhost:5432/goai?sslmode=disable" up

# 创建新的迁移文件
migrate-new:
	@if [ -z "$(NAME)" ]; then \
		echo "用法: make migrate-new NAME=xxx"; \
		exit 1; \
	fi
	migrate create -ext sql -dir internal/repository/migrations $(NAME)
