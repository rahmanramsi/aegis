package store

import "testing"

func TestCreateAndGetAgent(t *testing.T) {
	s := openTestDB(t)

	ws, _ := s.CreateWorkspace("W", "w")
	dm, _ := s.CreateDaemon(ws.ID, "d", "hash123")

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

	ws, _ := s.CreateWorkspace("W", "w")
	dm, _ := s.CreateDaemon(ws.ID, "d", "h")
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

	ws, _ := s.CreateWorkspace("W", "w")
	dm, _ := s.CreateDaemon(ws.ID, "d", "h")
	s.AuthenticateDaemon(dm.ID, "h", []string{"echo", "claude"})

	s.CreateAgent(ws.ID, dm.ID, "A", "echo", "", "", "", "", true)
	s.CreateAgent(ws.ID, dm.ID, "B", "claude", "", "", "", "", false)

	list, err := s.ListAgents(ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(list))
	}
	names := map[string]bool{}
	for _, a := range list {
		names[a.Name] = true
	}
	if !names["A"] || !names["B"] {
		t.Errorf("missing agents, got: %v", names)
	}
	if list[1].Enabled {
		t.Error("agent B should be disabled")
	}
}

func TestUpdateAgent(t *testing.T) {
	s := openTestDB(t)

	ws, _ := s.CreateWorkspace("W", "w")
	dm, _ := s.CreateDaemon(ws.ID, "d", "h")
	s.AuthenticateDaemon(dm.ID, "h", []string{"echo"})

	a, _ := s.CreateAgent(ws.ID, dm.ID, "Bot", "echo", "", "", "", "", true)

	updated, err := s.UpdateAgent(a.ID, "NewBot", "echo", "opus", "", "New personality", false)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "NewBot" || updated.Model != "opus" {
		t.Errorf("got %s/%s", updated.Name, updated.Model)
	}
	if updated.Personality != "New personality" {
		t.Error("personality not updated")
	}
	if updated.Enabled {
		t.Error("should be disabled")
	}
}

func TestDeleteAgent(t *testing.T) {
	s := openTestDB(t)

	ws, _ := s.CreateWorkspace("W", "w")
	dm, _ := s.CreateDaemon(ws.ID, "d", "h")
	s.AuthenticateDaemon(dm.ID, "h", []string{"echo"})

	a, _ := s.CreateAgent(ws.ID, dm.ID, "Bot", "echo", "", "", "", "", true)

	if err := s.DeleteAgent(a.ID); err != nil {
		t.Fatal(err)
	}
	_, err := s.GetAgent(a.ID)
	if err == nil {
		t.Error("expected not found")
	}
}
