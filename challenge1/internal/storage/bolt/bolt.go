package bolt

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/zacksfF/sepolia-sh/ch1/internal/model"
	"go.etcd.io/bbolt"
)

var (
	metaBucket   = []byte("meta")
	eventsBucket = []byte("events")
	indexKey     = []byte("next_index")

	// ErrNotFound is returned when an event doesn't exist
	ErrNotFound = errors.New("event not found")
)

// Store implements storage.Store using BoltDB
type Store struct {
	db *bbolt.DB
}

// Open creates a new Store with the given BoltDB file path
func Open(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(metaBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(eventsBucket); err != nil {
			return err
		}
		return nil
	})

	return &Store{db: db}, err
}

// GetNextIndex returns the next available event index
func (s *Store) GetNextIndex(ctx context.Context) (uint64, error) {
	var idx uint64

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(metaBucket)
		v := b.Get(indexKey)
		if v == nil {
			idx = 0
			return nil
		}
		idx = binary.BigEndian.Uint64(v)
		return nil
	})

	return idx, err
}

// SetNextIndex updates the next available event index
func (s *Store) SetNextIndex(ctx context.Context, idx uint64) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(metaBucket)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, idx)
		return b.Put(indexKey, buf)
	})
}

// SaveEvent persists an indexed event using binary serialization
func (s *Store) SaveEvent(ctx context.Context, e *model.IndexedEvent) error {
	if e == nil {
		return errors.New("nil event")
	}

	data, err := e.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(eventsBucket)

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, e.Index)

		return b.Put(key, data)
	})
}

// GetEvent retrieves an event by its index
func (s *Store) GetEvent(ctx context.Context, index uint64) (*model.IndexedEvent, error) {
	var event model.IndexedEvent

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(eventsBucket)

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, index)

		data := b.Get(key)
		if data == nil {
			return ErrNotFound
		}

		return event.UnmarshalBinary(data)
	})

	if err != nil {
		return nil, err
	}
	return &event, nil
}

// Close releases the database resources
func (s *Store) Close() error {
	return s.db.Close()
}
