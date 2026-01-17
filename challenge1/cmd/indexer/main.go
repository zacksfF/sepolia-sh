package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/sepolia-sh/ch1/config"
	"github.com/zacksfF/sepolia-sh/ch1/internal/eth"
	"github.com/zacksfF/sepolia-sh/ch1/internal/indexer"
	"github.com/zacksfF/sepolia-sh/ch1/internal/storage/bolt"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	ethClient, err := eth.Dial(ctx, cfg.RPCURL)
	if err != nil {
		log.Fatalf("failed to connect to ethereum rpc: %v", err)
	}

	store, err := bolt.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer store.Close()

	idx := indexer.New(
		ethClient,
		store,
		common.HexToAddress(cfg.Contract),
		common.HexToHash(cfg.Topic),
	)

	if err := idx.Run(
		ctx,
		cfg.StartBlock,
		cfg.EndBlock, 
		cfg.BatchSize,
	); err != nil {
		log.Fatalf("indexer failed: %v", err)
	}

	log.Println("indexer finished successfully")
}
