# 根目录聚合 Makefile
# 前后端构建完全分离，此处仅做命令转发

.PHONY: help up down restart logs ps

# 基础设施命令（转发到后端 Makefile）
up down restart logs ps:
	@$(MAKE) -C src/backend $@

# 后端命令转发: make backend-run / backend-test / backend-build ...
backend-%:
	@$(MAKE) -C src/backend $(subst backend-,,$@)

# 前端命令转发: make frontend-dev / frontend-build / frontend-install ...
frontend-%:
	@$(MAKE) -C src/frontend $(subst frontend-,,$@)

help:
	@echo "用法: make [命令]"
	@echo ""
	@echo "基础设施（由后端 Makefile 管理）:"
	@echo "  make up              - 启动 Docker Compose 基础设施"
	@echo "  make down            - 停止基础设施"
	@echo "  make restart         - 重启基础设施"
	@echo "  make logs            - 查看日志"
	@echo "  make ps              - 查看运行状态"
	@echo ""
	@echo "后端命令（make backend-xxx）:"
	@echo "  make backend-env     - 生成 .env 文件"
	@echo "  make backend-install - 下载 Go 依赖"
	@echo "  make backend-build   - 编译后端"
	@echo "  make backend-run     - 运行后端开发服务器"
	@echo "  make backend-test    - 运行后端测试"
	@echo "  make backend-lint    - 后端代码检查"
	@echo "  make backend-migrate - 执行数据库迁移"
	@echo ""
	@echo "前端命令（make frontend-xxx）:"
	@echo "  make frontend-install - 安装前端依赖"
	@echo "  make frontend-dev    - 启动前端开发服务器"
	@echo "  make frontend-build  - 前端生产构建"
	@echo "  make frontend-lint   - 前端代码检查"
	@echo "  make frontend-test   - 前端测试"
