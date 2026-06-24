package harness

import (
	"os"
	"os/exec"
	"strings"
)

type RegisteredHarness struct {
	Runner Runner
	Path   string
}

func Discover(runners []Runner) []RegisteredHarness {
	var out []RegisteredHarness
	for _, r := range runners {
		path := resolvePath(r.Name())
		if path == "" && !r.Available() {
			continue
		}
		if path == "" {
			path = r.Name()
		}
		out = append(out, RegisteredHarness{Runner: r, Path: path})
	}
	return out
}

func resolvePath(name string) string {
	envKey := "AEGIS_" + strings.ToUpper(name) + "_PATH"
	if p := os.Getenv(envKey); p != "" {
		return p
	}
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}
