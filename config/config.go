package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr       string
	KafkaBrokers   string
	KafkaTopic     string
	KafkaGroupID   string
	BucketCount    int
	BucketDuration time.Duration
	CacheRefresh   time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		KafkaBrokers:   getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:     getEnv("KAFKA_TOPIC", "search-events"),
		KafkaGroupID:   getEnv("KAFKA_GROUP_ID", "trending-service"),
		BucketCount:    getEnvInt("BUCKET_COUNT", 30),
		BucketDuration: getEnvDuration("BUCKET_DURATION_SEC", 10) * time.Second,
		CacheRefresh:   getEnvDuration("CACHE_REFRESH_MS", 500) * time.Millisecond,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Duration(n)
		}
	}
	return def
}
