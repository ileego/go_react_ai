package memory

import (
	"context"
	"testing"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

func TestAgentTaskRepository_CreateAndGet(t *testing.T) {
	repo := NewAgentTaskRepository()
	ctx := context.Background()

	task := &domain.AgentTask{
		ReportID:  1,
		AgentRole: domain.AgentRoleMaster,
		Status:    "running",
		Input:     "test input",
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if task.ID == 0 {
		t.Error("task id should be assigned")
	}

	tasks, err := repo.GetByReportID(ctx, 1)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Input != "test input" {
		t.Errorf("unexpected input: %s", tasks[0].Input)
	}
}

func TestAgentTaskRepository_UpdateResult(t *testing.T) {
	repo := NewAgentTaskRepository()
	ctx := context.Background()

	task := &domain.AgentTask{ReportID: 1, AgentRole: domain.AgentRoleMaster, Status: "running"}
	_ = repo.Create(ctx, task)

	if err := repo.UpdateResult(ctx, task.ID, "output", 100); err != nil {
		t.Fatalf("update result failed: %v", err)
	}

	tasks, _ := repo.GetByReportID(ctx, 1)
	if tasks[0].Output != "output" || tasks[0].CostMs != 100 {
		t.Errorf("unexpected task: %+v", tasks[0])
	}
}

func TestAgentTaskRepository_UpdateStatus(t *testing.T) {
	repo := NewAgentTaskRepository()
	ctx := context.Background()

	task := &domain.AgentTask{ReportID: 1, AgentRole: domain.AgentRoleMaster, Status: "running"}
	_ = repo.Create(ctx, task)

	if err := repo.UpdateStatus(ctx, task.ID, "completed"); err != nil {
		t.Fatalf("update status failed: %v", err)
	}

	tasks, _ := repo.GetByReportID(ctx, 1)
	if tasks[0].Status != "completed" {
		t.Errorf("expected status completed, got %s", tasks[0].Status)
	}
}

func TestAgentTaskRepository_UpdateResult_NotFound(t *testing.T) {
	repo := NewAgentTaskRepository()
	ctx := context.Background()

	err := repo.UpdateResult(ctx, 999, "output", 100)
	if err == nil {
		t.Error("expected error for non-existent task")
	}
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
