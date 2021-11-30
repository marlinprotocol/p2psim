package pubsub

import (
	"testing"
	"time"
)

func TestExpiry(t *testing.T) {
	seenCache := NewSeenCache(1 * time.Second)
	firstMsg := 531
	if seenCache.SeenMsg(firstMsg) {
		t.Error("Message was never marked seen!")
	}
	epoch := time.Time{}
	seenCache.MarkSeen(firstMsg, epoch)
	if !seenCache.SeenMsg(firstMsg) {
		t.Error("Message was marked seen!")
	}
	nextMsg := 864
	seenCache.MarkSeen(nextMsg, epoch.Add(2*time.Second))
	if seenCache.SeenMsg(firstMsg) {
		t.Error("Message was retired!")
	}
	if !seenCache.SeenMsg(nextMsg) {
		t.Error("Message was marked seen!")
	}
}
