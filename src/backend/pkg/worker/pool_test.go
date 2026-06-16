package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testJob struct {
	id      string
	execute func(ctx context.Context) error
}

func (j *testJob) ID() string                        { return j.id }
func (j *testJob) Execute(ctx context.Context) error { return j.execute(ctx) }

func TestPool_Submit(t *testing.T) {
	pool := NewPool(2, 10)
	pool.Start()
	defer pool.Stop()

	var count int32
	job := &testJob{
		id: "test-1",
		execute: func(_ context.Context) error {
			atomic.AddInt32(&count, 1)
			return nil
		},
	}

	if err := pool.Submit(job); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for atomic.LoadInt32(&count) != 1 {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("job should be executed once, got %d", atomic.LoadInt32(&count))
	}
}

func TestPool_QueueFull(t *testing.T) {
	pool := NewPool(1, 0) // 无缓冲队列
	pool.Start()
	defer pool.Stop()

	// 提交一个长时间运行的任务占满 worker
	if err := pool.Submit(&testJob{
		id: "blocking",
		execute: func(_ context.Context) error {
			time.Sleep(5 * time.Second)
			return nil
		},
	}); err != nil {
		t.Fatalf("first submit should succeed: %v", err)
	}

	// 第二个任务应该因为队列满而失败
	err := pool.Submit(&testJob{id: "second", execute: func(ctx context.Context) error { return nil }})
	if err == nil {
		t.Error("should return error when queue full")
	}
}

func TestPool_GracefulShutdown(t *testing.T) {
	pool := NewPool(2, 10)
	pool.Start()

	var completed int32
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("job-%d", i)
		if err := pool.Submit(&testJob{
			id: id,
			execute: func(_ context.Context) error {
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt32(&completed, 1)
				return nil
			},
		}); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	pool.Stop()
	if atomic.LoadInt32(&completed) != 5 {
		t.Errorf("all jobs should complete before shutdown, got %d", completed)
	}
}

func TestPool_Priority(t *testing.T) {
	pool := NewPool(1, 10) // 单 worker，方便观察优先级
	pool.Start()
	defer pool.Stop()

	var order []string
	var mu sync.Mutex
	started := make(chan struct{})

	// 先提交一个长时间运行的普通任务，占住 worker
	if err := pool.Submit(&testJob{
		id: "normal-1",
		execute: func(_ context.Context) error {
			close(started)
			time.Sleep(100 * time.Millisecond)
			mu.Lock()
			order = append(order, "normal-1")
			mu.Unlock()
			return nil
		},
	}); err != nil {
		t.Fatalf("submit normal failed: %v", err)
	}

	// 等待普通任务开始执行，确保 worker 已被占用
	<-started

	// 在普通任务执行期间提交高优先级任务
	if err := pool.SubmitWithPriority(&testJob{
		id: "high-1",
		execute: func(_ context.Context) error {
			mu.Lock()
			order = append(order, "high-1")
			mu.Unlock()
			return nil
		},
	}, true); err != nil {
		t.Fatalf("submit high failed: %v", err)
	}

	// 等待两个任务都完成
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(order))
	}
	// 普通任务先占用 worker，完成后 worker 空闲时优先选择高优先级任务
	if order[0] != "normal-1" || order[1] != "high-1" {
		t.Errorf("expected [normal-1, high-1], got %v", order)
	}
}

func TestPool_StopPreventsNewJobs(t *testing.T) {
	pool := NewPool(1, 10)
	pool.Start()
	pool.Stop()

	if err := pool.Submit(&testJob{id: "after-stop", execute: func(ctx context.Context) error { return nil }}); err == nil {
		t.Error("should reject new jobs after stop")
	}
}

func TestPool_JobTimeout(t *testing.T) {
	pool := NewPool(1, 1)
	pool.Start()
	defer pool.Stop()

	var errReturned error
	job := &testJob{
		id: "timeout",
		execute: func(ctx context.Context) error {
			select {
			case <-time.After(10 * time.Second):
				return nil
			case <-ctx.Done():
				errReturned = ctx.Err()
				return ctx.Err()
			}
		},
	}

	// 临时缩短超时时间用于测试
	oldTimeout := defaultJobTimeout
	defaultJobTimeout = 100 * time.Millisecond
	defer func() { defaultJobTimeout = oldTimeout }()

	if err := pool.Submit(job); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	if errReturned != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded, got %v", errReturned)
	}
}

func TestPool_UpstreamValuePropagated(t *testing.T) {
	pool := NewPool(1, 1)
	pool.Start()
	defer pool.Stop()

	type ctxKey struct{}
	const expectedValue = "upstream-request-id"

	done := make(chan string, 1)
	job := &testJob{
		id: "value",
		execute: func(ctx context.Context) error {
			if v, ok := ctx.Value(ctxKey{}).(string); ok {
				done <- v
			} else {
				done <- ""
			}
			return nil
		},
	}

	ctx := context.WithValue(context.Background(), ctxKey{}, expectedValue)
	if err := pool.SubmitBlocking(ctx, job); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	select {
	case got := <-done:
		if got != expectedValue {
			t.Errorf("upstream value = %q, want %q", got, expectedValue)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("job did not execute in time")
	}
}

func TestPool_UpstreamCancellationIgnored(t *testing.T) {
	pool := NewPool(1, 1)
	pool.Start()
	defer pool.Stop()

	var executed bool
	job := &testJob{
		id: "ignore-cancel",
		execute: func(ctx context.Context) error {
			// 即使上游 context 已被取消，异步任务仍应继续执行
			select {
			case <-time.After(100 * time.Millisecond):
				executed = true
				return nil
			case <-ctx.Done():
				t.Errorf("job should not be cancelled by upstream context")
				return ctx.Err()
			}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消上游 context

	if err := pool.SubmitBlocking(ctx, job); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	// 让队列中的任务有时间执行
	time.Sleep(200 * time.Millisecond)
	if !executed {
		t.Error("job should have been executed despite upstream cancellation")
	}
}
