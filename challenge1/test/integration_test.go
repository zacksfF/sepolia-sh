package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/sepolia-sh/ch1/internal/eth"
	"github.com/zacksfF/sepolia-sh/ch1/internal/indexer"
	"github.com/zacksfF/sepolia-sh/ch1/internal/storage/bolt"
)

func TestIndexer_Sepolia(t *testing.T) {
	if os.Getenv("RPC_URL") == "" {
		t.Skip("RPC_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ethClient, err := eth.Dial(ctx, os.Getenv("RPC_URL"))
	if err != nil {
		t.Fatal(err)
	}

	store, err := bolt.Open("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test.db")
	defer store.Close()

	idx := indexer.New(
		ethClient,
		store,
		common.HexToAddress(os.Getenv("CONTRACT_ADDRESS")),
		common.HexToHash(os.Getenv("EVENT_TOPIC")),
	)

	err = idx.Run(ctx, 0, nil, 1000) 
	if err != nil {
		t.Fatal(err)
	}

	next, err := store.GetNextIndex(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if next == 0 {
		t.Fatal("expected at least one indexed event")
	}
}
