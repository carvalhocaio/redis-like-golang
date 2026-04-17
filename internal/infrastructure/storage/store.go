package storage

import (
	"context"
	"path/filepath"
	"redis-like-golang/internal/domain/entity"
	"redis-like-golang/internal/domain/repository"
	"sync"
	"time"
)

type Store struct {
	data        map[string]*entity.Item
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

func NewStore() repository.KeyValueRepository {
	return &Store{
		data:        make(map[string]*entity.Item),
		stopCleanup: make(chan struct{}),
	}
}

func (s *Store) Set(ctx context.Context, key, value string) {
	if ctx.Err() != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = &entity.Item{
		Value: value, ExpiresAt: nil,
	}
}

func (s *Store) Get(ctx context.Context, key string) (string, bool) {
	if ctx.Err() != nil {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return "", false
	}

	if item.IsExpired(time.Now().Unix()) {
		return "", false
	}

	return item.Value, true
}

func (s *Store) Del(ctx context.Context, key string) int {
	if ctx.Err() != nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return 1
	}

	return 0
}

func (s *Store) Expire(ctx context.Context, key string, seconds int) bool {
	if ctx.Err() != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	expiresAt := time.Now().Unix() + int64(seconds)
	item.ExpiresAt = &expiresAt
	return true
}

func (s *Store) TTL(ctx context.Context, key string) int64 {
	if ctx.Err() != nil {
		return -1
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return -1
	}

	if item.ExpiresAt == nil {
		return -1
	}

	now := time.Now().Unix()
	remaining := *item.ExpiresAt - now

	if remaining <= 0 {
		return -1
	}

	return remaining
}

func (s *Store) Persist(ctx context.Context, key string) bool {
	if ctx.Err() != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	item.ExpiresAt = nil

	return true
}

func (s *Store) Keys(ctx context.Context, pattern string) []string {
	if ctx.Err() != nil {
		return []string{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().Unix()
	var matches []string

	for key, item := range s.data {
		if item.IsExpired(now) {
			continue
		}

		if matchPattern(key, pattern) {
			matches = append(matches, key)
		}
	}

	return matches
}

func (s *Store) Exists(ctx context.Context, key string) bool {
	if ctx.Err() != nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	return !item.IsExpired(time.Now().Unix())
}

func (s *Store) Size(ctx context.Context) int {
	if ctx.Err() != nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}

func (s *Store) StartCleanup(intervalMs int64) {
	interval := time.Duration(intervalMs) * time.Millisecond
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanupExpired()
			case <-s.stopCleanup:
				return
			}
		}
	}()
}

func (s *Store) StopCleanup() {
	close(s.stopCleanup)
}

func (s *Store) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Lock()

	now := time.Now().Unix()
	for key, item := range s.data {
		if item.IsExpired(now) {
			delete(s.data, key)
		}
	}
}

func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	matched, err := filepath.Match(pattern, key)
	if err != nil {
		return key == pattern
	}

	return matched
}
