# Go + React + AI 多智能体系统

基于 Go 1.26.3、React 19 和 AI 大模型的深度研究与报告生成平台。

## 技术栈

- **后端**: Go 1.26.3 + Gin + 原生 SQL (pgx)
- **数据库**: PostgreSQL 16 + pgvector
- **缓存**: Redis 7
- **文件存储**: MinIO
- **前端**: React 19 + TypeScript + Vite
- **AI**: OpenAI / Claude / DeepSeek / Kimi
- **部署**: Docker Compose (开发) / Kubernetes (生产)

## 快速开始

```bash
# 1. 克隆项目
git clone https://github.com/ileego/go_react_ai.git
cd go_react_ai/src

# 2. 生成后端环境文件并编辑
make backend-env
# 编辑 backend/.env 填入你的 AI API Keys

# 3. 安装依赖（前后端完全独立，顺序无关）
make backend-install
make frontend-install

# 4. 启动基础设施（PostgreSQL + Redis + MinIO）
make up

# 5. 启动后端
make backend-run

# 6. 启动前端（新终端）
make frontend-dev
```

访问 http://localhost:5173 查看前端页面，后端运行在 http://localhost:8080，API 文档在 http://localhost:8080/swagger。

## 项目结构

```
src/
  backend/          # Go 后端
    cmd/server/     # 服务入口
    internal/       # 私有业务代码
      config/       # 配置加载
      domain/       # 领域模型
      handler/      # HTTP 接口层
      service/      # 业务逻辑层
      repository/   # 数据访问层
    pkg/            # 公共库
    docs/           # API 文档 (OpenAPI)
  frontend/         # React 前端
    src/features/   # 功能模块
    src/shared/     # 共享组件和工具
  Makefile          # 聚合命令
docs/               # 设计文档、ADR
book/               # 书籍源码
```

## 开发规范

- **后端分层**: Handler -> Service -> Repository，Domain 位于最内层
- **接口契约**: 前后端以 `docs/openapi.yaml` 为准
- **Git 工作流**: GitHub Flow + Conventional Commits
- **代码规范**: `make lint` 自动检查

## 书籍

《全栈开发实战：Go + React + AI 多智能体系统构建》配套源码，详见 `book/` 目录。
