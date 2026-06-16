package domain

import "time"

// AgentRole 智能体角色
type AgentRole string

const (
	AgentRoleMaster   AgentRole = "master"   // 调度器
	AgentRoleResearch AgentRole = "research" // 研究员
	AgentRoleWriter   AgentRole = "writer"   // 写作员
	AgentRoleReview   AgentRole = "review"   // 审校员
)

// AgentTask 表示智能体执行的任务
type AgentTask struct {
	ID        int64
	ReportID  int64
	AgentRole AgentRole
	Status    string // pending / running / completed / failed
	Input     string // 输入提示词
	Output    string // 输出结果
	CostMs    int64  // 耗时（毫秒）
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TaskResult 任务执行结果
type TaskResult struct {
	Output string
	Error  error
	CostMs int64
}
