// Package service 实现业务逻辑层。
package service

import (
	"context"
	"fmt"

	"github.com/ileego/go_react_ai/internal/repository"
	"github.com/ileego/go_react_ai/pkg/errors"
)

// agentService 实现 AgentService 接口
type agentService struct {
	reportRepo repository.ReportRepository
	taskRepo   repository.AgentTaskRepository
}

// NewAgentService 创建 AgentService 实例
func NewAgentService(
	reportRepo repository.ReportRepository,
	taskRepo repository.AgentTaskRepository,
) AgentService {
	return &agentService{
		reportRepo: reportRepo,
		taskRepo:   taskRepo,
	}
}

// Dispatch 派发报告生成任务给智能体流水线
func (s *agentService) Dispatch(ctx context.Context, reportID int64) error {
	_, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return errors.NewNotFound("report", reportID)
	}
	// TODO: 实际调度逻辑在第25章实现
	fmt.Printf("dispatch report %d\n", reportID)
	return nil
}
