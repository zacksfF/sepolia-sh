package rpc

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	name   string
	client *ethclient.Client
	stats  *Stats
}

// Stats tracks per-RPC performance metrics.
type Stats struct {
	Name          string
	TotalRequests atomic.Int64
	Failures      atomic.Int64
	TotalLatency  atomic.Int64 // nanoseconds
	mu            sync.RWMutex
}

// NewClient creates a new RPC client wrapper.
func NewClient(ctx context.Context, name, url string) (*Client, error) {
	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}

	return &Client{
		name:   name,
		client: client,
		stats:  &Stats{Name: name},
	}, nil
}

// Name returns the client's identifier.
func (c *Client) Name() string {
	return c.name
}

// Stats returns the client's statistics.
func (c *Client) Stats() *Stats {
	return c.stats
}

// FilterLogs fetches logs for the given query, tracking latency.
func (c *Client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	start := time.Now()
	logs, err := c.client.FilterLogs(ctx, query)
	latency := time.Since(start)

	c.stats.TotalRequests.Add(1)
	c.stats.TotalLatency.Add(int64(latency))

	if err != nil {
		c.stats.Failures.Add(1)
	}

	return logs, err
}

// BlockNumber returns the latest block number.
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}

// Close closes the underlying client.
func (c *Client) Close() {
	c.client.Close()
}

// GetStats returns a snapshot of the statistics.
func (s *Stats) GetStats() (requests, failures int64, avgLatency time.Duration) {
	requests = s.TotalRequests.Load()
	failures = s.Failures.Load()
	totalLatency := s.TotalLatency.Load()

	if requests > 0 {
		avgLatency = time.Duration(totalLatency / requests)
	}
	return
}

// FilterQuery creates a filter query for the given block range.
func FilterQuery(contract common.Address, topic common.Hash, from, to uint64) ethereum.FilterQuery {
	return ethereum.FilterQuery{
		FromBlock: toBigInt(from),
		ToBlock:   toBigInt(to),
		Addresses: []common.Address{contract},
		Topics:    [][]common.Hash{{topic}},
	}
}

func toBigInt(v uint64) *big.Int {
	return new(big.Int).SetUint64(v)
}
