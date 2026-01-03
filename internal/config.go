package internal

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            int
	Host            string
	CleanupInterval time.Duration
	DefaultTTL      time.Duration
}

func LoadConfig() *Config {
	cfg := &Config{
		Port:            6379,
		Host:            "0.0.0.0",
		CleanupInterval: 60 * time.Second,
		DefaultTTL:      5 * time.Minute,
	}

	if port := os.Getenv("CAGO_Port"); port != "" {
		if portInt, err := strconv.Atoi(port); err == nil {
			cfg.Port = portInt
		}
	}

	if host := os.Getenv("CAGO_Host"); host != "" {
		cfg.Host = host
	}

	if cleanup := os.Getenv("CAGO_CleanupInterval"); cleanup != "" {
		if cleanupInt, err := strconv.Atoi(cleanup); err == nil {
			cfg.CleanupInterval = time.Duration(cleanupInt) * time.Second
		}
	}

	if ttl := os.Getenv("CAGO_DefaultTTL"); ttl != "" {
		if ttlInt, err := strconv.Atoi(ttl); err == nil {
			cfg.DefaultTTL = time.Duration(ttlInt) * time.Second
		}
	}

	return cfg
}
