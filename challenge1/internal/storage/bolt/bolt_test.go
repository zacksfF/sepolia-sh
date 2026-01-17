package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/sepolia-sh/ch1/internal/model"
)

func TestStore_SaveAndGetEvent(t *testing.T) {
	dbPath := "test_bolt.db"
	defer os.Remove(dbPath)

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test event
	event := &model.IndexedEvent{
		Index:       0,
		BlockNumber: 12345,
		BlockTime:   1700000000,
		ParentHash:  common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		TxHash:      common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		LogIndex:    2,
		InfoRoot:    common.HexToHash("0xdeadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678"),
	}

	// Save the event
	if err := store.SaveEvent(ctx, event); err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	// Retrieve the event
	retrieved, err := store.GetEvent(ctx, 0)
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	// Verify fields
	if retrieved.Index != event.Index {
		t.Errorf("Index mismatch: got %d, want %d", retrieved.Index, event.Index)
	}
	if retrieved.BlockNumber != event.BlockNumber {
		t.Errorf("BlockNumber mismatch: got %d, want %d", retrieved.BlockNumber, event.BlockNumber)
	}
	if retrieved.BlockTime != event.BlockTime {
		t.Errorf("BlockTime mismatch: got %d, want %d", retrieved.BlockTime, event.BlockTime)
	}
	if retrieved.ParentHash != event.ParentHash {
		t.Errorf("ParentHash mismatch: got %s, want %s", retrieved.ParentHash, event.ParentHash)
	}
	if retrieved.TxHash != event.TxHash {
		t.Errorf("TxHash mismatch: got %s, want %s", retrieved.TxHash, event.TxHash)
	}
	if retrieved.LogIndex != event.LogIndex {
		t.Errorf("LogIndex mismatch: got %d, want %d", retrieved.LogIndex, event.LogIndex)
	}
	if retrieved.InfoRoot != event.InfoRoot {
		t.Errorf("InfoRoot mismatch: got %s, want %s", retrieved.InfoRoot, event.InfoRoot)
	}
}

func TestStore_GetEventNotFound(t *testing.T) {
	dbPath := "test_bolt_notfound.db"
	defer os.Remove(dbPath)

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Try to get a non-existent event
	_, err = store.GetEvent(ctx, 999)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestStore_IndexManagement(t *testing.T) {
	dbPath := "test_bolt_index.db"
	defer os.Remove(dbPath)

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Initial index should be 0
	idx, err := store.GetNextIndex(ctx)
	if err != nil {
		t.Fatalf("Failed to get next index: %v", err)
	}
	if idx != 0 {
		t.Errorf("Initial index should be 0, got %d", idx)
	}

	// Set index to 5
	if err := store.SetNextIndex(ctx, 5); err != nil {
		t.Fatalf("Failed to set next index: %v", err)
	}

	// Verify index is now 5
	idx, err = store.GetNextIndex(ctx)
	if err != nil {
		t.Fatalf("Failed to get next index: %v", err)
	}
	if idx != 5 {
		t.Errorf("Expected index 5, got %d", idx)
	}
}

func TestStore_MultipleEventsPerBlock(t *testing.T) {
	dbPath := "test_bolt_multi.db"
	defer os.Remove(dbPath)

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	blockHash := common.HexToHash("0xaabbccdd1234567890abcdef1234567890abcdef1234567890abcdef12345678")

	// Simulate multiple events from the same block (different log indices)
	events := []*model.IndexedEvent{
		{
			Index:       0,
			BlockNumber: 100,
			BlockTime:   1700000000,
			ParentHash:  blockHash,
			TxHash:      common.HexToHash("0x1111"),
			LogIndex:    0,
			InfoRoot:    common.HexToHash("0x2222"),
		},
		{
			Index:       1,
			BlockNumber: 100, // Same block
			BlockTime:   1700000000,
			ParentHash:  blockHash,
			TxHash:      common.HexToHash("0x1111"),
			LogIndex:    1, // Different log index
			InfoRoot:    common.HexToHash("0x3333"),
		},
		{
			Index:       2,
			BlockNumber: 100, // Same block
			BlockTime:   1700000000,
			ParentHash:  blockHash,
			TxHash:      common.HexToHash("0x1111"),
			LogIndex:    2, // Different log index
			InfoRoot:    common.HexToHash("0x4444"),
		},
	}

	// Save all events
	for _, e := range events {
		if err := store.SaveEvent(ctx, e); err != nil {
			t.Fatalf("Failed to save event %d: %v", e.Index, err)
		}
	}

	// Verify we can retrieve each event independently
	for _, want := range events {
		got, err := store.GetEvent(ctx, want.Index)
		if err != nil {
			t.Fatalf("Failed to get event %d: %v", want.Index, err)
		}
		if got.Index != want.Index {
			t.Errorf("Index mismatch: got %d, want %d", got.Index, want.Index)
		}
		if got.LogIndex != want.LogIndex {
			t.Errorf("LogIndex mismatch for event %d: got %d, want %d", want.Index, got.LogIndex, want.LogIndex)
		}
	}
}
