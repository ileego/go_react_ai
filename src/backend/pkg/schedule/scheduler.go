// Package schedule 提供轻量分布式定时任务能力。
// 基于 Redis SET NX EX 实现单实例锁，保证多副本部署时同一任务只在一个实例执行。
package schedule

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Scheduler 是分布式定时任务调度器。
type Scheduler struct {
	client redis.Cmdable
	prefix string
}

// NewScheduler 创建 Scheduler。
// prefix 用于隔离不同服务的锁 key。
func NewScheduler(client redis.Cmdable, prefix string) *Scheduler {
	if prefix == "" {
		prefix = "scheduler"
	}
	return &Scheduler{
		client: client,
		prefix: prefix,
	}
}

// Task 是一个可执行的定时任务。
type Task func(ctx context.Context) error

// RunOnce 尝试获取分布式锁并执行任务。
// key: 锁的唯一标识；ttl: 锁过期时间；task: 要执行的任务。
// 如果锁已被其他实例持有，返回 ErrLockNotAcquired，不会执行任务。
func (s *Scheduler) RunOnce(ctx context.Context, key string, ttl time.Duration, task Task) error {
	lockKey := fmt.Sprintf("%s:lock:%s", s.prefix, key)
	execKey := fmt.Sprintf("%s:exec:%s", s.prefix, key)
	token := generateToken()

	acquired, err := s.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	if !acquired {
		return ErrLockNotAcquired
	}

	// 获取锁成功后，记录本次执行时间用于幂等性判断
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.client.Set(ctx, execKey, now, ttl).Err(); err != nil {
		slog.Warn("failed to record execution time", "key", key, "error", err)
	}

	slog.Info("scheduler acquired lock, running task", "key", key)
	if err := task(ctx); err != nil {
		// 任务失败不立即释放锁，让 ttl 自然过期，避免失败时其他实例立刻重试
		slog.Error("scheduled task failed", "key", key, "error", err)
		return fmt.Errorf("task execution: %w", err)
	}

	// 任务成功，主动释放锁（使用 token 保证不会误释放其他实例的锁）
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	if _, err := s.client.Eval(ctx, script, []string{lockKey}, token).Result(); err != nil {
		slog.Warn("failed to release lock", "key", key, "error", err)
	}

	slog.Info("scheduled task completed", "key", key)
	return nil
}

// LastExecutedAt 返回任务最近一次成功执行的时间（UTC RFC3339 字符串）。
// 若未执行过，返回空字符串。
func (s *Scheduler) LastExecutedAt(ctx context.Context, key string) (string, error) {
	execKey := fmt.Sprintf("%s:exec:%s", s.prefix, key)
	val, err := s.client.Get(ctx, execKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// ErrLockNotAcquired 表示锁已被其他实例持有。
var ErrLockNotAcquired = fmt.Errorf("lock not acquired")

// generateToken 生成 16 字节随机 token。
func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// 随机失败时使用时间戳作为兜底
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
