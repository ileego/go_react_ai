package domain

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// Embedding 表示一条文本片段及其向量。
type Embedding struct {
	ID         int64
	ReportID   int64
	ChunkIndex int
	Content    string
	SourceType string
	SourceID   *int64
	Metadata   map[string]any
	Vector     pgvector.Vector
	CreatedAt  time.Time
}
