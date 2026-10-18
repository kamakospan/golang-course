package task

import (
	"fmt"
	"sort"
)

// Manager creates, stores, retrieves and processes tasks in memory.
type Manager struct {
	processor Processor
	tasks     map[ID]*Task
	nextID    ID
}

// NewManager returns an empty Manager that processes tasks with p.
func NewManager(p Processor) *Manager {
	return &Manager{
		processor: p,
		tasks:     make(map[ID]*Task),
		nextID:    1,
	}
}

// The map stores *Task, not Task: an element of a map is not addressable,
// so with map[ID]Task we could not call pointer methods like Start()
// on it or change its fields in place.

// Create validates the input, stores a new pending task and returns a copy of it.
func (m *Manager) Create(input string) (Task, error) {
	if err := m.processor.Validate(input); err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	t := &Task{ID: m.nextID, Input: input, Status: StatusPending}
	m.tasks[t.ID] = t
	m.nextID++
	return *t, nil
}

// Get returns a copy of the task with the given ID.
// Returning a copy (not *Task) means callers cannot change the stored
// task behind the Manager's back, e.g. set Status to Completed directly.
func (m *Manager) Get(id ID) (Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("get task %d: %w", id, ErrTaskNotFound)
	}
	return *t, nil
}

// Process runs a pending task through its lifecycle and stores the result.
func (m *Manager) Process(id ID) (Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("process task %d: %w", id, ErrTaskNotFound)
	}
	if err := t.Start(); err != nil {
		return Task{}, fmt.Errorf("process: %w", err)
	}

	result, err := m.processor.Process(t.Input)
	if err != nil {
		_ = t.Fail(err) // cannot fail: we just moved the task to Processing
		return *t, fmt.Errorf("process task %d: %w", id, err)
	}
	if err := t.Complete(result); err != nil {
		return Task{}, fmt.Errorf("process: %w", err)
	}
	return *t, nil
}

// List returns copies of all tasks, ordered by ID.
func (m *Manager) List() []Task {
	out := make([]Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
