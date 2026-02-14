package task

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// compile-time check: GistStore implements Store
var _ Store = (*GistStore)(nil)

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestGistLoadEmptyID(t *testing.T) {
	store := NewGistStore("token", "")
	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected empty slice, got %d tasks", len(tasks))
	}
}

func TestGistLoadSuccess(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	want := []Task{
		{ID: 1, Title: "buy milk", Done: false, CreatedAt: now},
		{ID: 2, Title: "write code", Done: true, CreatedAt: now},
	}

	content, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/gists/abc123") {
			t.Errorf("path = %s, want suffix /gists/abc123", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("auth = %q, want %q", got, "Bearer test-token")
		}

		resp := gistResponse{
			ID: "abc123",
			Files: map[string]gistFileContent{
				gistFilename: {Content: string(content)},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("test-token", "abc123")
	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != len(want) {
		t.Fatalf("got %d tasks, want %d", len(tasks), len(want))
	}
	for i := range want {
		if tasks[i].ID != want[i].ID {
			t.Errorf("task %d: ID = %d, want %d", i, tasks[i].ID, want[i].ID)
		}
		if tasks[i].Title != want[i].Title {
			t.Errorf("task %d: Title = %q, want %q", i, tasks[i].Title, want[i].Title)
		}
		if tasks[i].Done != want[i].Done {
			t.Errorf("task %d: Done = %v, want %v", i, tasks[i].Done, want[i].Done)
		}
		if !tasks[i].CreatedAt.Equal(want[i].CreatedAt) {
			t.Errorf("task %d: CreatedAt = %v, want %v", i, tasks[i].CreatedAt, want[i].CreatedAt)
		}
	}
}

func TestGistLoadMissingFile(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		resp := gistResponse{
			ID:    "abc123",
			Files: map[string]gistFileContent{},
		}
		json.NewEncoder(w).Encode(resp)
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "abc123")
	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected empty slice, got %d tasks", len(tasks))
	}
}

func TestGistSaveUpdate(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "test", Done: false, CreatedAt: time.Now()},
	}

	var receivedBody gistRequest
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/gists/existing-id") {
			t.Errorf("path = %s, want suffix /gists/existing-id", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gistResponse{ID: "existing-id"})
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "existing-id")
	if err := store.Save(tasks); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// verify the request body structure
	f, ok := receivedBody.Files[gistFilename]
	if !ok {
		t.Fatal("missing tasks.json in request body")
	}
	if f.Content == "" {
		t.Fatal("empty content in request body")
	}

	// verify the tasks round-trip through the content
	var decoded []Task
	if err := json.Unmarshal([]byte(f.Content), &decoded); err != nil {
		t.Fatalf("unmarshal request content: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("got %d tasks in body, want 1", len(decoded))
	}
	if decoded[0].ID != 1 || decoded[0].Title != "test" {
		t.Errorf("task = %+v, want ID=1 Title=test", decoded[0])
	}
}

func TestGistSaveCreateNew(t *testing.T) {
	var receivedMethod string
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method

		var body gistRequest
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &body)

		if body.Public {
			t.Error("gist should be private")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(gistResponse{ID: "new-gist-id"})
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "")
	if err := store.Save([]Task{{ID: 1, Title: "first"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", receivedMethod)
	}

	// gist ID should be captured in memory
	if store.GistID != "new-gist-id" {
		t.Errorf("GistID = %q, want %q", store.GistID, "new-gist-id")
	}
}

func TestGistSaveCreateThenUpdate(t *testing.T) {
	callCount := 0
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			// first call: create
			if r.Method != http.MethodPost {
				t.Errorf("call 1: method = %s, want POST", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(gistResponse{ID: "created-id"})
		case 2:
			// second call: update
			if r.Method != http.MethodPatch {
				t.Errorf("call 2: method = %s, want PATCH", r.Method)
			}
			if !strings.HasSuffix(r.URL.Path, "/gists/created-id") {
				t.Errorf("call 2: path = %s, want suffix /gists/created-id", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gistResponse{ID: "created-id"})
		}
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "")

	// first save creates
	if err := store.Save([]Task{{ID: 1, Title: "first"}}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if store.GistID != "created-id" {
		t.Fatalf("GistID = %q after create, want %q", store.GistID, "created-id")
	}

	// second save updates
	if err := store.Save([]Task{{ID: 1, Title: "first"}, {ID: 2, Title: "second"}}); err != nil {
		t.Fatalf("second save: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestGistErrorUnauthorized(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("bad-token", "abc123")
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("error = %q, want mention of authentication failed", err.Error())
	}
}

func TestGistErrorNotFound(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "nonexistent")
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want mention of not found", err.Error())
	}
}

func TestGistErrorRateLimited(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "abc123")
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for 429")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("error = %q, want mention of rate limited", err.Error())
	}
}

func TestGistErrorForbidden(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "abc123")
	err := store.Save([]Task{{ID: 1, Title: "test"}})
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("error = %q, want mention of rate limited", err.Error())
	}
}

func TestGistErrorServerError(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "abc123")
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Errorf("error = %q, want mention of unexpected status 500", err.Error())
	}
}

func TestGistSaveRequestBodyStructure(t *testing.T) {
	var receivedBody []byte
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gistResponse{ID: "abc123"})
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("token", "abc123")
	tasks := []Task{
		{ID: 1, Title: "buy milk", Done: false, CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)},
	}
	if err := store.Save(tasks); err != nil {
		t.Fatal(err)
	}

	// verify the outer structure
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(receivedBody, &raw); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if _, ok := raw["files"]; !ok {
		t.Fatal("missing 'files' key in request body")
	}
	if _, ok := raw["public"]; !ok {
		t.Fatal("missing 'public' key in request body")
	}

	// verify public is false
	var req gistRequest
	json.Unmarshal(receivedBody, &req)
	if req.Public {
		t.Error("gist should be private (public=false)")
	}

	// verify file content is valid JSON tasks
	f := req.Files[gistFilename]
	var decoded []Task
	if err := json.Unmarshal([]byte(f.Content), &decoded); err != nil {
		t.Fatalf("content is not valid task JSON: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Title != "buy milk" {
		t.Errorf("decoded = %+v, want 1 task with title 'buy milk'", decoded)
	}
}

func TestGistAuthHeader(t *testing.T) {
	var authHeader string
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gistResponse{
			ID: "abc123",
			Files: map[string]gistFileContent{
				gistFilename: {Content: "[]"},
			},
		})
	})

	old := gistAPIBase
	gistAPIBase = srv.URL
	t.Cleanup(func() { gistAPIBase = old })

	store := NewGistStore("ghp_mytoken123", "abc123")
	store.Load()

	if authHeader != "Bearer ghp_mytoken123" {
		t.Errorf("Authorization = %q, want %q", authHeader, "Bearer ghp_mytoken123")
	}
}
