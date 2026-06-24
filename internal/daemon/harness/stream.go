package harness

import (
	"bufio"
	"io"
	"strings"
)

func scanStream(r io.Reader, typ EventType, ch chan<- StreamEvent) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		ch <- StreamEvent{Type: typ, Content: scanner.Text()}
	}
}

// parseModels splits CLI output into non-empty lines.
func parseModels(out []byte) []string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var models []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			models = append(models, line)
		}
	}
	return models
}
