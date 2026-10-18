package task

import "fmt"

// ID uniquely identifies a task inside a Manager.
type ID int

// Status is the current lifecycle stage of a task.
type Status int

// Task lifecycle:
//
//	Pending ──Start()──> Processing ──Complete()──> Completed
//	                          │
//	                          └────Fail()────────> Failed
//
// Completed and Failed are final: no transition leaves them.
const (
	StatusPending Status = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
)

// String implements fmt.Stringer, so statuses print as words, not numbers.
// Value receiver: Status is a small int and the method only reads it.
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusProcessing:
		return "processing"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	default:
		return fmt.Sprintf("Status(%d)", int(s))
	}
}

// Task is one unit of work: the input text, its lifecycle status,
// and the statistics produced by processing it.
type Task struct {
	ID     ID
	Input  string
	Status Status
	Result TextStats // zero value until the task is completed
	Err    string    // reason, when Status is StatusFailed
}

// The methods below change the task, so they use pointer receivers:
// with a value receiver they would change a copy and the stored task
// would never move to the next status.

// Start moves a pending task to processing.
func (t *Task) Start() error {
	if t.Status != StatusPending {
		return fmt.Errorf("task %d: start from %s: %w", t.ID, t.Status, ErrInvalidTransition)
	}
	t.Status = StatusProcessing
	return nil
}

// Complete stores the result and moves a processing task to completed.
func (t *Task) Complete(r TextStats) error {
	if t.Status != StatusProcessing {
		return fmt.Errorf("task %d: complete from %s: %w", t.ID, t.Status, ErrInvalidTransition)
	}
	t.Result = r
	t.Status = StatusCompleted
	return nil
}

// Fail records why processing failed and moves a processing task to failed.
func (t *Task) Fail(reason error) error {
	if t.Status != StatusProcessing {
		return fmt.Errorf("task %d: fail from %s: %w", t.ID, t.Status, ErrInvalidTransition)
	}
	t.Err = reason.Error()
	t.Status = StatusFailed
	return nil
}
