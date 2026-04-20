package repository

import "context"

type KeyValueRepository interface {
	Set(ctx context.Context, key, value string)
	Get(ctx context.Context, key string) (string, bool)
	Del(ctx context.Context, key string) int
	Expire(ctx context.Context, key string, seconds int) bool
	TTL(ctx context.Context, key string) int64
	Persist(ctx context.Context, key string) bool
	Keys(ctx context.Context, pattern string) []string
	Exists(ctx context.Context, key string) bool
	Size(ctx context.Context) int
	StartCleanup(intervalMs int64)
	StopCleanup()
}

// PersistenceRepository defines the interface for persistence operations
type PersistenceRepository interface {
	Append(ctx context.Context, command string, args []string) error
	Replay(ctx context.Context, store KeyValueRepository) error
	Close() error
}
