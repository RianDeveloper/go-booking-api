package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort            string
	PostgresDSN         string
	KafkaBrokers        []string
	KafkaTopic          string
	OutboxBatchSize     int
	OutboxPollInterval  time.Duration
	ShutdownTimeout     time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:           getOrDefault("HTTP_PORT", "8080"),
		PostgresDSN:        os.Getenv("POSTGRES_DSN"),
		KafkaBrokers:       splitCSV(getOrDefault("KAFKA_BROKERS", "kafka:9092")),
		KafkaTopic:         getOrDefault("KAFKA_TOPIC", "booking-events"),
		OutboxBatchSize:    getIntOrDefault("OUTBOX_BATCH_SIZE", 100),
		OutboxPollInterval: getDurationOrDefault("OUTBOX_POLL_INTERVAL", 2*time.Second),
		ShutdownTimeout:    getDurationOrDefault("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN is required")
	}
	return cfg, nil
}

func getOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getIntOrDefault(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func getDurationOrDefault(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return v
}

func splitCSV(raw string) []string {
	out := make([]string, 0)
	cur := ""
	for i := 0; i < len(raw); i++ {
		if raw[i] == ',' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		if raw[i] == ' ' {
			continue
		}
		cur += string(raw[i])
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
