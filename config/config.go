package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// DatabaseConfig holds the database connection configuration
type DatabaseConfig struct {
	DSN string `json:"dsn"`
}

// Config holds the application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
}

// LoadConfig reads the config from the given file and returns the config struct
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("could not decode config: %v", err)
	}

	return &config, nil
}
