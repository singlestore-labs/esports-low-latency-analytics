package src

import (
	"log"
	"os"

	"cuelang.org/go/pkg/strconv"
)

type ProcessorConfig struct {
	Verbose    int
	NumWorkers int
	ReplayDir  string
}

type PlayerConfig struct {
	Verbose   int
	ReplayDir string
	IconDir   string
	Port      int
	GinMode   string // move to env (GIN_ENV)
}

type SinglestoreConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func SinglestoreConfigFromEnv() SinglestoreConfig {
	return SinglestoreConfig{
		Host:     envOrDefault("SINGLESTORE_HOST", "127.0.0.1"),
		Port:     mustParseInt(envOrDefault("SINGLESTORE_PORT", "3306")),
		Username: envOrDefault("SINGLESTORE_USER", "root"),
		Password: envOrDefault("SINGLESTORE_PASSWORD", ""),
		Database: envOrDefault("SINGLESTORE_DATABASE", "sc2"),
	}
}

func envOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return i
}
