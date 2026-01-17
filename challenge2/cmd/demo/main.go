package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zacksfF/sepolia-sh/ch2/pkg/config"
	"github.com/zacksfF/sepolia-sh/ch2/pkg/rpc"
	"github.com/zacksfF/sepolia-sh/ch2/pkg/scheduler"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	cfg := config.Load()

	// Setup context with cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// RPC clients
	log.Println("Connecting to RPC endpoints...")
	var clients []*rpc.Client
	for _, ep := range cfg.Endpoints {
		client, err := rpc.NewClient(ctx, ep.Name, ep.URL)
		if err != nil {
			log.Printf("Warning: failed to connect to %s: %v", ep.Name, err)
			continue
		}
		clients = append(clients, client)
		log.Printf("Connected to %s", ep.Name)
	}

	if len(clients) == 0 {
		log.Fatal("No RPC endpoints available")
	}
	defer func() {
		for _, c := range clients {
			c.Close()
		}
	}()

	latestBlock, err := clients[0].BlockNumber(ctx)
	if err != nil {
		log.Fatalf("Failed to get latest block: %v", err)
	}
	log.Printf("Latest block: %d", latestBlock)

	// Limit demo to last 50k blocks to keep it reasonable
	startBlock := uint64(0)
	if latestBlock > 50000 {
		startBlock = latestBlock - 50000
	}
	log.Printf("Demo: scanning blocks %d to %d", startBlock, latestBlock)

	sched := scheduler.New(clients, scheduler.Config{
		Contract:  cfg.Contract,
		Topic:     cfg.Topic,
		BatchSize: cfg.BatchSize,
	})

	start := time.Now()
	totalLogs, err := sched.Run(ctx, startBlock, latestBlock)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("Scheduler stopped: %v", err)
	}

	log.Printf("=== Summary ===")
	log.Printf("Blocks scanned: %d", latestBlock-startBlock+1)
	log.Printf("Total logs found: %d", totalLogs)
	log.Printf("Time elapsed: %v", elapsed)
	log.Printf("Throughput: %.0f blocks/sec", float64(latestBlock-startBlock+1)/elapsed.Seconds())
}
