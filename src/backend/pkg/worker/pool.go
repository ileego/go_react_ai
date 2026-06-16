// Package worker 提供有界 Worker Pool 实现。
// 支持固定数量的 worker、缓冲任务队列、优先级调度与优雅关闭。
package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Job 表示一个可异步执行的任务。
type Job interface {
	// Execute 执行任务，ctx 用于传递取消、超时等信号。
	Execute(ctx context.Context) error
	// ID 返回任务唯一标识，用于日志与监控。
	ID() string
}

const (
	// defaultWorkers 默认 worker 数量。
	defaultWorkers = 4
	// defaultQueueSize 默认队列大小。
	defaultQueueSize = 8
)

// defaultJobTimeout 单个任务默认执行超时。
// 使用 var 而非 const，便于单元测试临时调整。
var defaultJobTimeout = 5 * time.Minute

// jobContext 把任务与它提交时的上游 context 绑定在一起。
// 上游 context 主要用于透传 request_id 等元数据，而不是取消信号。
type jobContext struct {
	job Job
	ctx context.Context
}

// valueOnlyContext 包装一个 context，只暴露它的值（Value），忽略它的取消信号。
// 在异步 Worker Pool 中使用，可以避免短生命周期的 HTTP 请求取消导致后台任务被中断。
type valueOnlyContext struct {
	context.Context
}

func (valueOnlyContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (valueOnlyContext) Done() <-chan struct{}       { return nil }
func (valueOnlyContext) Err() error                  { return nil }

// Pool 是一个有界 Worker Pool。
type Pool struct {
	workers     int
	queueSize   int
	highQueue   chan *jobContext // 高优先级队列
	normalQueue chan *jobContext // 普通优先级队列
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	once        sync.Once
	stopped     chan struct{}
}

// NewPool 创建 Worker Pool。
// workers 为并发 worker 数量；queueSize 为每个优先级队列的缓冲大小。
func NewPool(workers, queueSize int) *Pool {
	if workers <= 0 {
		workers = defaultWorkers
	}
	if queueSize < 0 {
		queueSize = defaultQueueSize
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		workers:     workers,
		queueSize:   queueSize,
		highQueue:   make(chan *jobContext, queueSize),
		normalQueue: make(chan *jobContext, queueSize),
		ctx:         ctx,
		cancel:      cancel,
		stopped:     make(chan struct{}),
	}
}

// Start 启动所有 worker goroutine。可安全多次调用，仅第一次生效。
func (p *Pool) Start() {
	p.once.Do(func() {
		for i := 0; i < p.workers; i++ {
			p.wg.Add(1)
			go p.worker(i)
		}
		slog.Info("worker pool started", "workers", p.workers, "queue_size", p.queueSize)
	})
}

// Submit 非阻塞提交普通优先级任务。
// 如果普通队列已满，返回错误。
func (p *Pool) Submit(job Job) error {
	return p.submit(context.Background(), job, false, false)
}

// SubmitBlocking 阻塞提交普通优先级任务，直到入队成功或 ctx 取消。
func (p *Pool) SubmitBlocking(ctx context.Context, job Job) error {
	return p.submit(ctx, job, false, true)
}

// SubmitWithPriority 非阻塞提交任务，high 为 true 时进入高优先级队列。
func (p *Pool) SubmitWithPriority(job Job, high bool) error {
	return p.submit(context.Background(), job, high, false)
}

// SubmitWithPriorityBlocking 阻塞提交任务，high 为 true 时进入高优先级队列。
func (p *Pool) SubmitWithPriorityBlocking(ctx context.Context, job Job, high bool) error {
	return p.submit(ctx, job, high, true)
}

func (p *Pool) submit(ctx context.Context, job Job, high, blocking bool) error {
	select {
	case <-p.ctx.Done():
		return errors.New("worker pool is stopping")
	default:
	}

	q := p.normalQueue
	if high {
		q = p.highQueue
	}

	jc := &jobContext{job: job, ctx: ctx}

	if blocking {
		select {
		case q <- jc:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	select {
	case q <- jc:
		return nil
	default:
		return fmt.Errorf("job queue full (size=%d)", p.queueSize)
	}
}

// Stop 优雅关闭 Worker Pool。
// 1. 取消 context，阻止新任务入队（Submit 返回错误）。
// 2. 关闭任务队列，通知 worker 退出。
// 3. 等待所有 worker 完成当前任务。
func (p *Pool) Stop() {
	p.cancel()
	close(p.highQueue)
	close(p.normalQueue)
	p.wg.Wait()
	close(p.stopped)
	slog.Info("worker pool stopped")
}

// worker 是单个 worker 的主循环。
// worker 只通过 channel 状态判断是否退出：当两个队列都关闭且为空时退出。
// 这种设计保证 Stop 时队列中已入队的任务都会被执行完。
func (p *Pool) worker(id int) {
	defer p.wg.Done()
	logger := slog.With("worker_id", id)

	for {
		// 优先处理高优先级任务（非阻塞尝试）
		select {
		case jc, ok := <-p.highQueue:
			if !ok {
				// 高优先级队列已关闭， drain 普通队列后退出
				p.drainNormal(logger)
				return
			}
			p.runJob(logger, jc)
			continue
		default:
		}

		// 没有高优先级任务时，阻塞等待普通任务或高优先级任务
		select {
		case jc, ok := <-p.highQueue:
			if !ok {
				p.drainNormal(logger)
				return
			}
			p.runJob(logger, jc)
		case jc, ok := <-p.normalQueue:
			if !ok {
				return
			}
			p.runJob(logger, jc)
		}
	}
}

// drainNormal 清空普通队列中的剩余任务。
func (p *Pool) drainNormal(logger *slog.Logger) {
	for {
		select {
		case jc, ok := <-p.normalQueue:
			if !ok {
				return
			}
			p.runJob(logger, jc)
		default:
			return
		}
	}
}

// runJob 执行单个任务并记录日志。
// 使用 valueOnlyContext 透传上游 context 的值（如 request_id），但屏蔽上游的取消信号，
// 避免短生命周期的 HTTP 请求中断异步的后台任务。
func (p *Pool) runJob(logger *slog.Logger, jc *jobContext) {
	logger = logger.With("job_id", jc.job.ID())
	logger.Info("processing job")
	start := time.Now()

	ctx, cancel := context.WithTimeout(valueOnlyContext{jc.ctx}, defaultJobTimeout)
	defer cancel()

	if err := jc.job.Execute(ctx); err != nil {
		logger.Error("job failed", "error", err, "cost", time.Since(start))
		return
	}
	logger.Info("job completed", "cost", time.Since(start))
}
