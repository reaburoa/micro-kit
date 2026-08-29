package rocketmq

import (
	"testing"
	"time"
)

func TestNewMessageSupportsDelayAndOrderingKeys(t *testing.T) {
	msg := newMessage("demo-topic", []byte("hello"), WithKeys("order-1", "order-2"), WithDelay(5*time.Second), WithMsgTag("tag-a"))
	if msg == nil {
		t.Fatal("newMessage returned nil")
	}
	if msg.Topic != "demo-topic" {
		t.Fatalf("topic = %q, want %q", msg.Topic, "demo-topic")
	}
	if msg.GetTags() != "tag-a" {
		t.Fatalf("tags = %q, want %q", msg.GetTags(), "tag-a")
	}
	if msg.GetKeys() == "" {
		t.Fatal("expected keys to be set for ordering")
	}
	if msg.GetShardingKey() == "" {
		t.Fatal("expected sharding key to be set for ordering")
	}
}
