package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	Port     int    `json:"port"`
	SSLMode  string `json:"sslmode"`
	TimeZone string `json:"timezone"`
}

// Use this to construct the DSN
func (c *DBConfig) BuildDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		c.Host, c.User, c.Password, c.DBName, c.Port, c.SSLMode, c.TimeZone)
}

// Use this to construct the DSN (no dbname)
func (c *DBConfig) BuildDSNnodb() string {
	return fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=%s TimeZone=%s",
		c.Host, c.User, c.Password, c.Port, c.SSLMode, c.TimeZone)
}

// DatabaseConfig holds the database connection configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	Port     int    `json:"port"`
	SSLMode  string `json:"sslmode"`
	TimeZone string `json:"timezone"`
}

// Config holds the application configuration
type Config struct {
	DB DatabaseConfig `json:"database"`
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
