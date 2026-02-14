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
	completed := now.Add(-30 * time.Minute)

	original := []Task{
		{ID: 1, Title: "buy milk", Done: false, CreatedAt: now},
		{ID: 2, Title: "write code", Done: true, CreatedAt: now, CompletedAt: &completed},
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

	// pending task: CompletedAt should be nil
	if loaded[0].CompletedAt != nil {
		t.Error("task 1: CompletedAt should be nil for pending task")
	}

	// done task: CompletedAt should round-trip
	if loaded[1].CompletedAt == nil {
		t.Fatal("task 2: CompletedAt should not be nil")
	}
	if !loaded[1].CompletedAt.Equal(completed) {
		t.Errorf("task 2: CompletedAt = %v, want %v", *loaded[1].CompletedAt, completed)
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
			result := Add(tt.existing, tt.title, PriorityNone)

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
			for i := range tt.tasks {
				if tt.tasks[i].ID == tt.id {
					if !tt.tasks[i].Done {
						t.Errorf("task %d should be done", tt.id)
					}
					if tt.tasks[i].CompletedAt == nil {
						t.Errorf("task %d: CompletedAt should be set", tt.id)
					}
				}
			}
		})
	}
}

func TestEdit(t *testing.T) {
	now := time.Now()
	completed := now.Add(-time.Hour)

	tests := []struct {
		name      string
		tasks     []Task
		id        int
		title     string
		wantErr   bool
		wantTitle string
	}{
		{
			name: "rename existing task",
			tasks: []Task{
				{ID: 1, Title: "buy milk", Done: false, CreatedAt: now},
				{ID: 2, Title: "write code", Done: true, CreatedAt: now, CompletedAt: &completed},
			},
			id:        1,
			title:     "buy oat milk",
			wantErr:   false,
			wantTitle: "buy oat milk",
		},
		{
			name: "preserves other fields",
			tasks: []Task{
				{ID: 1, Title: "old title", Done: true, CreatedAt: now, CompletedAt: &completed},
			},
			id:        1,
			title:     "new title",
			wantErr:   false,
			wantTitle: "new title",
		},
		{
			name: "not found",
			tasks: []Task{
				{ID: 1, Title: "a"},
			},
			id:      99,
			title:   "whatever",
			wantErr: true,
		},
		{
			name:    "empty list",
			tasks:   nil,
			id:      1,
			title:   "whatever",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Edit(tt.tasks, tt.id, tt.title)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// find the edited task and verify
			for _, tk := range tt.tasks {
				if tk.ID == tt.id {
					if tk.Title != tt.wantTitle {
						t.Errorf("Title = %q, want %q", tk.Title, tt.wantTitle)
					}
				}
			}
		})
	}

	// verify other fields are preserved after edit
	tasks := []Task{
		{ID: 1, Title: "old", Done: true, CreatedAt: now, CompletedAt: &completed},
	}
	if err := Edit(tasks, 1, "new"); err != nil {
		t.Fatalf("edit: %v", err)
	}
	if tasks[0].Done != true {
		t.Error("Done should be preserved")
	}
	if !tasks[0].CreatedAt.Equal(now) {
		t.Error("CreatedAt should be preserved")
	}
	if tasks[0].CompletedAt == nil || !tasks[0].CompletedAt.Equal(completed) {
		t.Error("CompletedAt should be preserved")
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

func TestAddCompletedAtNil(t *testing.T) {
	tasks := Add(nil, "new task", PriorityNone)
	if tasks[0].CompletedAt != nil {
		t.Error("new task should have nil CompletedAt")
	}
}

func TestDoneCompletedAtRoundTrip(t *testing.T) {
	store := tempStore(t)
	now := time.Now().Truncate(time.Second)
	completed := now.Add(-time.Hour)

	original := []Task{
		{ID: 1, Title: "pending", Done: false, CreatedAt: now},
		{ID: 2, Title: "done", Done: true, CreatedAt: now, CompletedAt: &completed},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// pending task: CompletedAt should remain nil
	if loaded[0].CompletedAt != nil {
		t.Error("task 1: CompletedAt should be nil")
	}

	// done task: CompletedAt should round-trip
	if loaded[1].CompletedAt == nil {
		t.Fatal("task 2: CompletedAt should not be nil")
	}
	if !loaded[1].CompletedAt.Equal(completed) {
		t.Errorf("task 2: CompletedAt = %v, want %v", *loaded[1].CompletedAt, completed)
	}
}

func TestClearDone(t *testing.T) {
	tests := []struct {
		name          string
		tasks         []Task
		wantRemoved   int
		wantRemaining int
	}{
		{
			name: "mix of done and pending",
			tasks: []Task{
				{ID: 1, Title: "done1", Done: true},
				{ID: 2, Title: "pending1", Done: false},
				{ID: 3, Title: "done2", Done: true},
				{ID: 4, Title: "pending2", Done: false},
			},
			wantRemoved:   2,
			wantRemaining: 2,
		},
		{
			name: "no done tasks",
			tasks: []Task{
				{ID: 1, Title: "pending1", Done: false},
				{ID: 2, Title: "pending2", Done: false},
			},
			wantRemoved:   0,
			wantRemaining: 2,
		},
		{
			name: "all done",
			tasks: []Task{
				{ID: 1, Title: "done1", Done: true},
				{ID: 2, Title: "done2", Done: true},
				{ID: 3, Title: "done3", Done: true},
			},
			wantRemoved:   3,
			wantRemaining: 0,
		},
		{
			name:          "empty slice",
			tasks:         nil,
			wantRemoved:   0,
			wantRemaining: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed, remaining := ClearDone(tt.tasks)
			if removed != tt.wantRemoved {
				t.Errorf("removed = %d, want %d", removed, tt.wantRemoved)
			}
			if len(remaining) != tt.wantRemaining {
				t.Errorf("remaining len = %d, want %d", len(remaining), tt.wantRemaining)
			}
			for _, tk := range remaining {
				if tk.Done {
					t.Errorf("task %d should not be done in remaining", tk.ID)
				}
			}
		})
	}
}

func TestFind(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "first"},
		{ID: 3, Title: "third"},
		{ID: 5, Title: "fifth"},
	}

	tests := []struct {
		name      string
		id        int
		wantTitle string
		wantNil   bool
	}{
		{"existing first", 1, "first", false},
		{"existing middle", 3, "third", false},
		{"existing last", 5, "fifth", false},
		{"not found", 99, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Find(tasks, tt.id)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil task")
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
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
	tasks = Add(tasks, "buy milk", PriorityNone)
	tasks = Add(tasks, "write tests", PriorityHigh)
	tasks = Add(tasks, "deploy", PriorityNone)

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
	if tasks[1].CompletedAt == nil {
		t.Error("task 2: CompletedAt should be set after done")
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

func TestValidPriority(t *testing.T) {
	tests := []struct {
		input string
		want  Priority
		ok    bool
	}{
		{"", PriorityNone, true},
		{"low", PriorityLow, true},
		{"medium", PriorityMedium, true},
		{"high", PriorityHigh, true},
		{"critical", "", false},
		{"urgent", "", false},
		{"HIGH", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ValidPriority(tt.input)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("priority = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddWithPriority(t *testing.T) {
	tasks := Add(nil, "urgent thing", PriorityHigh)
	if len(tasks) != 1 {
		t.Fatalf("len = %d, want 1", len(tasks))
	}
	if tasks[0].Priority != PriorityHigh {
		t.Errorf("priority = %q, want %q", tasks[0].Priority, PriorityHigh)
	}
}

func TestAddWithoutPriority(t *testing.T) {
	tasks := Add(nil, "normal thing", PriorityNone)
	if len(tasks) != 1 {
		t.Fatalf("len = %d, want 1", len(tasks))
	}
	if tasks[0].Priority != PriorityNone {
		t.Errorf("priority = %q, want empty", tasks[0].Priority)
	}
}

func TestPriorityJSONRoundTrip(t *testing.T) {
	store := tempStore(t)
	now := time.Now().Truncate(time.Second)

	original := []Task{
		{ID: 1, Title: "high task", Priority: PriorityHigh, CreatedAt: now},
		{ID: 2, Title: "no priority", Priority: PriorityNone, CreatedAt: now},
		{ID: 3, Title: "low task", Priority: PriorityLow, CreatedAt: now},
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
		if loaded[i].Priority != original[i].Priority {
			t.Errorf("task %d: priority = %q, want %q", i, loaded[i].Priority, original[i].Priority)
		}
	}
}

func TestBackwardsCompatibility(t *testing.T) {
	store := tempStore(t)

	// simulate old-format JSON without priority field
	data := []byte(`[{"id":1,"title":"old task","done":false,"created_at":"2025-01-15T10:00:00Z"}]`)
	if err := os.WriteFile(store.Path, data, 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Priority != PriorityNone {
		t.Errorf("priority = %q, want empty (backwards compatible)", tasks[0].Priority)
	}
}
