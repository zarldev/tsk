package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Color.Enabled != "auto" {
		t.Errorf("Color.Enabled = %q, want %q", cfg.Color.Enabled, "auto")
	}
	if cfg.Storage.Type != "file" {
		t.Errorf("Storage.Type = %q, want %q", cfg.Storage.Type, "file")
	}
	if !strings.HasSuffix(cfg.Storage.Path, ".tasks.json") {
		t.Errorf("Storage.Path = %q, want suffix %q", cfg.Storage.Path, ".tasks.json")
	}
}

func TestLoadNonExistent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nope.toml")
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// should return defaults
	def := DefaultConfig()
	if cfg.Color.Enabled != def.Color.Enabled {
		t.Errorf("Color.Enabled = %q, want default %q", cfg.Color.Enabled, def.Color.Enabled)
	}
	if cfg.Storage.Type != def.Storage.Type {
		t.Errorf("Storage.Type = %q, want default %q", cfg.Storage.Type, def.Storage.Type)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	p := writeConfig(t, "")
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def := DefaultConfig()
	if cfg.Color.Enabled != def.Color.Enabled {
		t.Errorf("Color.Enabled = %q, want default %q", cfg.Color.Enabled, def.Color.Enabled)
	}
}

func TestLoadFullConfig(t *testing.T) {
	content := `# tsk configuration

[color]
enabled = "always"

[storage]
type = "gist"
path = "/tmp/tasks.json"
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Color.Enabled != "always" {
		t.Errorf("Color.Enabled = %q, want %q", cfg.Color.Enabled, "always")
	}
	if cfg.Storage.Type != "gist" {
		t.Errorf("Storage.Type = %q, want %q", cfg.Storage.Type, "gist")
	}
	if cfg.Storage.Path != "/tmp/tasks.json" {
		t.Errorf("Storage.Path = %q, want %q", cfg.Storage.Path, "/tmp/tasks.json")
	}
}

func TestLoadPartialConfig(t *testing.T) {
	content := `[color]
enabled = "never"
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Color.Enabled != "never" {
		t.Errorf("Color.Enabled = %q, want %q", cfg.Color.Enabled, "never")
	}

	// storage should be defaults
	def := DefaultConfig()
	if cfg.Storage.Type != def.Storage.Type {
		t.Errorf("Storage.Type = %q, want default %q", cfg.Storage.Type, def.Storage.Type)
	}
	if cfg.Storage.Path != def.Storage.Path {
		t.Errorf("Storage.Path = %q, want default %q", cfg.Storage.Path, def.Storage.Path)
	}
}

func TestLoadCommentsOnly(t *testing.T) {
	content := `# just a comment
# another comment
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def := DefaultConfig()
	if cfg.Color.Enabled != def.Color.Enabled {
		t.Errorf("Color.Enabled = %q, want default %q", cfg.Color.Enabled, def.Color.Enabled)
	}
}

func TestLoadInlineComments(t *testing.T) {
	content := `[color]
enabled = "auto"  # auto-detect terminal

[storage]
type = "file"  # local file storage
path = "/tmp/test.json"  # custom path
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Color.Enabled != "auto" {
		t.Errorf("Color.Enabled = %q, want %q", cfg.Color.Enabled, "auto")
	}
	if cfg.Storage.Type != "file" {
		t.Errorf("Storage.Type = %q, want %q", cfg.Storage.Type, "file")
	}
	if cfg.Storage.Path != "/tmp/test.json" {
		t.Errorf("Storage.Path = %q, want %q", cfg.Storage.Path, "/tmp/test.json")
	}
}

func TestLoadBooleanValues(t *testing.T) {
	content := `[color]
enabled = true
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Color.Enabled != "true" {
		t.Errorf("Color.Enabled = %q, want %q", cfg.Color.Enabled, "true")
	}
}

func TestLoadUnquotedValues(t *testing.T) {
	content := `[storage]
type = file
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Storage.Type != "file" {
		t.Errorf("Storage.Type = %q, want %q", cfg.Storage.Type, "file")
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	content := `[storage]
path = "~/.tasks.json"
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(home, ".tasks.json")
	if cfg.Storage.Path != want {
		t.Errorf("Storage.Path = %q, want %q", cfg.Storage.Path, want)
	}
}

func TestExpandHomeTildeSlash(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	content := `[storage]
path = "~/custom/tasks.json"
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(home, "custom", "tasks.json")
	if cfg.Storage.Path != want {
		t.Errorf("Storage.Path = %q, want %q", cfg.Storage.Path, want)
	}
}

func TestNoExpandNonTilde(t *testing.T) {
	content := `[storage]
path = "/absolute/path/tasks.json"
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Storage.Path != "/absolute/path/tasks.json" {
		t.Errorf("Storage.Path = %q, want %q", cfg.Storage.Path, "/absolute/path/tasks.json")
	}
}

func TestParseErrorUnclosedSection(t *testing.T) {
	content := `[color
enabled = "auto"
`
	p := writeConfig(t, content)
	_, err := LoadFrom(p)
	if err == nil {
		t.Fatal("expected error for unclosed section header")
	}
	if !strings.Contains(err.Error(), "unclosed section") {
		t.Errorf("error = %q, want mention of unclosed section", err.Error())
	}
}

func TestParseErrorMissingEquals(t *testing.T) {
	content := `[color]
enabled "auto"
`
	p := writeConfig(t, content)
	_, err := LoadFrom(p)
	if err == nil {
		t.Fatal("expected error for missing equals")
	}
	if !strings.Contains(err.Error(), "key = value") {
		t.Errorf("error = %q, want mention of key = value", err.Error())
	}
}

func TestParseErrorUnclosedQuote(t *testing.T) {
	content := `[color]
enabled = "auto
`
	p := writeConfig(t, content)
	_, err := LoadFrom(p)
	if err == nil {
		t.Fatal("expected error for unclosed quote")
	}
	if !strings.Contains(err.Error(), "unclosed quote") {
		t.Errorf("error = %q, want mention of unclosed quote", err.Error())
	}
}

func TestConfigString(t *testing.T) {
	cfg := Config{
		Color: ColorConfig{
			Enabled: "always",
		},
		Storage: StorageConfig{
			Type: "file",
			Path: "/tmp/tasks.json",
		},
	}

	s := cfg.String()

	if !strings.Contains(s, "[color]") {
		t.Error("missing [color] section")
	}
	if !strings.Contains(s, `enabled = "always"`) {
		t.Error("missing enabled = always")
	}
	if !strings.Contains(s, "[storage]") {
		t.Error("missing [storage] section")
	}
	if !strings.Contains(s, `type = "file"`) {
		t.Error("missing type = file")
	}
	if !strings.Contains(s, `path = "/tmp/tasks.json"`) {
		t.Error("missing path value")
	}
}

func TestConfigStringRoundTrip(t *testing.T) {
	original := Config{
		Color: ColorConfig{
			Enabled: "never",
		},
		Storage: StorageConfig{
			Type: "gist",
			Path: "/custom/path.json",
		},
	}

	// write String() output to a file and reload
	p := writeConfig(t, original.String())
	loaded, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Color.Enabled != original.Color.Enabled {
		t.Errorf("Color.Enabled = %q, want %q", loaded.Color.Enabled, original.Color.Enabled)
	}
	if loaded.Storage.Type != original.Storage.Type {
		t.Errorf("Storage.Type = %q, want %q", loaded.Storage.Type, original.Storage.Type)
	}
	if loaded.Storage.Path != original.Storage.Path {
		t.Errorf("Storage.Path = %q, want %q", loaded.Storage.Path, original.Storage.Path)
	}
}

func TestTopLevelKeysIgnored(t *testing.T) {
	content := `foo = "bar"

[color]
enabled = "always"
`
	p := writeConfig(t, content)
	cfg, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// top-level keys are parsed but not mapped to anything; no crash
	if cfg.Color.Enabled != "always" {
		t.Errorf("Color.Enabled = %q, want %q", cfg.Color.Enabled, "always")
	}
}

func TestPath(t *testing.T) {
	p, err := Path()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(p, filepath.Join(".config", "tsk", "config.toml")) {
		t.Errorf("Path() = %q, want suffix .config/tsk/config.toml", p)
	}
}
