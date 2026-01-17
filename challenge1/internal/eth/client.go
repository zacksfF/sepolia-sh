package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	*ethclient.Client
}

func Dial(ctx context.Context, rpcURL string) (*Client, error) {
	c, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	return &Client{Client: c}, nil
}
