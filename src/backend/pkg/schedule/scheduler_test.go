package schedule

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *miniredis.Miniredis {
	s := miniredis.RunT(t)
	t.Cleanup(s.Close)
	return s
}

func TestScheduler_RunOnce_AcquireAndRun(t *testing.T) {
	s := newTestRedis(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	scheduler := NewScheduler(client, "test")

	var count int32
	err := scheduler.RunOnce(context.Background(), "job-1", 10*time.Second, func(_ context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("task should run once, got %d", count)
	}
}

func TestScheduler_RunOnce_LockNotAcquired(t *testing.T) {
	s := newTestRedis(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	scheduler := NewScheduler(client, "test")

	// 先占住锁
	if err := client.Set(context.Background(), "test:lock:job-2", "other", 10*time.Second).Err(); err != nil {
		t.Fatalf("set lock failed: %v", err)
	}

	var count int32
	err := scheduler.RunOnce(context.Background(), "job-2", 10*time.Second, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	if !errors.Is(err, ErrLockNotAcquired) {
		t.Errorf("expected ErrLockNotAcquired, got %v", err)
	}
	if atomic.LoadInt32(&count) != 0 {
		t.Errorf("task should not run, got %d", count)
	}
}

func TestScheduler_RunOnce_TaskFailureKeepsLock(t *testing.T) {
	s := newTestRedis(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	scheduler := NewScheduler(client, "test")

	expectedErr := errors.New("task failed")
	err := scheduler.RunOnce(context.Background(), "job-3", 10*time.Second, func(ctx context.Context) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected task error, got %v", err)
	}

	// 失败时锁应仍然存在
	val, err := client.Get(context.Background(), "test:lock:job-3").Result()
	if err != nil || val == "" {
		t.Error("lock should be kept after task failure")
	}
}

func TestScheduler_LastExecutedAt(t *testing.T) {
	s := newTestRedis(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	scheduler := NewScheduler(client, "test")

	ctx := context.Background()
	at, err := scheduler.LastExecutedAt(ctx, "job-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if at != "" {
		t.Errorf("expected empty, got %s", at)
	}

	_ = scheduler.RunOnce(ctx, "job-4", 10*time.Second, func(_ context.Context) error {
		return nil
	})

	at, err = scheduler.LastExecutedAt(ctx, "job-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if at == "" {
		t.Error("expected non-empty execution time")
	}
}
