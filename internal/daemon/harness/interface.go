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
	ExtraArgs []string
}

type Runner interface {
	Name() string
	Available() bool
	Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error)
}
