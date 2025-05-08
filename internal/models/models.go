package models

import (
	"time"
)

// Credentials stores username and password for SSH connection
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Device represents a device to be monitored or discovered
type Device struct {
	ID          int         `json:"id"`
	IP          string      `json:"ip"`
	SystemType  string      `json:"system_type"` // Added to support system_type field
	Port        int         `json:"port"`
	Credentials Credentials `json:"credentials"`
}

// MetricsResult represents the result of metrics collection
type MetricsResult struct {
	ID       int               `json:"id"`
	Success  bool              `json:"success"`
	Metrics  map[string]string `json:"metrics"`
	PolledAt string            `json:"polled_at"`
}

// DiscoveryResult represents the result of SSH discovery
type DiscoveryResult struct {
	ID      int    `json:"id"`
	Success bool   `json:"success"`
	Step    string `json:"step"`
}

// NewMetricsError creates a new metrics result with an error
func NewMetricsError(id int, errMsg string) MetricsResult {
	return MetricsResult{
		ID:      id,
		Success: false,
		Metrics: map[string]string{
			"error": errMsg,
		},
		PolledAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewMetricsSuccess creates a new successful metrics result
func NewMetricsSuccess(id int, data map[string]string) MetricsResult {
	return MetricsResult{
		ID:       id,
		Success:  true,
		Metrics:  data,
		PolledAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewDiscoveryResult creates a new discovery result
func NewDiscoveryResult(id int, success bool, step string) DiscoveryResult {
	return DiscoveryResult{
		ID:      id,
		Success: success,
		Step:    step,
	}
}
