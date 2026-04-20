package persistence

import (
	"redis-like-golang/internal/domain/repository"
)

// AOFProviderOption wraps enableAOF and filepath for wire
type AOFProviderOption struct {
	EnableAOF bool
	FilePath  string
}

// NewAOFProvider creates an AOF repository if enabled, otherwise returns nil
func NewAOFProvider(opt AOFProviderOption) (repository.PersistenceRepository, error) {
	if !opt.EnableAOF {
		return nil, nil
	}
	return NewAOF(opt.FilePath)
}
