package harness

import (
	"context"
	"os/exec"
	"sync"
)

type ClaudeRunner struct {
	path  string
	model string
}

func NewClaudeRunner(path, model string) *ClaudeRunner { return &ClaudeRunner{path: path, model: model} }

func (r *ClaudeRunner) Name() string    { return "claude" }
func (r *ClaudeRunner) Available() bool { _, err := exec.LookPath("claude"); return err == nil }
func (r *ClaudeRunner) Models(_ context.Context) ([]string, error) { return nil, nil }

func (r *ClaudeRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	path := r.path
	if path == "" {
		path = "claude"
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
		ch <- StreamEvent{Type: EventError, Content: "claude: " + err.Error()}
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
			ch <- StreamEvent{Type: EventError, Content: "claude: " + err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
	}()

	return ch, nil
}
