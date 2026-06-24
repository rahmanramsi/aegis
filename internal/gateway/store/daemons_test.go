package store

import "testing"

func TestCreateAndAuthDaemon(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@a.test", "pass")
	d, err := s.CreateDaemon(u.ID, "mac", "test-hash")
	if err != nil {
		t.Fatalf("create daemon: %v", err)
	}
	if d.Name != "mac" {
		t.Errorf("name: %s", d.Name)
	}
	if d.Status != "offline" {
		t.Errorf("status: %s", d.Status)
	}

	// Authenticate with harnesses
	err = s.AuthenticateDaemon(d.ID, "test-hash", []string{"echo", "claude"})
	if err != nil {
		t.Fatalf("auth: %v", err)
	}

	// Check harnesses
	h, err := s.GetDaemonHarnesses(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 2 || h[0] != "claude" || h[1] != "echo" {
		t.Errorf("harnesses: %v", h)
	}

	// Wrong token
	err = s.AuthenticateDaemon(d.ID, "wrong-hash", nil)
	if err == nil {
		t.Error("expected auth error")
	}

	got, err := s.GetDaemon(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "online" {
		t.Errorf("status: %s", got.Status)
	}
}

func TestListDaemonsByUser(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@b.test", "pass")
	s.CreateDaemon(u.ID, "d1", "h1")
	s.CreateDaemon(u.ID, "d2", "h2")

	list, err := s.ListDaemonsByUser(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestSetDaemonOffline(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@c.test", "pass")
	d, _ := s.CreateDaemon(u.ID, "d", "h")
	s.AuthenticateDaemon(d.ID, "h", nil)

	s.SetDaemonOffline(d.ID)
	got, _ := s.GetDaemon(d.ID)
	if got.Status != "offline" {
		t.Errorf("status: %s", got.Status)
	}
}

func TestDeleteDaemon(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@d.test", "pass")
	d, _ := s.CreateDaemon(u.ID, "d", "h")

	s.AuthenticateDaemon(d.ID, "h", []string{"echo"})
	s.DeleteDaemon(d.ID)

	// Harnesses should be cascade-deleted
	h, _ := s.GetDaemonHarnesses(d.ID)
	if len(h) != 0 {
		t.Error("harnesses should be deleted")
	}
}
