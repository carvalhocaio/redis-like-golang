package container

import (
	"redis-like-golang/internal/adapter/handler"
	"redis-like-golang/internal/adapter/protocol"
	"redis-like-golang/internal/domain/repository"
	"redis-like-golang/internal/usecase"
)

// Container holds all dependencies
type Container struct {
	Store          repository.KeyValueRepository
	Persistence    repository.PersistenceRepository
	CommandHandler *usecase.CommandHandler
	TCPHandler     *handler.TCPHandler
	Parser         *protocol.Parser
}

// NewContainer creates a new dependency injection container
func NewContainer(
	store repository.KeyValueRepository,
	persist repository.PersistenceRepository,
	parser *protocol.Parser,
	commandHandler *usecase.CommandHandler,
	tcpHandler *handler.TCPHandler,
) *Container {
	return &Container{
		Store:          store,
		Persistence:    persist,
		CommandHandler: commandHandler,
		TCPHandler:     tcpHandler,
		Parser:         parser,
	}
}

// Close closes all resources that need cleanup
func (c *Container) Close() error {
	if c.Persistence != nil {
		return c.Persistence.Close()
	}
	return nil
}
