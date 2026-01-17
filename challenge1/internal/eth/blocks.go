package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BlockFetcher struct {
	client *Client
}

func NewBlockFetcher(client *Client) *BlockFetcher {
	return &BlockFetcher{client: client}
}

func (f *BlockFetcher) ByHash(
	ctx context.Context,
	hash common.Hash,
) (*types.Block, error) {
	return f.client.BlockByHash(ctx, hash)
}
