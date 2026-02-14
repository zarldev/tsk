package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all tsk configuration.
type Config struct {
	Color   ColorConfig
	Storage StorageConfig
}

// ColorConfig controls colored output behavior.
type ColorConfig struct {
	Enabled string // "auto", "always", "never"
}

// StorageConfig controls task storage.
type StorageConfig struct {
	Type      string // "file", "gist"
	Path      string // file path for "file" type
	GistToken string // GitHub PAT with gist scope
	GistID    string // gist ID (created on first save if empty)
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return Config{
		Color: ColorConfig{
			Enabled: "auto",
		},
		Storage: StorageConfig{
			Type: "file",
			Path: filepath.Join(home, ".tasks.json"),
		},
	}
}

// Path returns the config file path (~/.config/tsk/config.toml).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "tsk", "config.toml"), nil
}

// Load reads config from the standard config file path.
// Missing file or missing values fall back to defaults.
func Load() (Config, error) {
	cfg := DefaultConfig()

	p, err := Path()
	if err != nil {
		return cfg, nil
	}

	return LoadFrom(p)
}

// LoadFrom reads config from the given path.
// Missing file or missing values fall back to defaults.
func LoadFrom(path string) (Config, error) {
	cfg := DefaultConfig()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	sections, err := parseTOML(f)
	if err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if color, ok := sections["color"]; ok {
		if v, ok := color["enabled"]; ok {
			cfg.Color.Enabled = v
		}
	}

	if storage, ok := sections["storage"]; ok {
		if v, ok := storage["type"]; ok {
			cfg.Storage.Type = v
		}
		if v, ok := storage["path"]; ok {
			cfg.Storage.Path = expandHome(v)
		}
		if v, ok := storage["gist_token"]; ok {
			cfg.Storage.GistToken = v
		}
		if v, ok := storage["gist_id"]; ok {
			cfg.Storage.GistID = v
		}
	}

	return cfg, nil
}

// String returns the config in TOML format.
func (c Config) String() string {
	var b strings.Builder
	b.WriteString("# tsk configuration\n\n")
	b.WriteString("[color]\n")
	fmt.Fprintf(&b, "enabled = %q\n", c.Color.Enabled)
	b.WriteString("\n[storage]\n")
	fmt.Fprintf(&b, "type = %q\n", c.Storage.Type)
	fmt.Fprintf(&b, "path = %q\n", c.Storage.Path)
	fmt.Fprintf(&b, "gist_token = %q\n", c.Storage.GistToken)
	fmt.Fprintf(&b, "gist_id = %q\n", c.Storage.GistID)
	return b.String()
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

// parseTOML is a minimal TOML parser supporting top-level sections
// with string, boolean, and integer values. Comments start with #.
func parseTOML(f *os.File) (map[string]map[string]string, error) {
	sections := make(map[string]map[string]string)
	currentSection := ""

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// section header
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, fmt.Errorf("line %d: unclosed section header", lineNum)
			}
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			if _, ok := sections[currentSection]; !ok {
				sections[currentSection] = make(map[string]string)
			}
			continue
		}

		// key = value
		eqIdx := strings.IndexByte(line, '=')
		if eqIdx < 0 {
			return nil, fmt.Errorf("line %d: expected key = value", lineNum)
		}

		key := strings.TrimSpace(line[:eqIdx])
		valRaw := strings.TrimSpace(line[eqIdx+1:])

		// strip inline comment (only outside quotes)
		val, err := parseValue(valRaw)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if currentSection == "" {
			// top-level keys go into an empty-string section
			if _, ok := sections[""]; !ok {
				sections[""] = make(map[string]string)
			}
			sections[""][key] = val
			continue
		}

		sections[currentSection][key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return sections, nil
}

// parseValue extracts a value from a TOML value string.
// Handles quoted strings, booleans, and integers.
func parseValue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	// quoted string
	if raw[0] == '"' {
		// find closing quote, respecting that inline comments may follow
		endQuote := strings.IndexByte(raw[1:], '"')
		if endQuote < 0 {
			return "", fmt.Errorf("unclosed quote")
		}
		return raw[1 : endQuote+1], nil
	}

	// unquoted value: strip inline comment
	if idx := strings.IndexByte(raw, '#'); idx >= 0 {
		raw = strings.TrimSpace(raw[:idx])
	}

	// boolean
	if raw == "true" || raw == "false" {
		return raw, nil
	}

	// integer
	if _, err := strconv.Atoi(raw); err == nil {
		return raw, nil
	}

	// unquoted string
	return raw, nil
}
