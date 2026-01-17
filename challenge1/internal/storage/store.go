package storage

import (
	"context"

	"github.com/zacksfF/sepolia-sh/ch1/internal/model"
)

// Store defines the interface for persisting and retrieving indexed events.
type Store interface {
	// GetNextIndex returns the next available index for storing events.
	GetNextIndex(ctx context.Context) (uint64, error)

	// SetNextIndex updates the next available index.
	SetNextIndex(ctx context.Context, idx uint64) error

	// SaveEvent persists an indexed event to storage.
	SaveEvent(ctx context.Context, event *model.IndexedEvent) error

	// GetEvent retrieves an event by its index.
	GetEvent(ctx context.Context, index uint64) (*model.IndexedEvent, error)

	// Close releases storage resources.
	Close() error
}
