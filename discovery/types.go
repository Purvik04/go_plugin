package discovery

import (
	"ssh-plugin/models"
	"time"
)

// DiscoveryPerformer defines the interface for performing SSH discovery
type DiscoveryPerformer interface {
	Perform(device models.Device, timeout time.Duration) models.DiscoveryResult
}

// GetDiscoveryPerformer returns the appropriate performer based on system type
func GetDiscoveryPerformer(systemType string) DiscoveryPerformer {
	switch systemType {
	case "linux":
		return &LinuxDiscoveryPerformer{}
	default:
		// Placeholder for unsupported system types
		return &UnsupportedDiscoveryPerformer{systemType: systemType}
	}
}

// LinuxDiscoveryPerformer implements DiscoveryPerformer for Linux systems
type LinuxDiscoveryPerformer struct{}

// Perform calls the existing PerformDiscovery function for Linux
func (p *LinuxDiscoveryPerformer) Perform(device models.Device, timeout time.Duration) models.DiscoveryResult {
	return PerformDiscovery(device, timeout)
}

// UnsupportedDiscoveryPerformer handles unsupported system types
type UnsupportedDiscoveryPerformer struct {
	systemType string
}

// Perform returns an error for unsupported system types
func (p *UnsupportedDiscoveryPerformer) Perform(device models.Device, timeout time.Duration) models.DiscoveryResult {
	return models.NewDiscoveryResult(device.ID, false, "unsupported system type: "+p.systemType)
}
