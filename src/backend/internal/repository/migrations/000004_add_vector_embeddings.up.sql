-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 报告片段向量表：存储报告内容分块后的 Embedding
CREATE TABLE IF NOT EXISTS report_embeddings (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL DEFAULT 0,
    content TEXT NOT NULL DEFAULT '',
    -- 来源元数据：用于区分文本来自报告正文、上传文件还是外部网页
    source_type VARCHAR(50) NOT NULL DEFAULT 'report',
    source_id BIGINT,
    metadata JSONB NOT NULL DEFAULT '{}',
    embedding vector(1536) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(report_id, chunk_index)
);

-- 按报告 ID 查询时常用
CREATE INDEX IF NOT EXISTS idx_report_embeddings_report_id
    ON report_embeddings(report_id);

-- HNSW 余弦相似度索引，加速近似最近邻搜索
CREATE INDEX IF NOT EXISTS idx_report_embeddings_hnsw
    ON report_embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
