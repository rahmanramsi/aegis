package harness

import (
	"context"
	"os/exec"
)

type GeminiRunner struct {
	path  string
	model string
}

func NewGeminiRunner(path, model string) *GeminiRunner {
	return &GeminiRunner{path: path, model: model}
}

func (r *GeminiRunner) Name() string    { return "gemini" }
func (r *GeminiRunner) Available() bool { _, err := exec.LookPath("gemini"); return err == nil }

func (r *GeminiRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	path := r.path
	if path == "" {
		path = "gemini"
	}

	args := []string{"chat", req.Prompt}
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
		ch <- StreamEvent{Type: EventError, Content: "gemini: " + err.Error()}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
		return ch, nil
	}

	go func() {
		defer close(ch)
		go scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)
		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: "gemini: " + err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}
