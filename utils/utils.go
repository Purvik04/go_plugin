package utils

import (
	"fmt"
	"net"
	"runtime/debug"
	"ssh-plugin/constants"
	"ssh-plugin/models"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// CreateSSHClient creates a new SSH client for the given device
// Panics are caught and converted to errors to prevent process crashes
func CreateSSHClient(device models.Device, timeout time.Duration) (client *ssh.Client, err error) {
	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			client = nil
			err = fmt.Errorf("panic recovered: %v, stack: %s", r, string(debug.Stack()))
		}
	}()

	// Use default SSH port if not specified
	port := device.Port
	if port == 0 {
		port = constants.DefaultSSHPort
	}

	// Set up SSH client configuration
	config := &ssh.ClientConfig{
		User: device.Credentials.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(device.Credentials.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", device.IP, port)
	client, err = ssh.Dial("tcp", addr, config)
	if err != nil {
		// Log the error before returning it
		if strings.Contains(err.Error(), "timeout") {
			return nil, fmt.Errorf("%s: %s", constants.ErrTimeout, err.Error())
		}
		if strings.Contains(err.Error(), "authentication") {
			return nil, fmt.Errorf("%s: %s", constants.ErrAuthFailed, err.Error())
		}
		// General connection failure
		return nil, fmt.Errorf("%s: %s", constants.ErrConnectionFailed, err.Error())
	}

	return client, nil
}

// ExecuteCommand executes a command on the SSH client
// Panics are caught and converted to errors to prevent process crashes
func ExecuteCommand(client *ssh.Client, command string) (output string, err error) {
	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			output = ""
			err = fmt.Errorf("panic recovered: %v, stack: %s", r, string(debug.Stack()))
		}
	}()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	outputBytes, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("%s: %v", constants.ErrExecutionFailed, err)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

// IsPortOpen checks if a port is open on a host
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
