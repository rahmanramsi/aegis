package store

import (
	"testing"
)

func openTestDB(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(dir + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateAndGetWorkspace(t *testing.T) {
	s := openTestDB(t)

	ws, err := s.CreateWorkspace("Test", "test")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if ws.ID == "" {
		t.Error("expected id")
	}
	if ws.Name != "Test" || ws.Slug != "test" {
		t.Errorf("got name=%s slug=%s", ws.Name, ws.Slug)
	}

	got, err := s.GetWorkspace(ws.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Test" {
		t.Errorf("got %s", got.Name)
	}
}

func TestListWorkspaces(t *testing.T) {
	s := openTestDB(t)

	s.CreateWorkspace("A", "a")
	s.CreateWorkspace("B", "b")

	list, err := s.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("got %d workspaces", len(list))
	}
}

func TestUpdateWorkspace(t *testing.T) {
	s := openTestDB(t)

	ws, _ := s.CreateWorkspace("Old", "old")
	updated, err := s.UpdateWorkspace(ws.ID, "New", "new")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "New" || updated.Slug != "new" {
		t.Errorf("got %s/%s", updated.Name, updated.Slug)
	}
}

func TestDeleteWorkspace(t *testing.T) {
	s := openTestDB(t)

	ws, _ := s.CreateWorkspace("Del", "del")
	if err := s.DeleteWorkspace(ws.ID); err != nil {
		t.Fatal(err)
	}
	_, err := s.GetWorkspace(ws.ID)
	if err == nil {
		t.Error("expected not found")
	}
}
