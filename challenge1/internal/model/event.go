package model

import (
	"encoding/binary"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

// IndexedEvent represents a stored L1InfoRoot event with block metadata.
// Each event is keyed by an incrementing index starting at 0.
type IndexedEvent struct {
	Index       uint64      // Sequential index (key in storage)
	BlockNumber uint64      // Block where the event was emitted
	BlockTime   uint64      // Timestamp of the block
	ParentHash  common.Hash // Parent hash of the block

	TxHash   common.Hash // Transaction that emitted the event
	LogIndex uint        // Log index within the block

	InfoRoot common.Hash // L1 info root data from the event
}

// Binary layout (fixed size for efficient storage):
// - Index:       8 bytes
// - BlockNumber: 8 bytes
// - BlockTime:   8 bytes
// - ParentHash:  32 bytes
// - TxHash:      32 bytes
// - LogIndex:    4 bytes (uint32)
// - InfoRoot:    32 bytes
// Total: 124 bytes

const binaryEventSize = 8 + 8 + 8 + 32 + 32 + 4 + 32 // 124 bytes

// MarshalBinary encodes the event to a compact binary format.
func (e *IndexedEvent) MarshalBinary() ([]byte, error) {
	buf := make([]byte, binaryEventSize)
	offset := 0

	binary.BigEndian.PutUint64(buf[offset:], e.Index)
	offset += 8

	binary.BigEndian.PutUint64(buf[offset:], e.BlockNumber)
	offset += 8

	binary.BigEndian.PutUint64(buf[offset:], e.BlockTime)
	offset += 8

	copy(buf[offset:], e.ParentHash[:])
	offset += 32

	copy(buf[offset:], e.TxHash[:])
	offset += 32

	binary.BigEndian.PutUint32(buf[offset:], uint32(e.LogIndex))
	offset += 4

	copy(buf[offset:], e.InfoRoot[:])

	return buf, nil
}

// UnmarshalBinary decodes an event from its binary representation.
func (e *IndexedEvent) UnmarshalBinary(data []byte) error {
	if len(data) != binaryEventSize {
		return errors.New("invalid binary event size")
	}

	offset := 0

	e.Index = binary.BigEndian.Uint64(data[offset:])
	offset += 8

	e.BlockNumber = binary.BigEndian.Uint64(data[offset:])
	offset += 8

	e.BlockTime = binary.BigEndian.Uint64(data[offset:])
	offset += 8

	copy(e.ParentHash[:], data[offset:offset+32])
	offset += 32

	copy(e.TxHash[:], data[offset:offset+32])
	offset += 32

	e.LogIndex = uint(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	copy(e.InfoRoot[:], data[offset:offset+32])

	return nil
}
