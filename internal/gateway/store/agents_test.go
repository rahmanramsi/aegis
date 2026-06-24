package store

import "testing"

func TestCreateAndGetAgent(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@a.test", "pass")
	ws, _ := s.CreateWorkspace("W", "w")
	s.AddMember(ws.ID, u.ID, "admin")
	dm, _ := s.CreateDaemon(u.ID, "d", "hash123")

	// Register harnesses
	s.AuthenticateDaemon(dm.ID, "hash123", []string{"echo", "claude"})

	a, err := s.CreateAgent(ws.ID, dm.ID, "Bot", "echo", "sonnet", "", "Be helpful.", "", true)
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if a.Name != "Bot" || a.Harness != "echo" {
		t.Errorf("got %s/%s", a.Name, a.Harness)
	}
	if a.Personality != "Be helpful." {
		t.Errorf("personality: %s", a.Personality)
	}
	if !a.Enabled {
		t.Error("should be enabled")
	}

	got, err := s.GetAgent(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Bot" {
		t.Error("wrong agent")
	}
}

func TestAgentWithTelegramToken(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@b.test", "pass")
	ws, _ := s.CreateWorkspace("W", "w")
	s.AddMember(ws.ID, u.ID, "admin")
	dm, _ := s.CreateDaemon(u.ID, "d", "h")
	s.AuthenticateDaemon(dm.ID, "h", []string{"echo"})

	// Create with token
	tokenHash := sha256Hex("bot-token-123")
	a, _ := s.CreateAgent(ws.ID, dm.ID, "Bot", "echo", "", "", "", tokenHash, true)

	if !a.HasTelegramToken {
		t.Error("should have telegram token")
	}
	if a.TelegramTokenHash != tokenHash {
		t.Error("token hash mismatch")
	}

	// Lookup by token hash
	found, err := s.GetAgentByTelegramToken(tokenHash)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != a.ID {
		t.Error("wrong agent from token lookup")
	}

	// Lookup with wrong hash
	_, err = s.GetAgentByTelegramToken("wrong")
	if err == nil {
		t.Error("expected error for wrong hash")
	}
}

func TestListAgentsByWorkspace(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@c.test", "pass")
	ws, _ := s.CreateWorkspace("W", "w")
	s.AddMember(ws.ID, u.ID, "admin")
	dm, _ := s.CreateDaemon(u.ID, "d", "h")
	s.AuthenticateDaemon(dm.ID, "h", []string{"echo"})

	s.CreateAgent(ws.ID, dm.ID, "a1", "echo", "", "", "", "", true)
	s.CreateAgent(ws.ID, dm.ID, "a2", "claude", "", "", "", "", true)

	list, err := s.ListAgents(ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestUpdateAgent(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@d.test", "pass")
	ws, _ := s.CreateWorkspace("W", "w")
	s.AddMember(ws.ID, u.ID, "admin")
	dm, _ := s.CreateDaemon(u.ID, "d", "h")
	s.AuthenticateDaemon(dm.ID, "h", []string{"echo"})

	a, _ := s.CreateAgent(ws.ID, dm.ID, "Bot", "echo", "", "", "", "", true)

	_, err := s.UpdateAgent(a.ID, "Bot", "echo", "", "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetAgent(a.ID)
	if got.Enabled {
		t.Error("should be disabled")
	}
}

func TestDeleteAgent(t *testing.T) {
	s := openTestDB(t)

	u, _, _ := s.CreateUser("test@e.test", "pass")
	ws, _ := s.CreateWorkspace("W", "w")
	s.AddMember(ws.ID, u.ID, "admin")
	dm, _ := s.CreateDaemon(u.ID, "d", "h")
	s.AuthenticateDaemon(dm.ID, "h", []string{"echo"})

	a, _ := s.CreateAgent(ws.ID, dm.ID, "Bot", "echo", "", "", "", "", true)
	s.DeleteAgent(a.ID)

	_, err := s.GetAgent(a.ID)
	if err == nil {
		t.Error("should be gone")
	}
}
