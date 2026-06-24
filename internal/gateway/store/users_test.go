package store

import (
	"testing"
)

func TestCreateAndLoginUser(t *testing.T) {
	s := openTestDB(t)

	user, apiKey, err := s.CreateUser("test@test.dev", "secret123")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.Email != "test@test.dev" {
		t.Errorf("email: %s", user.Email)
	}
	if apiKey == "" {
		t.Error("expected api key")
	}
	if len(apiKey) < 20 {
		t.Errorf("key too short: %d", len(apiKey))
	}

	// Login
	u, _, err := s.LoginUser("test@test.dev", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if u.Email != "test@test.dev" {
		t.Error("wrong user")
	}

	// Wrong password
	_, _, err = s.LoginUser("test@test.dev", "wrong")
	if err == nil {
		t.Error("expected error for wrong password")
	}

	// Wrong email
	_, _, err = s.LoginUser("notfound@test.dev", "secret123")
	if err == nil {
		t.Error("expected error for wrong email")
	}
}

func TestDuplicateEmail(t *testing.T) {
	s := openTestDB(t)

	_, _, err := s.CreateUser("dup@test.dev", "pass1")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = s.CreateUser("dup@test.dev", "pass2")
	if err != ErrEmailTaken {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestVerifyAPIKey(t *testing.T) {
	s := openTestDB(t)

	_, apiKey, _ := s.CreateUser("key@test.dev", "pass")

	user, err := s.VerifyAPIKey(apiKey)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if user.Email != "key@test.dev" {
		t.Error("wrong user")
	}

	// Invalid key
	_, err = s.VerifyAPIKey("invalid-key")
	if err == nil {
		t.Error("expected error")
	}

	// Empty key
	_, err = s.VerifyAPIKey("")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestRotateAPIKey(t *testing.T) {
	s := openTestDB(t)

	user, oldKey, _ := s.CreateUser("rotate@test.dev", "pass")

	newKey, err := s.RotateAPIKey(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if newKey == oldKey {
		t.Error("keys should differ")
	}

	// Old key invalid
	_, err = s.VerifyAPIKey(oldKey)
	if err == nil {
		t.Error("old key should be invalid")
	}

	// New key valid
	u, err := s.VerifyAPIKey(newKey)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != user.ID {
		t.Error("wrong user from new key")
	}
}

func TestWorkspaceMembership(t *testing.T) {
	s := openTestDB(t)

	user1, _, _ := s.CreateUser("u1@test.dev", "pass")
	user2, _, _ := s.CreateUser("u2@test.dev", "pass")

	ws1, _ := s.CreateWorkspace("WS1", "ws1")
	ws2, _ := s.CreateWorkspace("WS2", "ws2")

	// Add members
	if err := s.AddMember(ws1.ID, user1.ID, "admin"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMember(ws2.ID, user2.ID, "member"); err != nil {
		t.Fatal(err)
	}

	// Check membership
	isMember, err := s.IsMember(ws1.ID, user1.ID)
	if err != nil || !isMember {
		t.Error("user1 should be member of ws1")
	}
	isMember, _ = s.IsMember(ws1.ID, user2.ID)
	if isMember {
		t.Error("user2 should NOT be member of ws1")
	}

	// List user workspaces
	list, err := s.ListUserWorkspaces(user1.ID)
	if err != nil || len(list) != 1 {
		t.Errorf("user1 should have 1 workspace, got %d", len(list))
	}
	if len(list) > 0 && list[0].Name != "WS1" {
		t.Error("wrong workspace")
	}

	// User2 list
	list, _ = s.ListUserWorkspaces(user2.ID)
	if len(list) != 1 {
		t.Errorf("user2 should have 1 workspace, got %d", len(list))
	}
}
