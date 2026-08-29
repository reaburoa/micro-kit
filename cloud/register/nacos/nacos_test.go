package nacos

import (
	"testing"

	"github.com/reaburoa/micro-kit/protos"
)

func TestNormalizeClientConfig(t *testing.T) {
	cfg := &protos.NacosClientConfig{TimeoutMs: 0, LogLevel: ""}
	normalizeClientConfig(cfg)
	if cfg.TimeoutMs == 0 {
		t.Fatal("expected default timeout to be applied")
	}
	if cfg.LogLevel == "" {
		t.Fatal("expected default log level to be applied")
	}
}

func TestBuildRegisterInstanceParam(t *testing.T) {
	param, err := buildRegisterInstanceParam("127.0.0.1", 8080, "order-service", "custom-group", map[string]string{"env": "dev"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if param.Ip != "127.0.0.1" {
		t.Fatalf("ip mismatch: got %q", param.Ip)
	}
	if param.ServiceName != "order-service" {
		t.Fatalf("service name mismatch: got %q", param.ServiceName)
	}
	if param.GroupName != "custom-group" {
		t.Fatalf("group mismatch: got %q", param.GroupName)
	}
	if param.Weight <= 0 {
		t.Fatal("weight must be positive")
	}
	if !param.Enable || !param.Healthy {
		t.Fatal("instance should be enabled and healthy by default")
	}
	if param.Metadata["env"] != "dev" {
		t.Fatalf("metadata mismatch: got %q", param.Metadata["env"])
	}
}
