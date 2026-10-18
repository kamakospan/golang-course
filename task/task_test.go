package task

import (
	"errors"
	"testing"
)

func newTestManager() *Manager { return NewManager(TextAnalyzer{}) }

// 1. Successful task creation.
func TestCreate_Success(t *testing.T) {
	m := newTestManager()

	got, err := m.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if got.Status != StatusPending {
		t.Errorf("Status = %s, want %s", got.Status, StatusPending)
	}
	if got.Input != "Hello Go world" {
		t.Errorf("Input = %q, want %q", got.Input, "Hello Go world")
	}
}

// 2. Invalid task input (table-driven).
func TestCreate_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"spaces only", "    ", true},
		{"tabs and newlines", "\t\n \n", true},
		{"single word", "go", false},
		{"cyrillic text", "Привет мир", false},
		{"text with surrounding spaces", "  hi  ", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestManager()
			_, err := m.Create(tc.input)

			if tc.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("Create(%q) error = %v, want ErrInvalidInput", tc.input, err)
				}
				if n := len(m.List()); n != 0 {
					t.Errorf("rejected task was stored: %d tasks in manager", n)
				}
				return
			}
			if err != nil {
				t.Fatalf("Create(%q) unexpected error: %v", tc.input, err)
			}
		})
	}
}

// 3. Successful task retrieval.
func TestGet_Success(t *testing.T) {
	m := newTestManager()
	created, _ := m.Create("some text")

	got, err := m.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}
	if got.ID != created.ID || got.Input != "some text" {
		t.Errorf("Get returned %+v, want task %d with input %q", got, created.ID, "some text")
	}
}

// 4. Retrieval of a nonexistent task.
func TestGet_NotFound(t *testing.T) {
	m := newTestManager()

	_, err := m.Get(999)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("Get(999) error = %v, want ErrTaskNotFound", err)
	}

	_, err = m.Process(999)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("Process(999) error = %v, want ErrTaskNotFound", err)
	}
}

// 5, 6, 7. Successful processing, correct result, correct state.
func TestProcess_ResultAndState(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TextStats
	}{
		{
			name:  "simple english",
			input: "Hello Go world",
			want:  TextStats{Words: 3, Runes: 14, Bytes: 14, Letters: 12, Sentences: 0, TopWord: "go", TopWordCount: 1},
		},
		{
			name:  "repeated word and punctuation",
			input: "Go is fast. Go is simple! Is Go fun?",
			want:  TextStats{Words: 9, Runes: 36, Bytes: 36, Letters: 25, Sentences: 3, TopWord: "go", TopWordCount: 3},
		},
		{
			name:  "cyrillic: runes differ from bytes",
			input: "мир мир",
			want:  TextStats{Words: 2, Runes: 7, Bytes: 13, Letters: 6, Sentences: 0, TopWord: "мир", TopWordCount: 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestManager()
			created, err := m.Create(tc.input)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}

			done, err := m.Process(created.ID)
			if err != nil {
				t.Fatalf("Process: %v", err)
			}
			if done.Status != StatusCompleted {
				t.Errorf("Status after Process = %s, want %s", done.Status, StatusCompleted)
			}
			if done.Result != tc.want {
				t.Errorf("Result = %+v, want %+v", done.Result, tc.want)
			}

			// The result must be stored, not only returned.
			stored, _ := m.Get(created.ID)
			if stored.Status != StatusCompleted || stored.Result != tc.want {
				t.Errorf("stored task = %+v, want completed with %+v", stored, tc.want)
			}
		})
	}
}

// 8. Invalid operation / state transition.
func TestProcess_TwiceIsInvalid(t *testing.T) {
	m := newTestManager()
	created, _ := m.Create("Hello Go world")
	first, _ := m.Process(created.ID)

	_, err := m.Process(created.ID)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("second Process error = %v, want ErrInvalidTransition", err)
	}

	// The failed attempt must not change the stored task.
	stored, _ := m.Get(created.ID)
	if stored != first {
		t.Errorf("task changed after rejected Process: got %+v, want %+v", stored, first)
	}
}

func TestTaskTransitions(t *testing.T) {
	tests := []struct {
		name string
		from Status
		op   func(*Task) error
		ok   bool
	}{
		{"start pending", StatusPending, (*Task).Start, true},
		{"start processing", StatusProcessing, (*Task).Start, false},
		{"start completed", StatusCompleted, (*Task).Start, false},
		{"complete processing", StatusProcessing, func(t *Task) error { return t.Complete(TextStats{}) }, true},
		{"complete pending", StatusPending, func(t *Task) error { return t.Complete(TextStats{}) }, false},
		{"fail processing", StatusProcessing, func(t *Task) error { return t.Fail(errors.New("x")) }, true},
		{"fail completed", StatusCompleted, func(t *Task) error { return t.Fail(errors.New("x")) }, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := &Task{ID: 1, Status: tc.from}
			err := tc.op(task)
			if tc.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.ok {
				if !errors.Is(err, ErrInvalidTransition) {
					t.Fatalf("error = %v, want ErrInvalidTransition", err)
				}
				if task.Status != tc.from {
					t.Errorf("status changed to %s after rejected transition", task.Status)
				}
			}
		})
	}
}

// 9. Multiple independent tasks.
func TestMultipleTasksAreIndependent(t *testing.T) {
	m := newTestManager()
	a, _ := m.Create("one two three")
	b, _ := m.Create("go go")
	c, _ := m.Create("left pending")

	if a.ID == b.ID || b.ID == c.ID || a.ID == c.ID {
		t.Fatalf("IDs are not unique: %d %d %d", a.ID, b.ID, c.ID)
	}

	if _, err := m.Process(a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Process(b.ID); err != nil {
		t.Fatal(err)
	}

	gotA, _ := m.Get(a.ID)
	gotB, _ := m.Get(b.ID)
	gotC, _ := m.Get(c.ID)

	if gotA.Result.Words != 3 || gotB.Result.Words != 2 {
		t.Errorf("results mixed up: A words=%d (want 3), B words=%d (want 2)", gotA.Result.Words, gotB.Result.Words)
	}
	if gotB.Result.TopWord != "go" || gotB.Result.TopWordCount != 2 {
		t.Errorf("B top word = %q x%d, want \"go\" x2", gotB.Result.TopWord, gotB.Result.TopWordCount)
	}
	if gotC.Status != StatusPending || gotC.Result != (TextStats{}) {
		t.Errorf("unprocessed task C changed: %+v", gotC)
	}
	if n := len(m.List()); n != 3 {
		t.Errorf("List() has %d tasks, want 3", n)
	}
}

// Get returns a copy: changing it must not change the stored task.
func TestGet_ReturnsCopy(t *testing.T) {
	m := newTestManager()
	created, _ := m.Create("hello")

	copyOfTask, _ := m.Get(created.ID)
	copyOfTask.Status = StatusCompleted

	stored, _ := m.Get(created.ID)
	if stored.Status != StatusPending {
		t.Errorf("stored status = %s, want pending: Get must return a copy", stored.Status)
	}
}

// failingProcessor lets us test the Failed path through the Processor interface.
type failingProcessor struct{}

func (failingProcessor) Validate(string) error { return nil }
func (failingProcessor) Process(string) (TextStats, error) {
	return TextStats{}, errors.New("analyzer crashed")
}

func TestProcess_ProcessorErrorMarksFailed(t *testing.T) {
	m := NewManager(failingProcessor{})
	created, _ := m.Create("anything")

	got, err := m.Process(created.ID)
	if err == nil {
		t.Fatal("Process: expected an error")
	}
	if got.Status != StatusFailed || got.Err != "analyzer crashed" {
		t.Errorf("task = %+v, want status failed with reason", got)
	}
	if _, err := m.Process(created.ID); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("processing a failed task: error = %v, want ErrInvalidTransition", err)
	}
}
