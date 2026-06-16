// Package memory 提供基于内存的 Repository 实现，用于测试和开发阶段。
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// AgentTaskRepository 内存版智能体任务数据访问实现
type AgentTaskRepository struct {
	mu       sync.RWMutex
	tasks    map[int64]*domain.AgentTask
	byReport map[int64][]int64
	nextID   int64
}

// NewAgentTaskRepository 创建内存版 AgentTaskRepository
func NewAgentTaskRepository() *AgentTaskRepository {
	return &AgentTaskRepository{
		tasks:    make(map[int64]*domain.AgentTask),
		byReport: make(map[int64][]int64),
		nextID:   1,
	}
}

// GetByReportID 根据报告 ID 获取任务列表
func (r *AgentTaskRepository) GetByReportID(_ context.Context, reportID int64) ([]*domain.AgentTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.AgentTask
	for _, id := range r.byReport[reportID] {
		if task, ok := r.tasks[id]; ok {
			result = append(result, copyAgentTask(task))
		}
	}
	return result, nil
}

// Create 创建任务
func (r *AgentTaskRepository) Create(_ context.Context, task *domain.AgentTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.nextID
	r.nextID++
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	r.tasks[task.ID] = copyAgentTask(task)
	r.byReport[task.ReportID] = append(r.byReport[task.ReportID], task.ID)
	return nil
}

// UpdateResult 更新任务结果
func (r *AgentTaskRepository) UpdateResult(_ context.Context, id int64, output string, costMs int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return repository.ErrNotFound
	}
	task.Output = output
	task.CostMs = costMs
	task.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus 更新任务状态
func (r *AgentTaskRepository) UpdateStatus(_ context.Context, id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return repository.ErrNotFound
	}
	task.Status = status
	task.UpdatedAt = time.Now()
	return nil
}

// copyAgentTask 创建任务的深拷贝
func copyAgentTask(t *domain.AgentTask) *domain.AgentTask {
	return &domain.AgentTask{
		ID:        t.ID,
		ReportID:  t.ReportID,
		AgentRole: t.AgentRole,
		Status:    t.Status,
		Input:     t.Input,
		Output:    t.Output,
		CostMs:    t.CostMs,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
