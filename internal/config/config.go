package config

import (
	"os"
	"time"

	"github.com/As71er/once/internal/utils"
)

type Config struct {
	Addr               string
	DBDSN              string
	ArchivePath        string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	MaxUploadMemoryMiB int64
}

func Load() *Config {

	return &Config{
		Addr:               getEnv("ADDR", "127.0.0.1:8080"),
		DBDSN:              getEnv("DB_DSN", "file:./data/once.db?_pragma=foreign_keys(1)"), // SQLITE
		ArchivePath:        getEnv("ARCHIVE_PATH", "./archive"),
		ReadTimeout:        getEnvDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:       getEnvDuration("WRITE_TIMEOUT", 20*time.Second),
		IdleTimeout:        getEnvDuration("IDLE_TIMEOUT", time.Minute),
		MaxUploadMemoryMiB: int64(getEnvInt("MAX_UPLOAD_MEMORY_MIB", 20)),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	venv := utils.ParseInt(v, fallback)

	return venv
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
