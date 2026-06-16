package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// AgentTaskRepository PostgreSQL 版智能体任务数据访问实现
type AgentTaskRepository struct {
	db *sql.DB
}

// NewAgentTaskRepository 创建 PostgreSQL 版 AgentTaskRepository
func NewAgentTaskRepository(db *sql.DB) *AgentTaskRepository {
	return &AgentTaskRepository{db: db}
}

// GetByReportID 根据报告 ID 获取任务列表
func (r *AgentTaskRepository) GetByReportID(ctx context.Context, reportID int64) ([]*domain.AgentTask, error) {
	rows, err := r.db.QueryContext(ctx, queryAgentTasksByReportID, reportID)
	if err != nil {
		return nil, fmt.Errorf("list agent tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tasks []*domain.AgentTask
	for rows.Next() {
		task, err := scanAgentTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agent tasks: %w", err)
	}
	return tasks, nil
}

// Create 创建任务
func (r *AgentTaskRepository) Create(ctx context.Context, task *domain.AgentTask) error {
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	return r.db.QueryRowContext(ctx, queryInsertAgentTask,
		task.ReportID,
		task.AgentRole,
		task.Status,
		task.Input,
		task.Output,
		task.CostMs,
		task.CreatedAt,
		task.UpdatedAt,
	).Scan(&task.ID)
}

// UpdateResult 更新任务结果
func (r *AgentTaskRepository) UpdateResult(ctx context.Context, id int64, output string, costMs int64) error {
	result, err := r.db.ExecContext(ctx, queryUpdateAgentTaskResult, output, costMs, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update agent task result: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// UpdateStatus 更新任务状态
func (r *AgentTaskRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateAgentTaskStatus, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update agent task status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// agentTaskScanner 统一扫描接口
type agentTaskScanner interface {
	Scan(dest ...any) error
}

func scanAgentTask(scanner agentTaskScanner) (*domain.AgentTask, error) {
	var task domain.AgentTask
	err := scanner.Scan(
		&task.ID,
		&task.ReportID,
		&task.AgentRole,
		&task.Status,
		&task.Input,
		&task.Output,
		&task.CostMs,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("scan agent task: %w", err)
	}
	return &task, nil
}

// SQL 常量
const (
	queryAgentTasksByReportID = `
		SELECT id, report_id, agent_role, status, input, output, cost_ms, created_at, updated_at
		FROM agent_tasks
		WHERE report_id = $1
		ORDER BY created_at ASC
	`

	queryInsertAgentTask = `
		INSERT INTO agent_tasks (report_id, agent_role, status, input, output, cost_ms, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	queryUpdateAgentTaskResult = `
		UPDATE agent_tasks
		SET output = $1, cost_ms = $2, updated_at = $3
		WHERE id = $4
	`

	queryUpdateAgentTaskStatus = `
		UPDATE agent_tasks
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
)
