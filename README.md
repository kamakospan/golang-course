# golang-sis — Text Analysis Task Processor

SIS project for **Golang Application Development (KBTU, Fall 2026)**.
This repository contains **Checkpoint 1 — Core Go**: a synchronous, in-memory
application that creates, validates, stores and processes tasks.

## What it does

A **task** is one piece of text to analyze. Processing a task produces statistics:

| Field | Meaning |
| --- | --- |
| `Words` | words separated by whitespace (`strings.Fields`) |
| `Runes` | characters (Unicode code points) — `utf8.RuneCountInString` |
| `Bytes` | size in UTF-8 bytes — `len` (Cyrillic letters take 2 bytes) |
| `Letters` | letters of any alphabet (`unicode.IsLetter`) |
| `Sentences` | sentence-ending marks `.` `!` `?` |
| `TopWord`, `TopWordCount` | most frequent word (lower-cased, punctuation trimmed; ties → alphabetically first) |

Example:

```
Input:  "Go is fast. Go is simple! Is Go fun?"
Result: words=9 runes=36 bytes=36 letters=25 sentences=3 top="go" x3
```

**Validation rule:** empty or whitespace-only text is rejected with `ErrInvalidInput` —
there is nothing to analyze.

## How to run

Requires **Go 1.27**.

```bash
go run .          # demo: creates, processes and prints several tasks
go test ./...     # unit tests
go test -v ./...  # tests with names
go vet ./...
gofmt -l .        # prints nothing when the code is formatted
```

## Task lifecycle

```
Pending ──Start()──> Processing ──Complete()──> Completed
                          │
                          └────Fail()────────> Failed
```

| From | Allowed operation | To |
| --- | --- | --- |
| Pending | `Start` | Processing |
| Processing | `Complete` | Completed |
| Processing | `Fail` (processor returned an error) | Failed |
| Completed, Failed | — (final states) | — |

Any other transition returns `ErrInvalidTransition` and leaves the task unchanged.
For example, processing a task that is already completed is rejected.

## Project structure

```
golang-sis/
├── go.mod
├── main.go            demo program
└── task/
    ├── errors.go      sentinel errors: ErrInvalidInput, ErrTaskNotFound, ErrInvalidTransition
    ├── task.go        Task, Status, lifecycle methods Start / Complete / Fail
    ├── analyzer.go    Processor interface, TextAnalyzer, TextStats
    ├── manager.go     Manager: in-memory storage (map) and Create / Get / Process / List
    └── task_test.go   unit tests (table-driven where appropriate)
```

## Design decisions

- **`Processor` interface.** `Manager` depends on `Processor`, not on `TextAnalyzer`.
  Another domain (or a fake processor in tests — see `failingProcessor`) can be plugged in
  without changing the storage code.
- **Storage is `map[ID]*Task`.** Lookup by ID is O(1). Pointers are stored because a map
  element is not addressable: with `map[ID]Task` the pointer methods `Start`/`Complete`
  could not be called on stored tasks, and changes would apply to a copy.
- **`Get`, `Create`, `Process` and `List` return copies (`Task`, not `*Task`).** Callers cannot
  change a stored task directly (e.g. set `Status = Completed`) and bypass the lifecycle rules.
- **Receivers.** Lifecycle methods (`Start`, `Complete`, `Fail`) and all `Manager` methods use
  pointer receivers because they modify state. `Status.String` and `TextAnalyzer` methods use
  value receivers: they only read small values.
- **Errors.** Expected problems are returned as errors, never `panic`. Each layer adds
  context with `fmt.Errorf("...: %w", err)`, and callers check the cause with `errors.Is`.
- **Deterministic results.** Map iteration order is random, so ties for `TopWord` are
  broken alphabetically — the same input always gives the same result, which keeps tests stable.

## Tests

`go test ./...` covers: successful creation, invalid input (table-driven), retrieval,
retrieval of a nonexistent task, processing, result correctness (table-driven, including
Cyrillic text where runes ≠ bytes), status after processing, invalid transitions
(processing twice, a full transition table), multiple independent tasks, `Get` returning a
copy, and the `Failed` path through a fake processor.
