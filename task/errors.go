package task

import "errors"

// Sentinel errors. Callers check them with errors.Is, so every layer
// can add context with fmt.Errorf("...: %w", err) without breaking the check.
var (
	// ErrInvalidInput is returned when task input fails validation.
	ErrInvalidInput = errors.New("invalid task input")

	// ErrTaskNotFound is returned when no task has the requested ID.
	ErrTaskNotFound = errors.New("task not found")

	// ErrInvalidTransition is returned when an operation is not allowed
	// in the task's current status (for example, processing a completed task).
	ErrInvalidTransition = errors.New("invalid status transition")
)
