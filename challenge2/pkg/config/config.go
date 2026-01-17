package config

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
)

type RPCEndpoint struct {
	Name string
	URL  string
}

type Config struct {
	Contract  common.Address
	Topic     common.Hash
	Endpoints []RPCEndpoint
	BatchSize uint64
}

func DefaultEndpoints() []RPCEndpoint {
	return []RPCEndpoint{
		{Name: "publicnode", URL: "https://ethereum-sepolia-rpc.publicnode.com"},
		{Name: "ankr", URL: "https://rpc.ankr.com/eth_sepolia"},
		{Name: "drpc", URL: "https://sepolia.drpc.org"},
	}
}

func Load() Config {
	return Config{
		Contract:  common.HexToAddress(getEnv("CONTRACT_ADDRESS", "0x761d53b47334bee6612c0bd1467fb881435375b2")),
		Topic:     common.HexToHash(getEnv("EVENT_TOPIC", "0x3e54d0825ed78523037d00a81759237eb436ce774bd546993ee67a1b67b6e766")),
		Endpoints: DefaultEndpoints(),
		BatchSize: 1000,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
