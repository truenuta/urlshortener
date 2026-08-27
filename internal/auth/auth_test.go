package auth

import (
	"testing"
)

func TestBuildAndParseToken(t *testing.T) {
	token, _ := BuildJWTString("abc")
	t.Log("token", token)
	id, _ := GetUserID(token)
	t.Log("id", id)
}
