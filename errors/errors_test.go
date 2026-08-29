package errors

import (
	"fmt"
	"testing"
)

func TestKitErrorWrapAndUnwrap(t *testing.T) {
	base := fmt.Errorf("db timeout")
	err := Wrap(base, 500, "db_error", "database request failed")
	if err == nil {
		t.Fatal("Wrap() returned nil")
	}
	if !IsKitError(err) {
		t.Fatal("IsKitError() should return true for wrapped KitError")
	}
	if got := err.Error(); got == "" || len(got) == 0 {
		t.Fatal("error string should not be empty")
	}
}

func TestNewKitError(t *testing.T) {
	err := New(400, "bad request", "request_invalid")
	if err == nil {
		t.Fatal("New() returned nil")
	}
	if got := err.GetCode(); got != 400 {
		t.Fatalf("GetCode() = %d, want 400", got)
	}
}
