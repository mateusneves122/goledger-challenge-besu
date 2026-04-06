package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	BesuRPCURL      string
	ContractAddress string
	PrivateKey      string
	DBHost          string
	DBPort          uint16
	DBUser          string
	DBPass          string
	DBName          string
	ServerAddr      string
}

func Load() (*Config, error) {
	contractAddr := os.Getenv("CONTRACT_ADDRESS")
	if contractAddr == "" {
		return nil, fmt.Errorf("CONTRACT_ADDRESS is required")
	}

	dbPortStr := getEnv("DB_PORT", "5432")
	dbPort, err := strconv.ParseUint(dbPortStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT %q: %w", dbPortStr, err)
	}

	return &Config{
		BesuRPCURL:      getEnv("BESU_RPC_URL", "http://127.0.0.1:8545"),
		ContractAddress: contractAddr,
		PrivateKey:      getEnv("PRIVATE_KEY", ""),
		DBHost:          getEnv("DB_HOST", ""),
		DBPort:          uint16(dbPort),
		DBUser:          getEnv("DB_USER", ""),
		DBPass:          getEnv("DB_PASS", ""),
		DBName:          getEnv("DB_NAME", ""),
		ServerAddr:      getEnv("SERVER_ADDR", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
