package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// gistAPIBase is the GitHub API base URL. Overridden in tests.
var gistAPIBase = "https://api.github.com"

// GistStore persists tasks as JSON in a private GitHub Gist.
type GistStore struct {
	Token  string // GitHub personal access token
	GistID string // existing gist ID (empty = create new on first save)
	client *http.Client
}

// NewGistStore returns a GistStore that syncs tasks via the GitHub Gist API.
func NewGistStore(token, gistID string) *GistStore {
	return &GistStore{
		Token:  token,
		GistID: gistID,
		client: http.DefaultClient,
	}
}

const gistFilename = "tasks.json"

// gistRequest is the JSON body for create/update gist API calls.
type gistRequest struct {
	Files  map[string]gistFile `json:"files"`
	Public bool                `json:"public"`
}

// gistFile represents a single file in a gist.
type gistFile struct {
	Content string `json:"content"`
}

// gistResponse is the relevant subset of the Gist API response.
type gistResponse struct {
	ID    string                     `json:"id"`
	Files map[string]gistFileContent `json:"files"`
}

// gistFileContent is the file content returned by the Gist API.
type gistFileContent struct {
	Content string `json:"content"`
}

// Load reads tasks from the gist. Returns an empty slice if no gist ID is set.
func (s *GistStore) Load() ([]Task, error) {
	if s.GistID == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%s/gists/%s", gistAPIBase, s.GistID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("gist: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gist: network error: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var gist gistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, fmt.Errorf("gist: decode response: %w", err)
	}

	f, ok := gist.Files[gistFilename]
	if !ok || f.Content == "" {
		return nil, nil
	}

	var tasks []Task
	if err := json.Unmarshal([]byte(f.Content), &tasks); err != nil {
		return nil, fmt.Errorf("gist: unmarshal tasks: %w", err)
	}
	return tasks, nil
}

// Save writes tasks to the gist. Creates a new private gist if GistID is empty.
func (s *GistStore) Save(tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("gist: marshal tasks: %w", err)
	}

	body := gistRequest{
		Files: map[string]gistFile{
			gistFilename: {Content: string(data)},
		},
		Public: false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("gist: marshal request: %w", err)
	}

	var method, url string
	if s.GistID == "" {
		method = http.MethodPost
		url = fmt.Sprintf("%s/gists", gistAPIBase)
	} else {
		method = http.MethodPatch
		url = fmt.Sprintf("%s/gists/%s", gistAPIBase, s.GistID)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("gist: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("gist: network error: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}

	// on create, capture the new gist ID
	if s.GistID == "" {
		var gist gistResponse
		if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
			return fmt.Errorf("gist: decode response: %w", err)
		}
		s.GistID = gist.ID
		fmt.Fprintf(os.Stderr, "created gist: %s — add to config to persist\n", s.GistID)
	}

	return nil
}

// checkResponse maps HTTP error statuses to clear error messages.
func checkResponse(resp *http.Response) error {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusUnauthorized:
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("gist: authentication failed (check gist_token or TSK_GIST_TOKEN)")
	case resp.StatusCode == http.StatusNotFound:
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("gist: not found (check gist_id)")
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests:
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("gist: rate limited, try again later")
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gist: unexpected status %d: %s", resp.StatusCode, string(body))
	}
}
