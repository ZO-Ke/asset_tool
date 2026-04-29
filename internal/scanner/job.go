package scanner

import (
	"context"
	"sync"
	"sync/atomic"
)

// Job 通用扫描任务（httpx / rustscan / subdomain 共用）
type Job struct {
	ID     string
	Cancel context.CancelFunc

	mu      sync.Mutex
	paused  atomic.Bool
	pauseCh chan struct{}
}

func NewJob(id string, cancel context.CancelFunc) *Job {
	return &Job{
		ID:      id,
		Cancel:  cancel,
		pauseCh: make(chan struct{}),
	}
}

func (j *Job) Pause() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.paused.Load() {
		return
	}
	j.paused.Store(true)
	j.pauseCh = make(chan struct{})
}

func (j *Job) Resume() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if !j.paused.Load() {
		return
	}
	j.paused.Store(false)
	close(j.pauseCh)
}

func (j *Job) IsPaused() bool { return j.paused.Load() }

// WaitIfPaused 暂停时阻塞，直到 Resume 或 ctx 取消
func (j *Job) WaitIfPaused(ctx context.Context) {
	if !j.paused.Load() {
		return
	}
	j.mu.Lock()
	ch := j.pauseCh
	j.mu.Unlock()
	select {
	case <-ch:
	case <-ctx.Done():
	}
}

// ── 全局 job 注册表 ───────────────────────────────────────────────────

var (
	jobsMu sync.Mutex
	jobs   = map[string]*Job{}
)

func RegisterJob(j *Job) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	jobs[j.ID] = j
}

func UnregisterJob(id string) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	delete(jobs, id)
}

func GetJob(id string) *Job {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	return jobs[id]
}
