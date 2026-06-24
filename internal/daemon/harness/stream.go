package harness

import (
	"bufio"
	"io"
)

func scanStream(r io.Reader, typ EventType, ch chan<- StreamEvent) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		ch <- StreamEvent{Type: typ, Content: scanner.Text()}
	}
}
