package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type LogFetcher struct {
	client *Client
}

func NewLogFetcher(client *Client) *LogFetcher {
	return &LogFetcher{client: client}
}

func (f *LogFetcher) Fetch(
	ctx context.Context,
	contract common.Address,
	topic common.Hash,
	fromBlock uint64,
	toBlock uint64,
) ([]types.Log, error) {

	query := ethereum.FilterQuery{
		FromBlock: uint64ToBig(fromBlock),
		ToBlock:   uint64ToBig(toBlock),
		Addresses: []common.Address{contract},
		Topics:    [][]common.Hash{{topic}},
	}

	return f.client.FilterLogs(ctx, query)
}

// helper
func uint64ToBig(v uint64) *big.Int {
	b := new(big.Int)
	b.SetUint64(v)
	return b
}
