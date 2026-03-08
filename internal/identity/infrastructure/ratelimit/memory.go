package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count    int
	windowAt time.Time
}

type InMemoryLimiter struct {
	mu       sync.Mutex
	entries  map[string]*entry
	limit    int
	window   time.Duration
}

func NewInMemoryLimiter(limit int, window time.Duration) *InMemoryLimiter {
	return &InMemoryLimiter{
		entries: make(map[string]*entry),
		limit:   limit,
		window:  window,
	}
}

func (l *InMemoryLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, exists := l.entries[key]
	if !exists || now.After(e.windowAt.Add(l.window)) {
		l.entries[key] = &entry{count: 1, windowAt: now}
		return true
	}
	e.count++
	return e.count <= l.limit
}

func (l *InMemoryLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}
