package harness

import "context"

type EventType int

const (
	EventStdout EventType = iota
	EventStderr
	EventDone
	EventError
)

type StreamEvent struct {
	Type    EventType
	Content string
}

type RunRequest struct {
	TaskID    string
	Prompt    string
	WorkDir   string
	Model     string
	SessionID string
	ExtraArgs []string
}

type Runner interface {
	Name() string
	Available() bool
	Models(ctx context.Context) ([]string, error)
	Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error)
}
