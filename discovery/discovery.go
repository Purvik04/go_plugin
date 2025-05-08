package discovery

import (
	"fmt"
	"runtime/debug"
	"ssh-plugin/models"
	"ssh-plugin/models/utils"
	"time"
)

// PerformDiscovery attempts to establish an SSH connection to discover if a device is accessible
// It checks port availability, SSH authentication, and executes a test command
// Panics are caught and converted to error results to prevent process crashes
func PerformDiscovery(device models.Device, timeout time.Duration) (result models.DiscoveryResult) {
	// Recover from panics to ensure the process continues for other devices
	defer func() {
		if r := recover(); r != nil {
			result = models.NewDiscoveryResult(device.ID, false, fmt.Sprintf("panic recovered: %v, stack: %s", r, string(debug.Stack())))
		}
	}()

	// Step 1: Check if the port is open
	if !utils.IsPortOpen(device.IP, device.Port, timeout/2) {
		return models.NewDiscoveryResult(device.ID, false, "port")
	}

	// Step 2: Establish SSH connection
	client, err := utils.CreateSSHClient(device, timeout)
	if err != nil {
		return models.NewDiscoveryResult(device.ID, false, "sshAuth")
	}
	defer client.Close()

	// Step 3: Execute a basic command (e.g., uptime)
	session, err := client.NewSession()
	if err != nil {
		return models.NewDiscoveryResult(device.ID, false, "session")
	}
	defer session.Close()

	if err := session.Run("uptime"); err != nil {
		return models.NewDiscoveryResult(device.ID, false, "uptime")
	}

	// Step 4: If all steps succeeded
	return models.NewDiscoveryResult(device.ID, true, "")
}
