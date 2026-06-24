package harness

import (
	"context"
	"os/exec"
	"sync"
)

type PiRunner struct {
	path  string
	model string
}

func NewPiRunner(path, model string) *PiRunner { return &PiRunner{path: path, model: model} }

func (r *PiRunner) Name() string    { return "pi" }
func (r *PiRunner) Available() bool { _, err := exec.LookPath("pi"); return err == nil }
func (r *PiRunner) Models(ctx context.Context) ([]string, error) {
	path := r.path
	if path == "" {
		path = "pi"
	}
	cmd := exec.CommandContext(ctx, path, "models")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return parseModels(out), nil
	}
	return nil, nil // pi doesn't support model listing
}

func (r *PiRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	path := r.path
	if path == "" {
		path = "pi"
	}

	args := []string{"-p", req.Prompt}
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
		ch <- StreamEvent{Type: EventError, Content: "pi: " + err.Error()}
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
			ch <- StreamEvent{Type: EventError, Content: "pi: " + err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
	}()

	return ch, nil
}
