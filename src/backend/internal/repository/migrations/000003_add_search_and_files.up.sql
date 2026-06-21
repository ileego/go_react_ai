-- 为 reports 表添加全文搜索向量列
ALTER TABLE reports
    ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- 创建 GIN 索引，加速全文搜索
CREATE INDEX IF NOT EXISTS idx_reports_search ON reports USING GIN(search_vector);

-- 触发器函数：自动更新 search_vector
CREATE OR REPLACE FUNCTION reports_search_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.topic, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(NEW.content, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS reports_search_trigger ON reports;
CREATE TRIGGER reports_search_trigger
    BEFORE INSERT OR UPDATE ON reports
    FOR EACH ROW
    EXECUTE FUNCTION reports_search_update();

-- 初始化已有数据的 search_vector
UPDATE reports SET search_vector =
    setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(topic, '')), 'B') ||
    setweight(to_tsvector('simple', COALESCE(content, '')), 'C')
WHERE search_vector IS NULL;

-- 文件元数据表
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    storage_key VARCHAR(500) NOT NULL UNIQUE,
    content_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    bucket VARCHAR(100) NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_files_created_by ON files(created_by);
CREATE INDEX IF NOT EXISTS idx_files_storage_key ON files(storage_key);
