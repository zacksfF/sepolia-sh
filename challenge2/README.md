# Multi-RPC Load Distribution

## Overview

This is a pull-based scheduler that distributes work across multiple RPC endpoints. The key insight is simple: fast RPCs complete more work naturally, without any coordination or synchronization.

Instead of assigning work to RPCs (push model), workers pull tasks from a shared queue when they're ready. A fast RPC finishes quickly and grabs the next task. A slow RPC is still working on its previous task. No waiting, no blocking.

## Quick Start

```bash
cd challenge2
go build ./cmd/demo
./demo
```

output like:
```
[publicnode] completed task 0 (blocks 8950000-8950999): 0 logs
[ankr] completed task 1 (blocks 8951000-8951999): 0 logs
[drpc] completed task 2 (blocks 8952000-8952999): 0 logs
[publicnode] completed task 3 (blocks 8953000-8953999): 0 logs
[publicnode] completed task 4 (blocks 8954000-8954999): 0 logs  <- fast RPC gets more
...
=== RPC Statistics ===
[publicnode] requests=25 failures=0 avg_latency=89ms
[ankr] requests=15 failures=0 avg_latency=156ms
[drpc] requests=10 failures=2 avg_latency=312ms
```

Notice how faster RPCs naturally complete more tasks.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Task Generator                         │
│         (produces block range tasks: 0-999, 1000-1999...)  │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Task Queue (channel)                     │
│              buffered channel of Task structs               │
└──────┬──────────────────┬──────────────────┬────────────────┘
       │                  │                  │
       ▼                  ▼                  ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│   Worker 1   │   │   Worker 2   │   │   Worker 3   │
│  pulls tasks │   │  pulls tasks │   │  pulls tasks │
│   when ready │   │   when ready │   │   when ready │
└──────────────┘   └──────────────┘   └──────────────┘
```

## How It Works

1. **Task Generator** creates work items (block ranges) and pushes them to a buffered channel
2. **Workers** (one per RPC) run their own goroutine, pulling tasks when idle
3. **Go's channel semantics** handle fair distribution - whoever is ready gets the next task
4. **Result Collector** aggregates results and tracks completion

This is Go's CSP (Communicating Sequential Processes) model in action.

## Key Design Decisions

**Pull vs Push**: Push-based scheduling requires knowing which RPC is fastest. Pull-based scheduling lets workers self-select - fast workers pull more.

**No global synchronization**: Workers are completely independent. A slow RPC doesn't block others. No mutexes, no condition variables, no coordination overhead.

**Exponential backoff**: When an RPC fails, the worker backs off (1s → 2s → 4s → max 30s) before retrying. This prevents hammering failed endpoints.

**Per-RPC statistics**: Each client tracks request count, failures, and latency. Useful for monitoring and debugging.

**Graceful shutdown**: Context cancellation propagates to all workers. In-flight tasks complete before exit.

## Project Structure

```
pkg/
  scheduler/
    task.go           Task and Result types
    worker.go         Pull-based worker with backoff
    scheduler.go      Main orchestrator
    scheduler_test.go Unit tests
  rpc/
    client.go         RPC wrapper with latency tracking
  config/
    config.go         Configuration and default endpoints
cmd/demo/
    main.go           Demo application
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONTRACT_ADDRESS` | (from challenge1) | Contract to query |
| `EVENT_TOPIC` | (from challenge1) | Event topic to filter |

Default RPC endpoints:
- https://ethereum-sepolia-rpc.publicnode.com
- https://rpc.ankr.com/eth_sepolia  
- https://sepolia.drpc.org

## Tradeoffs

1. **Failed tasks aren't retried**: In this implementation, if a task fails, it's logged but not re-queued. For production, you'd want a retry queue.

2. **No adaptive scoring**: We don't dynamically weight RPCs by performance. The pull model handles this implicitly, but explicit scoring could further optimize.

3. **Fixed batch size**: All tasks are the same size. Variable batching based on RPC capabilities could improve efficiency.

## Future Improvements

- Retry queue for failed tasks
- Dynamic endpoint health checking
- Weighted task assignment based on historical performance
- Prometheus metrics for monitoring
- Connection pooling for high-throughput scenarios

## Running Tests

```bash
go test ./pkg/... -v
```

The tests verify:
- Pull-based distribution (fast workers get more tasks)
- Backoff calculation
- Task generation
