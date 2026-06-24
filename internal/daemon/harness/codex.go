package harness

import (
	"context"
	"os/exec"
)

type CodexRunner struct {
	path  string
	model string
}

func NewCodexRunner(path, model string) *CodexRunner { return &CodexRunner{path: path, model: model} }

func (r *CodexRunner) Name() string    { return "codex" }
func (r *CodexRunner) Available() bool { _, err := exec.LookPath("codex"); return err == nil }

func (r *CodexRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	path := r.path
	if path == "" {
		path = "codex"
	}

	args := []string{"exec", req.Prompt}
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
		ch <- StreamEvent{Type: EventError, Content: "codex: " + err.Error()}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
		return ch, nil
	}

	go func() {
		defer close(ch)
		go scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)
		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: "codex: " + err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}
