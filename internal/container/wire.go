//go:build wireinject
// +build wireinject

package container

import (
	"github.com/google/wire"

	"redis-like-golang/internal/adapter/handler"
	"redis-like-golang/internal/adapter/protocol"
	"redis-like-golang/internal/infrastructure/persistence"
	"redis-like-golang/internal/infrastructure/storage"
	"redis-like-golang/internal/usecase"
)

// InitializeContainer creates a new container with all dependencies using Wire
func InitializeContainer(opt persistence.AOFProviderOption) (*Container, func(), error) {
	wire.Build(
		// Infrastructure providers
		storage.NewStore,
		persistence.NewAOFProvider,

		// Adapter providers
		protocol.NewParser,

		// Use case providers
		usecase.NewStats,
		usecase.NewCommandHandler,

		// Handler providers
		handler.NewTCPHandler,

		// Container provider
		NewContainer,
	)
	return nil, nil, nil
}
