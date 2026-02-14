package task

import "time"

// Priority represents the urgency level of a task.
type Priority string

const (
	PriorityNone   Priority = ""
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// ValidPriority checks whether s is a recognized priority level.
// Accepts shorthand: h, m, l.
func ValidPriority(s string) (Priority, bool) {
	switch s {
	case "":
		return PriorityNone, true
	case "low", "l":
		return PriorityLow, true
	case "medium", "m":
		return PriorityMedium, true
	case "high", "h":
		return PriorityHigh, true
	default:
		return "", false
	}
}

// Task represents a single tracked item.
type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	Priority    Priority   `json:"priority,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
