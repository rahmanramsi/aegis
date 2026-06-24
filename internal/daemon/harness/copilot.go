package harness

import (
	"context"
	"os/exec"
)

type CopilotRunner struct {
	path string
}

func NewCopilotRunner(path string) *CopilotRunner { return &CopilotRunner{path: path} }

func (r *CopilotRunner) Name() string    { return "copilot" }
func (r *CopilotRunner) Available() bool { _, err := exec.LookPath("copilot"); return err == nil }

func (r *CopilotRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	path := r.path
	if path == "" {
		path = "copilot"
	}

	args := []string{"suggest", req.Prompt}
	args = append(args, req.ExtraArgs...)

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = req.WorkDir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		ch <- StreamEvent{Type: EventError, Content: "copilot: " + err.Error()}
		ch <- StreamEvent{Type: EventDone}
		close(ch)
		return ch, nil
	}

	go func() {
		defer close(ch)
		go scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)
		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: "copilot: " + err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}
