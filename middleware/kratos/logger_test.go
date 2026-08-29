package kratos

import (
	"testing"

	"github.com/go-kratos/kratos/v2/transport"
)

func TestGetIpNilHeader(t *testing.T) {
	if got := getIp(nil); got != "" {
		t.Fatalf("getIp(nil) = %q, want empty string", got)
	}
	if got := getUa(nil); got != "" {
		t.Fatalf("getUa(nil) = %q, want empty string", got)
	}
	if got := getIp(transport.Header(nil)); got != "" {
		t.Fatalf("getIp(transport.Header(nil)) = %q, want empty string", got)
	}
}
