package kratos

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSuccessResponseIncludesZeroCode(t *testing.T) {
	body, err := json.Marshal(Success(map[string]string{"token": "abc"}))
	if err != nil {
		t.Fatalf("marshal success response: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, `"code":0`) {
		t.Fatalf("expected success response to include code=0, got %s", text)
	}
	if !strings.Contains(text, `"message":"success"`) {
		t.Fatalf("expected success response to include message, got %s", text)
	}
}
