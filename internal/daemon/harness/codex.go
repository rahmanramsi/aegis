package harness

import (
	"context"
	"os/exec"
	"sync"
)

type CodexRunner struct {
	path  string
	model string
}

func NewCodexRunner(path, model string) *CodexRunner { return &CodexRunner{path: path, model: model} }

func (r *CodexRunner) Name() string    { return "codex" }
func (r *CodexRunner) Available() bool { _, err := exec.LookPath("codex"); return err == nil }
func (r *CodexRunner) Models(ctx context.Context) ([]string, error) {
	path := r.path
	if path == "" {
		path = "codex"
	}
	cmd := exec.CommandContext(ctx, path, "models")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return parseModels(out), nil
	}
	return []string{"gpt-5.1-codex", "gpt-5.1", "gpt-5", "gpt-4.1", "o3", "o4-mini"}, nil
}

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
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); scanStream(stdout, EventStdout, ch) }()
		go func() { defer wg.Done(); scanStream(stderr, EventStderr, ch) }()
		wg.Wait()
		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: "codex: " + err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
	}()

	return ch, nil
}
