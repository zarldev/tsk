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
func ValidPriority(s string) (Priority, bool) {
	switch Priority(s) {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh:
		return Priority(s), true
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
