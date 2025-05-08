package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	SSH struct {
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"ssh"`
	Metrics struct {
		Commands map[string]string `yaml:"commands"`
	} `yaml:"metrics"`
	Encryption struct {
		Key string `yaml:"key"` // Hex-encoded AES key
	} `yaml:"encryption"`
}

// LoadConfig loads configuration from config.yaml with safe defaults
func LoadConfig() (*Config, error) {
	// Set default configuration
	defaultConfig := &Config{}
	defaultConfig.SSH.Timeout = 5 * time.Second
	defaultConfig.Metrics.Commands = map[string]string{
		"hostname":  "hostname",
		"uptime":    "uptime -p",
		"cpu":       "top -bn1 | grep 'Cpu(s)' | awk '{print $2 + $4}'",
		"memory":    "free -g | awk '/Mem:/ {print $3}'",
		"disk":      "df -BG / | awk 'NR==2 {print $3}'",
		"processes": "ps aux | wc -l",
	}
	defaultConfig.Encryption.Key = "" // No default key for security

	// If config.yaml does not exist, return defaults
	configFile, err := os.Open("/home/purvik/IdeaProjectsUltimate/nms-main/go/config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, nil
		}
		return nil, err
	}
	defer configFile.Close()

	// Decode config.yaml into temp struct
	userConfig := &Config{}
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(userConfig); err != nil {
		return nil, err
	}

	// Merge userConfig over defaultConfig safely
	if userConfig.SSH.Timeout != 0 {
		defaultConfig.SSH.Timeout = userConfig.SSH.Timeout
	}

	for key, defaultCmd := range defaultConfig.Metrics.Commands {
		if userCmd, exists := userConfig.Metrics.Commands[key]; exists && userCmd != "" {
			defaultConfig.Metrics.Commands[key] = userCmd
		} else {
			defaultConfig.Metrics.Commands[key] = defaultCmd
		}
	}

	if userConfig.Encryption.Key != "" {
		defaultConfig.Encryption.Key = userConfig.Encryption.Key
	}

	return defaultConfig, nil
}
