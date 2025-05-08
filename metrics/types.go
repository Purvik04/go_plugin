package metrics

import (
	"ssh-plugin/models"
	"time"
)

// MetricsCollector defines the interface for collecting metrics from different system types
type MetricsCollector interface {
	Collect(device models.Device, timeout time.Duration) models.MetricsResult
}

// GetMetricsCollector returns the appropriate collector based on system type
func GetMetricsCollector(systemType string) MetricsCollector {
	switch systemType {
	case "linux":
		return &LinuxMetricsCollector{}
	default:
		// Placeholder for unsupported system types
		return &UnsupportedMetricsCollector{systemType: systemType}
	}
}

// LinuxMetricsCollector implements MetricsCollector for Linux systems
type LinuxMetricsCollector struct{}

// Collect calls the existing CollectMetrics function for Linux
func (c *LinuxMetricsCollector) Collect(device models.Device, timeout time.Duration) models.MetricsResult {
	return CollectMetrics(device, timeout)
}

// UnsupportedMetricsCollector handles unsupported system types
type UnsupportedMetricsCollector struct {
	systemType string
}

// Collect returns an error for unsupported system types
func (c *UnsupportedMetricsCollector) Collect(device models.Device, timeout time.Duration) models.MetricsResult {
	return models.NewMetricsError(device.ID, "unsupported system type: "+c.systemType)
}
