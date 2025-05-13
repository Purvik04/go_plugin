package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config represents the application configuration
type Config struct {
	SSH struct {
		Timeout int `json:"timeout"` // Timeout in seconds
	} `json:"ssh"`
	Metrics struct {
		Commands map[string]string `json:"commands"`
	} `json:"metrics"`
	Encryption struct {
		Key string `json:"key"` // Hex-encoded AES key
	} `json:"encryption"`
}

// LoadConfig loads configuration from config.json with safe defaults
func LoadConfig() (*Config, error) {
	// Set default configuration
	defaultConfig := &Config{}
	defaultConfig.SSH.Timeout = 5 // 5 seconds default
	defaultConfig.Metrics.Commands = map[string]string{
		"hostname":  "hostname",
		"uptime":    "uptime -p",
		"cpu":       "top -bn1 | grep 'Cpu(s)' | awk '{print $2 + $4}'",
		"memory":    "free -g | awk '/Mem:/ {print $3}'",
		"disk":      "df -BG / | awk 'NR==2 {print $3}'",
		"processes": "ps aux | wc -l",
	}
	defaultConfig.Encryption.Key = "" // No default key for security

	// If config.json does not exist, return defaults
	configFile, err := os.Open("/home/purvik/IdeaProjectsUltimate/nms-main/go/config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, nil
		}
		return nil, err
	}
	defer configFile.Close()

	// Decode config.json into temp struct
	userConfig := &Config{}
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(userConfig); err != nil {
		return nil, err
	}

	// Merge userConfig over defaultConfig safely
	if userConfig.SSH.Timeout > 0 {
		defaultConfig.SSH.Timeout = userConfig.SSH.Timeout
	}

	if userConfig.Metrics.Commands != nil {
		for key, defaultCmd := range defaultConfig.Metrics.Commands {
			if userCmd, exists := userConfig.Metrics.Commands[key]; exists && userCmd != "" {
				defaultConfig.Metrics.Commands[key] = userCmd
			} else {
				defaultConfig.Metrics.Commands[key] = defaultCmd
			}
		}
	}

	if userConfig.Encryption.Key != "" {
		defaultConfig.Encryption.Key = userConfig.Encryption.Key
	}

	return defaultConfig, nil
}

// GetSSHTimeout returns the SSH timeout as a time.Duration
func (c *Config) GetSSHTimeout() time.Duration {
	return time.Duration(c.SSH.Timeout) * time.Second
}
