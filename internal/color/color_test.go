package color

import (
	"os"
	"testing"
)

func TestPaletteEnabled(t *testing.T) {
	p := Palette{enabled: true}

	tests := []struct {
		name string
		fn   func(string) string
		want string
	}{
		{"Bold", p.Bold, "\033[1mhello\033[0m"},
		{"Dim", p.Dim, "\033[2mhello\033[0m"},
		{"Green", p.Green, "\033[32mhello\033[0m"},
		{"Cyan", p.Cyan, "\033[36mhello\033[0m"},
		{"BoldCyan", p.BoldCyan, "\033[1m\033[36mhello\033[0m"},
		{"DimStrikethrough", p.DimStrikethrough, "\033[2m\033[9mhello\033[0m"},
		{"Red", p.Red, "\033[31mhello\033[0m"},
		{"Yellow", p.Yellow, "\033[33mhello\033[0m"},
		{"BoldRed", p.BoldRed, "\033[1m\033[31mhello\033[0m"},
		{"BoldYellow", p.BoldYellow, "\033[1m\033[33mhello\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn("hello")
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPaletteDisabled(t *testing.T) {
	p := Palette{enabled: false}

	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"Bold", p.Bold},
		{"Dim", p.Dim},
		{"Green", p.Green},
		{"Cyan", p.Cyan},
		{"BoldCyan", p.BoldCyan},
		{"DimStrikethrough", p.DimStrikethrough},
		{"Red", p.Red},
		{"Yellow", p.Yellow},
		{"BoldRed", p.BoldRed},
		{"BoldYellow", p.BoldYellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn("hello")
			if got != "hello" {
				t.Errorf("got %q, want plain %q", got, "hello")
			}
		})
	}
}

func TestPaletteEmptyString(t *testing.T) {
	p := Palette{enabled: true}
	got := p.Bold("")
	want := "\033[1m\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestShouldColorNever(t *testing.T) {
	if shouldColor("never") {
		t.Error("expected false for 'never'")
	}
}

func TestShouldColorAlways(t *testing.T) {
	// clear NO_COLOR if set
	prev, hadNoColor := os.LookupEnv("NO_COLOR")
	if hadNoColor {
		os.Unsetenv("NO_COLOR")
		defer os.Setenv("NO_COLOR", prev)
	}

	if !shouldColor("always") {
		t.Error("expected true for 'always' without NO_COLOR")
	}
}

func TestShouldColorNoColorEnv(t *testing.T) {
	prev, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", prev)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	if shouldColor("always") {
		t.Error("expected false when NO_COLOR is set, even with 'always'")
	}
	if shouldColor("auto") {
		t.Error("expected false when NO_COLOR is set, with 'auto'")
	}
}

func TestShouldColorNoColorEnvEmpty(t *testing.T) {
	// NO_COLOR spec: presence matters, value does not
	prev, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Setenv("NO_COLOR", "")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", prev)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	if shouldColor("always") {
		t.Error("expected false when NO_COLOR is set (even empty)")
	}
}

func TestNewPalette(t *testing.T) {
	prev, hadNoColor := os.LookupEnv("NO_COLOR")
	if hadNoColor {
		os.Unsetenv("NO_COLOR")
		defer os.Setenv("NO_COLOR", prev)
	}

	p := New("never")
	if p.enabled {
		t.Error("expected disabled for 'never'")
	}

	p = New("always")
	if !p.enabled {
		t.Error("expected enabled for 'always'")
	}
}

func TestIsTerminalPipe(t *testing.T) {
	// os.Pipe gives us file descriptors that are NOT terminals
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	if isTerminal(r) {
		t.Error("pipe read end should not be a terminal")
	}
	if isTerminal(w) {
		t.Error("pipe write end should not be a terminal")
	}
}
