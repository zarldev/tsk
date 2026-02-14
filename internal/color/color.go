package color

import (
	"os"
)

const (
	reset         = "\033[0m"
	bold          = "\033[1m"
	dim           = "\033[2m"
	strikethrough = "\033[9m"
	cyan          = "\033[36m"
	green         = "\033[32m"
)

// Palette applies ANSI color codes to strings.
// When disabled, all methods return input unchanged.
type Palette struct {
	enabled bool
}

// New creates a Palette based on the config setting and environment.
// configEnabled is the color.enabled config value: "auto", "always", or "never".
func New(configEnabled string) Palette {
	return Palette{enabled: shouldColor(configEnabled)}
}

// shouldColor determines whether to emit ANSI codes.
func shouldColor(configEnabled string) bool {
	if configEnabled == "never" {
		return false
	}

	// respect https://no-color.org
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}

	if configEnabled == "always" {
		return true
	}

	// "auto": only color when stdout is a terminal
	return isTerminal(os.Stdout)
}

// isTerminal reports whether f is a terminal device.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// wrap surrounds s with the given ANSI code and a reset sequence.
func (p Palette) wrap(code, s string) string {
	if !p.enabled {
		return s
	}
	return code + s + reset
}

// Bold returns s in bold.
func (p Palette) Bold(s string) string {
	return p.wrap(bold, s)
}

// Dim returns s in dim/faint text.
func (p Palette) Dim(s string) string {
	return p.wrap(dim, s)
}

// Green returns s in green.
func (p Palette) Green(s string) string {
	return p.wrap(green, s)
}

// Cyan returns s in cyan.
func (p Palette) Cyan(s string) string {
	return p.wrap(cyan, s)
}

// BoldCyan returns s in bold cyan.
func (p Palette) BoldCyan(s string) string {
	return p.wrap(bold+cyan, s)
}

// DimStrikethrough returns s in dim with strikethrough.
func (p Palette) DimStrikethrough(s string) string {
	return p.wrap(dim+strikethrough, s)
}
