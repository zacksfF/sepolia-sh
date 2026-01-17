# Gateway.fm Core R&D Technical Challenge

## Overview

Two challenges demonstrating Go backend skills for Ethereum infrastructure:

1. **Challenge 1** - Sepolia event indexer with efficient storage
2. **Challenge 2** - Multi-RPC load distribution with pull-based scheduling

## Quick Start

```bash
git clone https://github.com/zacksfF/sepolia-sh.git
cd sepolia-sh
```

### Challenge 1: Event Indexer

```bash
cd challenge1
go build ./cmd/indexer
./indexer
```

[Full documentation](./challenge1/README.md)

### Challenge 2: Multi-RPC Scheduler

```bash
cd challenge2
go build ./cmd/demo
./demo
```

[Full documentation](./challenge2/README.md)
