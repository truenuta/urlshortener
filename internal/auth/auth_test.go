package auth

import (
	"testing"
)

func TestBuildAndParseToken(t *testing.T) {
	userID := "abc"

	token, err := BuildJWTString(userID)
	if err != nil {
		t.Fatalf("BuildJWTString returned error: %v", err)
	}
	if token == "" {
		t.Fatal("BuildJWTString returned empty token")
	}
	id, err := GetUserID(token)
	if err != nil {
		t.Fatalf("GetUserID returned error: %v", err)
	}
	if id != userID {
		t.Errorf("GetUserID = %q, want %q", id, userID)
	}
}

func TestGetUserID_InvalidToken(t *testing.T) {
	id, err := GetUserID("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if id != "" {
		t.Errorf("expected empty id for invalid token, got %q", id)
	}
}
