# Sepolia L1InfoRoot Indexer

## Overview

This is an event indexer that pulls L1InfoRoot events from a contract on Sepolia and stores them locally in BoltDB. I built this for an L2 project that needs efficient access to historical event data.

The indexer queries for a specific event topic, grabs the block metadata (timestamp and parent hash), and stores everything in a key-value store with sequential indices starting at 0. Multiple events in the same block are handled correctly - each gets its own incrementing index.

## Quick Start

```bash
# Build
cd challenge1
go build ./cmd/indexer

# Run (auto-loads .env)
./indexer
```

The indexer scans from block 0 to latest, stores all matching events, then exits cleanly.

## Configuration

All config comes from environment variables. The included `.env` file is auto-loaded via godotenv:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CONTRACT_ADDRESS` | Yes | - | Contract to query events from |
| `EVENT_TOPIC` | Yes | - | Event signature hash to filter |
| `RPC_URL` | No | publicnode.com | Ethereum RPC endpoint |
| `START_BLOCK` | No | 0 | Block to start indexing from |
| `END_BLOCK` | No | latest | Block to stop at (omit for latest) |
| `BATCH_SIZE` | No | 5000 | Blocks per eth_getLogs call |
| `DB_PATH` | No | ./sepolia.db | BoltDB file path |

## Data Model

Each indexed event contains:

| Field | Type | Size | Description |
|-------|------|------|-------------|
| Index | uint64 | 8B | Sequential key (0, 1, 2...) |
| BlockNumber | uint64 | 8B | Block containing the event |
| BlockTime | uint64 | 8B | Block timestamp |
| ParentHash | Hash | 32B | Parent block hash |
| TxHash | Hash | 32B | Transaction hash |
| LogIndex | uint32 | 4B | Position within block |
| InfoRoot | Hash | 32B | L1 info root from event data |

**Total: 124 bytes per event** (binary encoded, not JSON)

## Project Structure

```
cmd/indexer/main.go       Entry point, wires dependencies
config/config.go          Environment loading + defaults
internal/
  eth/
    client.go             RPC connection wrapper
    logs.go               eth_getLogs with FilterQuery
    blocks.go             Block metadata fetching
  indexer/indexer.go      Main sync loop
  model/event.go          Event struct + binary marshal/unmarshal
  storage/
    store.go              Storage interface
    bolt/bolt.go          BoltDB implementation
test/
  integration_test.go     End-to-end test against Sepolia
```

## Design Decisions

**Batch querying**: Logs are fetched in batches (default: 5000 blocks). Larger batches = fewer RPC calls = faster sync. Tunable via `BATCH_SIZE` for rate-limited endpoints.

**Binary serialization**: Events are stored as 124-byte fixed-size binary instead of JSON (~300+ bytes). Benefits: smaller DB, faster encode/decode, predictable sizing. Layout is documented in `model/event.go`.

**Block caching**: Block metadata is cached by hash during each batch. If a block contains multiple events, we fetch it once. Cache resets between batches to bound memory.

**Sequential indexing**: The challenge requires events keyed by incrementing index. Logs from `eth_getLogs` come sorted by (blockNumber, logIndex), so we simply increment a counter. The counter persists in DB across restarts.

**Graceful shutdown**: Handles SIGINT/SIGTERM. Context cancellation propagates through the call stack.

## Tradeoffs & Assumptions

1. **InfoRoot parsing**: I assume `event.Data` contains the raw L1 info root hash (32 bytes). If the event ABI encodes it differently, parsing would need adjustment.

2. **No reorg handling**: Chain reorgs aren't detected. For production L2 use, you'd either wait for finality or implement reorg detection by validating parent hashes.

3. **Restart behavior**: On restart, the indexer re-scans from `START_BLOCK` but skips storing events (since next_index is persisted). A production version would persist `lastProcessedBlock` to avoid re-scanning.

4. **Single-writer**: BoltDB allows one writer. Fine for a CLI tool; concurrent indexing would need a different store.

5. **Timeout**: 10-minute context timeout in main.go. Adjust for very large historical syncs.

## Future Improvements

With more time, I'd add:

- **Last block tracking** - persist last processed block, resume without rescanning
- **Progress indicator** - percentage complete, blocks/sec, ETA
- **Retry logic** - exponential backoff for transient RPC failures
- **Parallel block fetching** - fetch multiple block headers concurrently
- **Prometheus metrics** - for production monitoring
- **CLI flags** - alongside env vars for flexibility

## Running Tests

```bash
# Unit tests
go test ./internal/...

# Integration test (requires RPC)
go test -v ./test/...
```

## Dependencies

- `github.com/ethereum/go-ethereum` - Ethereum client
- `go.etcd.io/bbolt` - Embedded key-value store
- `github.com/joho/godotenv` - .env file loading
