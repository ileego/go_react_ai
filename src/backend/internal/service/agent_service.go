// Package service 实现业务逻辑层。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/internal/repository"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
	"github.com/ileego/go_react_ai/pkg/httpx"
	"github.com/ileego/go_react_ai/pkg/worker"
)

// agentService 实现 AgentService 接口
type agentService struct {
	reportSvc  ReportService
	taskRepo   repository.AgentTaskRepository
	workerPool *worker.Pool
	httpClient *httpx.Client
	aiEndpoint string
}

// NewAgentService 创建 AgentService 实例
func NewAgentService(
	reportSvc ReportService,
	taskRepo repository.AgentTaskRepository,
	workerPool *worker.Pool,
	httpClient *httpx.Client,
	aiEndpoint string,
) AgentService {
	return &agentService{
		reportSvc:  reportSvc,
		taskRepo:   taskRepo,
		workerPool: workerPool,
		httpClient: httpClient,
		aiEndpoint: aiEndpoint,
	}
}

// Dispatch 将报告生成任务异步提交到 Worker Pool。
// 调用方会立即收到"任务已提交"响应，实际执行在后台 goroutine 中完成。
func (s *agentService) Dispatch(ctx context.Context, reportID int64) error {
	report, err := s.reportSvc.GetByID(ctx, reportID)
	if err != nil {
		return err
	}

	if report.Status != domain.ReportStatusPending {
		return apperrors.NewValidation("status", fmt.Sprintf("当前状态 %s 不允许派发", report.Status)).
			WithCode("REPORT_NOT_PENDING")
	}

	if err := s.reportSvc.UpdateStatus(ctx, reportID, domain.ReportStatusRunning); err != nil {
		return apperrors.NewInternal("更新报告状态失败", err)
	}

	job := &reportGenerationJob{
		reportID:   reportID,
		title:      report.Title,
		topic:      report.Topic,
		aiEndpoint: s.aiEndpoint,
		reportSvc:  s.reportSvc,
		taskRepo:   s.taskRepo,
		httpClient: s.httpClient,
		logger:     middleware.GetLoggerFromContext(ctx).With("report_id", reportID),
	}

	if err := s.workerPool.SubmitBlocking(ctx, job); err != nil {
		// 提交失败，回滚报告状态，避免任务丢失
		_ = s.reportSvc.UpdateStatus(ctx, reportID, domain.ReportStatusPending)
		return apperrors.NewInternal("任务提交失败", err)
	}

	return nil
}

// reportGenerationJob 实现 worker.Job 接口，负责执行报告生成流水线。
type reportGenerationJob struct {
	reportID   int64
	title      string
	topic      string
	aiEndpoint string
	reportSvc  ReportService
	taskRepo   repository.AgentTaskRepository
	httpClient *httpx.Client
	logger     *slog.Logger
}

func (j *reportGenerationJob) ID() string {
	return fmt.Sprintf("report-%d", j.reportID)
}

// Execute 执行任务流水线：创建 agent_task → 调用 AI → 更新结果 → 更新报告状态。
func (j *reportGenerationJob) Execute(ctx context.Context) error {
	j.logger.Info("开始执行报告生成任务")
	start := time.Now()

	task := &domain.AgentTask{
		ReportID:  j.reportID,
		AgentRole: domain.AgentRoleMaster,
		Status:    "running",
		Input:     fmt.Sprintf("标题：%s；主题：%s", j.title, j.topic),
	}
	if err := j.taskRepo.Create(ctx, task); err != nil {
		j.logger.Error("创建任务记录失败", "error", err)
		j.markReportFailed(ctx)
		return err
	}

	output, err := j.callAI(ctx, task.Input)
	costMs := time.Since(start).Milliseconds()

	if err != nil {
		j.logger.Error("AI 调用失败", "error", err, "cost_ms", costMs)
		_ = j.taskRepo.UpdateResult(ctx, task.ID, "", costMs)
		_ = j.taskRepo.UpdateStatus(ctx, task.ID, "failed")
		j.markReportFailed(ctx)
		return err
	}

	if err := j.taskRepo.UpdateResult(ctx, task.ID, output, costMs); err != nil {
		j.logger.Error("更新任务结果失败", "error", err)
		j.markReportFailed(ctx)
		return err
	}
	if err := j.taskRepo.UpdateStatus(ctx, task.ID, "completed"); err != nil {
		j.logger.Error("更新任务状态失败", "error", err)
	}

	if err := j.reportSvc.UpdateStatus(ctx, j.reportID, domain.ReportStatusCompleted); err != nil {
		j.logger.Error("更新报告状态失败", "error", err)
		return err
	}

	j.logger.Info("报告生成任务完成", "cost_ms", costMs)
	return nil
}

// callAI 调用 AI 服务生成报告内容。
// 当前使用 httpx.Client 做 HTTP 调用，失败时通过 fallback 返回兜底内容。
func (j *reportGenerationJob) callAI(ctx context.Context, input string) (string, error) {
	payload := map[string]string{"input": input}
	body, _ := json.Marshal(payload)

	resp, err := j.httpClient.PostJSON(ctx, j.aiEndpoint, body)
	if err != nil {
		return "", fmt.Errorf("AI 调用失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI 服务返回非 200: %d", resp.StatusCode)
	}

	var result struct {
		Output string `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析 AI 响应失败: %w", err)
	}
	return result.Output, nil
}

func (j *reportGenerationJob) markReportFailed(ctx context.Context) {
	_ = j.reportSvc.UpdateStatus(ctx, j.reportID, domain.ReportStatusFailed)
}
