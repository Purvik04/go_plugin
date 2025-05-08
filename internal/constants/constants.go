package constants

// SSH related constants
const (
	DefaultSSHPort = 22
	MinSSHTimeout  = 5  // Minimum SSH timeout in seconds
	MaxSSHTimeout  = 60 // Maximum SSH timeout in seconds
)

// Error messages
const (
	ErrConnectionFailed  = "failed to establish SSH connection"
	ErrAuthFailed        = "authentication failed"
	ErrExecutionFailed   = "command execution failed"
	ErrTimeout           = "operation timed out"
	ErrInvalidParameters = "invalid parameters"
)

// Command timeout in seconds
const (
	CommandTimeout = 10 // Default timeout for individual command execution
)
