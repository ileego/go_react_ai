package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/pgvector/pgvector-go"
)

// EmbeddingRepository 提供报告片段向量的增删查。
type EmbeddingRepository struct {
	db *sql.DB
}

// NewEmbeddingRepository 创建 PostgreSQL 版 EmbeddingRepository。
func NewEmbeddingRepository(db *sql.DB) *EmbeddingRepository {
	return &EmbeddingRepository{db: db}
}

// Create 批量插入或更新某报告的向量记录。
// sourceType 为默认值 "report"；如需更细粒度的来源控制，可后续扩展 CreateWithMetadata。
func (r *EmbeddingRepository) Create(ctx context.Context, reportID int64, chunks []string, vectors [][]float32) error {
	if len(chunks) != len(vectors) {
		return fmt.Errorf("chunks and vectors length mismatch")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO report_embeddings (report_id, chunk_index, content, source_type, embedding, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (report_id, chunk_index) DO UPDATE SET
			content = EXCLUDED.content,
			source_type = EXCLUDED.source_type,
			embedding = EXCLUDED.embedding,
			created_at = EXCLUDED.created_at
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now()
	for i, content := range chunks {
		if _, err := stmt.ExecContext(ctx, reportID, i, content, "report", pgvector.NewVector(vectors[i]), now); err != nil {
			return fmt.Errorf("insert embedding %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// SearchSimilar 在指定报告内搜索与 queryVector 最相似的片段。
func (r *EmbeddingRepository) SearchSimilar(ctx context.Context, reportID int64, queryVector []float32, limit int) ([]domain.Embedding, error) {
	rows, err := r.db.QueryContext(ctx, querySimilarEmbeddings, reportID, pgvector.NewVector(queryVector), limit)
	if err != nil {
		return nil, fmt.Errorf("query similar embeddings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []domain.Embedding
	for rows.Next() {
		var e domain.Embedding
		var v pgvector.Vector
		var sourceID sql.NullInt64
		var metadataJSON []byte
		if err := rows.Scan(
			&e.ID,
			&e.ReportID,
			&e.ChunkIndex,
			&e.Content,
			&e.SourceType,
			&sourceID,
			&metadataJSON,
			&v,
			&e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan embedding: %w", err)
		}
		if sourceID.Valid {
			sid := sourceID.Int64
			e.SourceID = &sid
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &e.Metadata)
		}
		e.Vector = v
		results = append(results, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate embeddings: %w", err)
	}
	return results, nil
}

// DeleteByReport 删除某报告的所有向量。
func (r *EmbeddingRepository) DeleteByReport(ctx context.Context, reportID int64) error {
	if _, err := r.db.ExecContext(ctx, deleteEmbeddingsByReport, reportID); err != nil {
		return fmt.Errorf("delete embeddings: %w", err)
	}
	return nil
}

const querySimilarEmbeddings = `
	SELECT id, report_id, chunk_index, content, source_type, source_id, metadata, embedding, created_at
	FROM report_embeddings
	WHERE report_id = $1
	ORDER BY embedding <=> $2
	LIMIT $3
`

const deleteEmbeddingsByReport = `
	DELETE FROM report_embeddings WHERE report_id = $1
`
