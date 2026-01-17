package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RPCURL     string
	DBPath     string
	StartBlock uint64
	EndBlock   *uint64 
	BatchSize  uint64
	Contract   string
	Topic      string
}

func Load() Config {
	// load .env file if present (silently ignore if missing)
	_ = godotenv.Load()
	cfg := Config{
		RPCURL:     getEnv("RPC_URL", "https://ethereum-sepolia-rpc.publicnode.com"),
		DBPath:     getEnv("DB_PATH", "./sepolia.db"),
		BatchSize:  getEnvUint("BATCH_SIZE", 5000),
		Contract:   mustEnv("CONTRACT_ADDRESS"),
		Topic:      mustEnv("EVENT_TOPIC"),
		StartBlock: getEnvUint("START_BLOCK", 0),
	}

	if v := os.Getenv("END_BLOCK"); v != "" {
		end, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Fatalf("invalid END_BLOCK: %v", err)
		}
		cfg.EndBlock = &end
	}

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

func getEnvUint(key string, def uint64) uint64 {
	if v := os.Getenv(key); v != "" {
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Fatalf("invalid %s: %v", key, err)
		}
		return u
	}
	return def
}
