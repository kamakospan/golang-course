// Command golang-sis is a small in-memory text-analysis task processor
// (SIS project, Checkpoint 1).
package main

import (
	"errors"
	"fmt"

	"golang-sis/task"
)

func main() {
	m := task.NewManager(task.TextAnalyzer{})

	inputs := []string{
		"Hello Go world",
		"Go is fast. Go is simple! Is Go fun?",
		"Привет, мир! Қазақстан — Go.",
		"   ", // invalid: whitespace only
	}

	fmt.Println("== Create tasks ==")
	for _, in := range inputs {
		t, err := m.Create(in)
		if err != nil {
			fmt.Printf("rejected %q: %v\n", in, err)
			continue
		}
		fmt.Printf("created task %d (%s): %q\n", t.ID, t.Status, t.Input)
	}

	fmt.Println("\n== Process tasks ==")
	for _, t := range m.List() {
		done, err := m.Process(t.ID)
		if err != nil {
			fmt.Printf("task %d: %v\n", t.ID, err)
			continue
		}
		r := done.Result
		fmt.Printf("task %d (%s): words=%d runes=%d bytes=%d letters=%d sentences=%d top=%q x%d\n",
			done.ID, done.Status, r.Words, r.Runes, r.Bytes, r.Letters, r.Sentences, r.TopWord, r.TopWordCount)
	}

	fmt.Println("\n== Invalid operations ==")
	if _, err := m.Process(1); errors.Is(err, task.ErrInvalidTransition) {
		fmt.Println("process task 1 again:", err)
	}
	if _, err := m.Get(42); errors.Is(err, task.ErrTaskNotFound) {
		fmt.Println("get task 42:", err)
	}
}
