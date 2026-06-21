-- 初始化脚本，在 PostgreSQL 首次启动时执行。
-- pgvector 扩展必须在创建任何向量列之前启用。

CREATE EXTENSION IF NOT EXISTS vector;
