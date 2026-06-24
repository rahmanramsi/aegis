package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type WorkspaceManager struct {
	Root string
}

func NewWorkspaceManager(root string) *WorkspaceManager {
	return &WorkspaceManager{Root: root}
}

func (wm *WorkspaceManager) Create(workspaceID, agentID, taskID string) (string, error) {
	dir := filepath.Join(wm.Root, workspaceID, agentID, taskID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	return dir, nil
}

func (wm *WorkspaceManager) Cleanup(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	return filepath.Walk(wm.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.ModTime().Before(cutoff) && path != wm.Root {
			os.RemoveAll(path)
		}
		return nil
	})
}
