package storage

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"redis-like-golang/internal/domain/entity"
	"redis-like-golang/internal/domain/repository"
)

// Store implements KeyValueRepository - thread-safe key-value store with TTL support
type Store struct {
	data        map[string]*entity.Item
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

// NewStore creates a new Store instance
func NewStore() repository.KeyValueRepository {
	return &Store{
		data:        make(map[string]*entity.Item),
		stopCleanup: make(chan struct{}),
	}
}

// Set stores or updates a key-value pair
func (s *Store) Set(ctx context.Context, key, value string) {
	if ctx.Err() != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &entity.Item{Value: value, ExpiresAt: nil}
}

// Get retrieves a value by key. Returns the value and true if found and not expired, false otherwise
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

	// Check if expired
	if item.IsExpired(time.Now().Unix()) {
		return "", false
	}

	return item.Value, true
}

// Del removes a key and returns the number of keys removed (0 or 1)
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

// Expire sets the TTL for a key in seconds. Returns true if the key exists, false otherwise
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

// TTL returns the remaining time-to-live in seconds for a key.
// Returns -1 if the key does not exist or has no expiration, or the remaining seconds
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
		return -1 // Already expired (will be cleaned up)
	}

	return remaining
}

// Persist removes the expiration from a key. Returns true if the key exists, false otherwise
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

// Keys returns all keys matching the pattern (supports * and ? wildcards)
func (s *Store) Keys(ctx context.Context, pattern string) []string {
	if ctx.Err() != nil {
		return []string{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().Unix()
	var matches []string

	for key, item := range s.data {
		// Skip expired keys
		if item.IsExpired(now) {
			continue
		}

		// Match pattern (simple wildcard: * and ?)
		if matchPattern(key, pattern) {
			matches = append(matches, key)
		}
	}

	return matches
}

// Exists There are checks if a key exists and is not expired
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

	// Check if expired
	return !item.IsExpired(time.Now().Unix())
}

// Size returns the total number of keys (including expired ones, for counting)
func (s *Store) Size(ctx context.Context) int {
	if ctx.Err() != nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// StartCleanup starts a background goroutine that periodically removes expired keys
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

// StopCleanup stops the cleanup goroutine
func (s *Store) StopCleanup() {
	close(s.stopCleanup)
}

// cleanupExpired removes all expired keys
func (s *Store) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	for key, item := range s.data {
		if item.IsExpired(now) {
			delete(s.data, key)
		}
	}
}

// matchPattern matches a key against a pattern with wildcards
// Supports * (matches any sequence) and? (matches single character)
func matchPattern(key, pattern string) bool {
	// Convert Redis-style pattern to Go filepath.Match pattern
	// Redis uses * for any sequence and ? for single char
	// Go filepath.Match uses the same, but we need to handle edge cases

	// If pattern is "*", match everything
	if pattern == "*" {
		return true
	}

	// Use filepath.Match which supports * and?
	matched, err := filepath.Match(pattern, key)
	if err != nil {
		// If pattern is invalid, do exact match
		return key == pattern
	}

	return matched
}
