package usecase

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"redis-like-golang/internal/domain/repository"
)

// Stats holds server statistics
type Stats struct {
	startTime        time.Time
	totalCommands    int64
	totalConnections int64
	keyspace         repository.KeyValueRepository
}

// NewStats creates a new stats instance
func NewStats(keyspace repository.KeyValueRepository) *Stats {
	return &Stats{
		startTime: time.Now(),
		keyspace:  keyspace,
	}
}

// IncrementCommands increments the command counter
func (s *Stats) IncrementCommands() {
	atomic.AddInt64(&s.totalCommands, 1)
}

// IncrementConnections increments the connection counter
func (s *Stats) IncrementConnections() {
	atomic.AddInt64(&s.totalConnections, 1)
}

// GetInfo returns formatted INFO string
func (s *Stats) GetInfo(ctx context.Context) string {
	uptime := time.Since(s.startTime).Seconds()
	commands := atomic.LoadInt64(&s.totalCommands)
	connections := atomic.LoadInt64(&s.totalConnections)
	dbSize := s.keyspace.Size(ctx)

	info := fmt.Sprintf(`# Server
redis_version:redis-like-go/1.0.0
os:Go
uptime_in_seconds:%.0f
uptime_in_days:%.0f

# Clients
connected_clients:%d
total_connections_received:%d

# Stats
total_commands_processed:%d
keyspace_hits:0
keyspace_misses:0

# Keyspace
db0:keys=%d
`, uptime, uptime/86400, connections, connections, commands, dbSize)

	return info
}
