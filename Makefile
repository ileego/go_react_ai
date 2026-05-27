.PHONY: help env up down restart logs ps backend frontend test migrate

# 默认显示帮助
help:
	@echo "可用命令:"
	@echo "  make env       - 生成 .env 文件（从模板）"
	@echo "  make up        - 启动所有基础设施 (Docker Compose)"
	@echo "  make down      - 停止所有基础设施"
	@echo "  make restart   - 重启基础设施"
	@echo "  make logs      - 查看基础设施日志"
	@echo "  make ps        - 查看运行状态"
	@echo "  make backend   - 运行后端开发服务器"
	@echo "  make frontend  - 运行前端开发服务器"
	@echo "  make test      - 运行后端测试"
	@echo "  make migrate   - 执行数据库迁移"
	@echo "  make install   - 安装前后端依赖"

# 生成 .env 文件
env:
	@if not exist .env (copy .env.example .env && echo "已生成 .env，请编辑填入你的配置")
	@if exist .env (echo ".env 已存在，跳过")

# 启动基础设施
up:
	docker compose up -d
	@echo "等待服务就绪..."
	@timeout /t 5 /nobreak >nul
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
	cd src/backend && go mod download
	cd src/frontend && pnpm install

# 运行后端
backend:
	cd src/backend && go run cmd/server/main.go

# 运行前端
frontend:
	cd src/frontend && pnpm dev

# 运行测试
test:
	cd src/backend && go test -v ./...

# 数据库迁移（需要安装 golang-migrate）
migrate:
	migrate -path src/backend/internal/repository/migrations \
		-database "postgres://goai:goai_dev@localhost:5432/goai?sslmode=disable" up

# 创建新的迁移文件
migrate-new:
	@read -p "迁移名称: " name; \
	migrate create -ext sql -dir src/backend/internal/repository/migrations $$name
