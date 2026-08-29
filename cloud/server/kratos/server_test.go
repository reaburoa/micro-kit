package kratos

import (
	"testing"

	"github.com/reaburoa/micro-kit/cloud/server"
)

func TestNewHttp_HandlesNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("newHttp(nil) panicked: %v", r)
		}
	}()

	srv := newHttp(nil)
	if srv == nil {
		t.Fatal("newHttp(nil) returned nil server")
	}
}

func TestNewGrpc_HandlesNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("newGrpc(nil) panicked: %v", r)
		}
	}()

	srv := newGrpc(nil)
	if srv == nil {
		t.Fatal("newGrpc(nil) returned nil server")
	}
}

func TestNewHttp_UsesDefaultsWhenConfigIsEmpty(t *testing.T) {
	srv := newHttp(&server.Server{})
	if srv == nil {
		t.Fatal("newHttp(empty config) returned nil server")
	}
}
