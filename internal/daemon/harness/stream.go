package harness

import (
	"bufio"
	"fmt"
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

// parsePiModels parses `pi --list-models` table output into provider/model format.
// Format: "provider      model                       context  max-out  thinking  images"
func parsePiModels(out []byte) []string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil
	}
	// Skip header row
	var models []string
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			models = append(models, fmt.Sprintf("%s/%s", fields[0], fields[1]))
		}
	}
	return models
}
