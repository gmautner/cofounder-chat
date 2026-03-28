package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	DatabaseURL     string
	GoogleClientID  string
	GoogleSecret    string
	SessionSecret   string
	BlobStoragePath string
	DevMode         bool
	BaseURL         string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	blobPath := os.Getenv("BLOB_STORAGE_PATH")
	if blobPath == "" {
		blobPath = "./uploads"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	return &Config{
		Port:            port,
		DatabaseURL:     dbURL,
		GoogleClientID:  os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:    os.Getenv("GOOGLE_CLIENT_SECRET"),
		SessionSecret:   os.Getenv("SESSION_SECRET"),
		BlobStoragePath: blobPath,
		DevMode:         os.Getenv("DEV_MODE") == "1",
		BaseURL:         baseURL,
	}, nil
}
