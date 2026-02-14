package task

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// compile-time check: FileStore implements Store
var _ Store = (*FileStore)(nil)

func tempStore(t *testing.T) *FileStore {
	t.Helper()
	return NewFileStore(filepath.Join(t.TempDir(), "tasks.json"))
}

func TestLoadNonExistent(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "nope.json"))
	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected empty slice, got %d tasks", len(tasks))
	}
}

func TestLoadEmptyFile(t *testing.T) {
	store := tempStore(t)
	if err := os.WriteFile(store.Path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected empty slice, got %d tasks", len(tasks))
	}
}

func TestSaveAndLoad(t *testing.T) {
	store := tempStore(t)
	now := time.Now().Truncate(time.Second)

	original := []Task{
		{ID: 1, Title: "buy milk", Done: false, CreatedAt: now},
		{ID: 2, Title: "write code", Done: true, CreatedAt: now},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded) != len(original) {
		t.Fatalf("expected %d tasks, got %d", len(original), len(loaded))
	}

	for i := range original {
		if loaded[i].ID != original[i].ID {
			t.Errorf("task %d: ID = %d, want %d", i, loaded[i].ID, original[i].ID)
		}
		if loaded[i].Title != original[i].Title {
			t.Errorf("task %d: Title = %q, want %q", i, loaded[i].Title, original[i].Title)
		}
		if loaded[i].Done != original[i].Done {
			t.Errorf("task %d: Done = %v, want %v", i, loaded[i].Done, original[i].Done)
		}
		if !loaded[i].CreatedAt.Equal(original[i].CreatedAt) {
			t.Errorf("task %d: CreatedAt = %v, want %v", i, loaded[i].CreatedAt, original[i].CreatedAt)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		existing []Task
		title    string
		wantID   int
		wantLen  int
	}{
		{
			name:     "first task",
			existing: nil,
			title:    "buy milk",
			wantID:   1,
			wantLen:  1,
		},
		{
			name: "second task",
			existing: []Task{
				{ID: 1, Title: "first"},
			},
			title:   "second",
			wantID:  2,
			wantLen: 2,
		},
		{
			name: "after gap in IDs",
			existing: []Task{
				{ID: 1, Title: "first"},
				{ID: 5, Title: "fifth"},
			},
			title:   "sixth",
			wantID:  6,
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.existing, tt.title)

			if len(result) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(result), tt.wantLen)
			}

			last := result[len(result)-1]
			if last.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", last.ID, tt.wantID)
			}
			if last.Title != tt.title {
				t.Errorf("Title = %q, want %q", last.Title, tt.title)
			}
			if last.Done {
				t.Error("new task should not be done")
			}
			if last.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}
		})
	}
}

func TestDone(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []Task
		id      int
		wantErr bool
	}{
		{
			name: "mark existing",
			tasks: []Task{
				{ID: 1, Title: "a", Done: false},
				{ID: 2, Title: "b", Done: false},
			},
			id:      1,
			wantErr: false,
		},
		{
			name: "not found",
			tasks: []Task{
				{ID: 1, Title: "a"},
			},
			id:      99,
			wantErr: true,
		},
		{
			name:    "empty list",
			tasks:   nil,
			id:      1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Done(tt.tasks, tt.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			for _, tk := range tt.tasks {
				if tk.ID == tt.id && !tk.Done {
					t.Errorf("task %d should be done", tt.id)
				}
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []Task
		id      int
		wantLen int
		wantErr bool
	}{
		{
			name: "remove existing",
			tasks: []Task{
				{ID: 1, Title: "a"},
				{ID: 2, Title: "b"},
				{ID: 3, Title: "c"},
			},
			id:      2,
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "remove only task",
			tasks: []Task{
				{ID: 1, Title: "a"},
			},
			id:      1,
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "not found",
			tasks: []Task{
				{ID: 1, Title: "a"},
			},
			id:      99,
			wantLen: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Remove(tt.tasks, tt.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if len(result) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(result), tt.wantLen)
			}
			if err != nil {
				return
			}
			for _, tk := range result {
				if tk.ID == tt.id {
					t.Errorf("task %d should be removed", tt.id)
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "done task", Done: true},
		{ID: 2, Title: "pending task", Done: false},
		{ID: 3, Title: "another done", Done: true},
	}

	tests := []struct {
		name    string
		filter  Filter
		wantLen int
	}{
		{"all", FilterAll, 3},
		{"done", FilterDone, 2},
		{"pending", FilterPending, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := List(tasks, tt.filter)
			if len(result) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	store := tempStore(t)

	// start empty
	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("initial load: %v", err)
	}

	// add tasks
	tasks = Add(tasks, "buy milk")
	tasks = Add(tasks, "write tests")
	tasks = Add(tasks, "deploy")

	if err := store.Save(tasks); err != nil {
		t.Fatalf("save after add: %v", err)
	}

	// reload and mark done
	tasks, err = store.Load()
	if err != nil {
		t.Fatalf("load after add: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	if err := Done(tasks, 2); err != nil {
		t.Fatalf("done: %v", err)
	}
	if err := store.Save(tasks); err != nil {
		t.Fatalf("save after done: %v", err)
	}

	// reload and remove
	tasks, err = store.Load()
	if err != nil {
		t.Fatalf("load after done: %v", err)
	}
	if !tasks[1].Done {
		t.Error("task 2 should be done")
	}

	tasks, err = Remove(tasks, 1)
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := store.Save(tasks); err != nil {
		t.Fatalf("save after remove: %v", err)
	}

	// final reload
	tasks, err = store.Load()
	if err != nil {
		t.Fatalf("final load: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	// verify remaining IDs
	if tasks[0].ID != 2 || tasks[1].ID != 3 {
		t.Errorf("remaining IDs = [%d, %d], want [2, 3]", tasks[0].ID, tasks[1].ID)
	}
}
