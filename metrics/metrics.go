package metrics

import (
	"fmt"
	"runtime/debug"
	"ssh-plugin/constants"
	"ssh-plugin/models"
	"ssh-plugin/utils"
	"strings"
	"time"

	"ssh-plugin/config"
)

// CollectMetrics collects metrics from a device using SSH for Linux systems
// It executes all configured commands in a single SSH session and parses the output
// Panics are caught and converted to error results to prevent process crashes
func CollectMetrics(device models.Device, timeout time.Duration) (result models.MetricsResult) {
	// Recover from panics to ensure the process continues for other devices
	defer func() {
		if r := recover(); r != nil {
			result = models.NewMetricsError(device.ID, fmt.Sprintf("panic recovered: %v, stack: %s", r, string(debug.Stack())))
		}
	}()

	client, err := utils.CreateSSHClient(device, timeout)
	if err != nil {
		return models.NewMetricsError(device.ID, fmt.Sprintf("SSH connection error: %s", err.Error()))
	}
	defer client.Close()

	cfg, err := config.LoadConfig()
	if err != nil {
		return models.NewMetricsError(device.ID, fmt.Sprintf("Config load error: %s", err.Error()))
	}

	// Prepare single combined command
	var combinedCommands []string

	// Add each command to the combined command list
	for name, cmd := range cfg.Metrics.Commands {
		combinedCommands = append(combinedCommands, fmt.Sprintf("echo '__%s__'; %s", name, cmd))
	}

	finalCommand := strings.Join(combinedCommands, " && ")

	// Execute all commands in one go
	rawOutput, err := utils.ExecuteCommand(client, finalCommand)
	if err != nil {
		return models.NewMetricsError(device.ID, fmt.Sprintf("Command execution error: %s", err.Error()))
	}

	// Parse the output
	lines := strings.Split(strings.TrimSpace(rawOutput), "\n")
	metrics := make(map[string]string)
	var currentMetric string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "__") && strings.HasSuffix(line, "__") {
			currentMetric = strings.TrimSuffix(strings.TrimPrefix(line, "__"), "__")
		} else if currentMetric != "" {
			metrics[currentMetric] = line
			currentMetric = ""
		}
	}

	if len(metrics) == 0 {
		return models.NewMetricsError(device.ID, constants.ErrExecutionFailed)
	}

	return models.NewMetricsSuccess(device.ID, metrics)
}
