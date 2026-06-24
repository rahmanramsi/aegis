package harness

import (
	"context"
	"os/exec"
	"sync"
)

type OpenCodeRunner struct {
	path  string
	model string
}

func NewOpenCodeRunner(path, model string) *OpenCodeRunner {
	return &OpenCodeRunner{path: path, model: model}
}

func (r *OpenCodeRunner) Name() string    { return "opencode" }
func (r *OpenCodeRunner) Available() bool { _, err := exec.LookPath("opencode"); return err == nil }

func (r *OpenCodeRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	path := r.path
	if path == "" {
		path = "opencode"
	}

	args := []string{"run", req.Prompt}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	} else if r.model != "" {
		args = append(args, "--model", r.model)
	}
	args = append(args, req.ExtraArgs...)

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = req.WorkDir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		ch <- StreamEvent{Type: EventError, Content: "opencode: " + err.Error()}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
		return ch, nil
	}

	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); scanStream(stdout, EventStdout, ch) }()
		go func() { defer wg.Done(); scanStream(stderr, EventStderr, ch) }()
		wg.Wait()
		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: "opencode: " + err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
	}()

	return ch, nil
}
