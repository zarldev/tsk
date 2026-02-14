package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Store abstracts task persistence, enabling pluggable backends.
type Store interface {
	Load() ([]Task, error)
	Save([]Task) error
}

// FileStore persists tasks as JSON in a local file.
type FileStore struct {
	Path string
}

// NewFileStore returns a FileStore that reads/writes the given path.
func NewFileStore(path string) *FileStore {
	return &FileStore{Path: path}
}

// Load reads tasks from the JSON file.
// Returns an empty slice if the file does not exist.
func (s *FileStore) Load() ([]Task, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", s.Path, err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}
	return tasks, nil
}

// Save writes tasks to the JSON file.
func (s *FileStore) Save(tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}
	if err := os.WriteFile(s.Path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", s.Path, err)
	}
	return nil
}

// DefaultPath returns the default storage path (~/.tasks.json).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".tasks.json"), nil
}

// nextID returns the next auto-incrementing ID.
func nextID(tasks []Task) int {
	max := 0
	for _, t := range tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}

// Add creates a new task with the given title and appends it to the list.
func Add(tasks []Task, title string) []Task {
	t := Task{
		ID:        nextID(tasks),
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
	}
	return append(tasks, t)
}

// Done marks the task with the given ID as done.
// Returns an error if the ID is not found.
func Done(tasks []Task, id int) error {
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = true
			now := time.Now()
			tasks[i].CompletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("task %d: not found", id)
}

// Find returns a pointer to the task with the given ID, or nil if not found.
func Find(tasks []Task, id int) *Task {
	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i]
		}
	}
	return nil
}

// Remove deletes the task with the given ID and returns the updated slice.
// Returns an error if the ID is not found.
func Remove(tasks []Task, id int) ([]Task, error) {
	for i, t := range tasks {
		if t.ID == id {
			return append(tasks[:i], tasks[i+1:]...), nil
		}
	}
	return tasks, fmt.Errorf("task %d: not found", id)
}

// Filter controls which tasks List returns.
type Filter int

const (
	FilterAll     Filter = iota
	FilterDone           // only completed tasks
	FilterPending        // only incomplete tasks
)

// List returns tasks matching the given filter.
func List(tasks []Task, f Filter) []Task {
	if f == FilterAll {
		return tasks
	}

	var out []Task
	for _, t := range tasks {
		switch {
		case f == FilterDone && t.Done:
			out = append(out, t)
		case f == FilterPending && !t.Done:
			out = append(out, t)
		}
	}
	return out
}
