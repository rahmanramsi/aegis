package harness

import (
	"context"
	"fmt"
	"strings"
)

type EchoRunner struct{}

func NewEchoRunner() *EchoRunner { return &EchoRunner{} }

func (e *EchoRunner) Name() string    { return "echo" }
func (e *EchoRunner) Available() bool { return true }

func (e *EchoRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 8)
	go func() {
		defer close(ch)
		words := strings.Split(req.Prompt, " ")
		ch <- StreamEvent{Type: EventStdout, Content: fmt.Sprintf("[echo] Processing '%s' with model '%s'", req.Prompt, req.Model)}
		ch <- StreamEvent{Type: EventStdout, Content: fmt.Sprintf("[echo] Word count: %d", len(words))}
		ch <- StreamEvent{Type: EventStdout, Content: "[echo] Done."}
		ch <- StreamEvent{Type: EventDone}
	}()
	return ch, nil
}
