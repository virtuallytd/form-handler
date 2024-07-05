package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Field represents a form field with its properties
type Field struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	Required         bool     `json:"required"`
	MaxLength        int      `json:"max_length,omitempty"`
	MaxFileSize      int64    `json:"max_file_size,omitempty"`
	AllowedFileTypes []string `json:"allowed_file_types,omitempty"`
}

// RateLimit represents the rate limit configuration for a form
type RateLimit struct {
	Requests int    `json:"requests"`
	Duration string `json:"duration"`
}

// FormConfig holds the configuration for a specific form
type FormConfig struct {
	ReferralURL    string    `json:"referral_url"`
	AllowedOrigins []string  `json:"allowed_origins"`
	RateLimit      RateLimit `json:"rate_limit"`
	Fields         []Field   `json:"fields"`
}

// Config represents the application's configuration
type Config struct {
	Forms map[string]FormConfig `json:"forms"`
}

// Load the configuration from a JSON file
func loadConfig(configPath string) (Config, error) {
	var config Config

	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Errorf("Configuration file does not exist: %s", configPath)
		return config, fmt.Errorf("configuration file does not exist: %s", configPath)
	}

	// Try opening the configuration file
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Errorf("Error opening configuration file: %v", err)
		return config, fmt.Errorf("error opening configuration file: %v", err)
	}
	defer configFile.Close()

	// Read the configuration file
	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		log.Errorf("Error reading configuration file: %v", err)
		return config, fmt.Errorf("error reading configuration file: %v", err)
	}

	// Unmarshal the JSON data into the Config struct
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		log.Errorf("Error unmarshalling configuration JSON: %v", err)
		return config, fmt.Errorf("error unmarshalling configuration JSON: %v", err)
	}

	log.Infof("Configuration loaded successfully from %s", configPath)
	return config, nil
}
