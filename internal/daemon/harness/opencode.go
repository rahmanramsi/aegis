package harness

import (
	"context"
	"os/exec"
	"strings"
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

func (r *OpenCodeRunner) Models(ctx context.Context) ([]string, error) {
	path := r.path
	if path == "" {
		path = "opencode"
	}
	cmd := exec.CommandContext(ctx, path, "models")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil // models listing not critical
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var models []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			models = append(models, line)
		}
	}
	return models, nil
}

func (r *OpenCodeRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	path := r.path
	if path == "" {
		path = "opencode"
	}

	args := []string{"run", req.Prompt}
	if req.SessionID != "" {
		args = append(args, "--session", req.SessionID)
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	} else if r.model != "" {
		args = append(args, "--model", r.model)
	}

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
