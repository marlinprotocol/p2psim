package pubsub

import (
	"time"

	"github.com/marlinprotocol/p2psim/core"
)

// Usecases of a seen cache
// - check if we already processed a message
// - in gossipsub, we request only those messages that are not present
// NOTE: Negative TTL is disallowed in the stat collector module

type SeenCache struct {
	// set of msgs already processed
	seenIDs *core.Set
	// msg IDs sorted in non-decreasing order of entry times
	// entry times of existing messages are not updated to mimic the official golang implementation
	// entries are swept on seeing new messages to mimic the official golang implementation
	entryTimes []SeenEntry
	// configurable duration until which entries are kept
	seenTTL time.Duration
}

// helps seen cache retire expired msgs
type SeenEntry struct {
	// cached msg ID
	msgID interface{}
	// time when the entry was inserted into the cache
	entryTime time.Time
}

// Construct a new seen cache that retains elements that are no older than `seenTTL` from current simulator time
//   older messages are retired when new messages are added
func NewSeenCache(seenTTL time.Duration) *SeenCache {
	return &SeenCache{
		seenIDs:    core.NewSet(),
		entryTimes: []SeenEntry{},
		seenTTL:    seenTTL,
	}
}

// Marks the message as seen
// Returns true if the message was not already seen
func (cache *SeenCache) MarkSeen(msgID interface{}, curTime time.Time) bool {
	if cache.seenIDs.Exists(msgID) {
		// we do not sweep here to mimic the official golang implementation
		return false
	}

	cache.sweep(curTime)
	cache.entryTimes = append(cache.entryTimes, SeenEntry{
		msgID:     msgID,
		entryTime: curTime,
	})
	cache.seenIDs.Add(msgID)
	return true
}

func (cache *SeenCache) SeenMsg(msgID interface{}) bool {
	return cache.seenIDs.Exists(msgID)
}

func (cache *SeenCache) sweep(curTime time.Time) {
	oldestValidTime := curTime.Add(-1 * cache.seenTTL)
	// Check whether the next entry has an entry time lower than the oldest valid time
	// - if it doesn't we are done since the slice is non-decreasing in entry times
	// - if it is lower, then we remove the msg from both the entry times and seen IDs
	for 0 < len(cache.entryTimes) && cache.entryTimes[0].entryTime.Before(oldestValidTime) {
		cache.seenIDs.Remove(cache.entryTimes[0].msgID)
		// remove the first element
		// the first element is garbage collected on reallocation
		cache.entryTimes = cache.entryTimes[1:]
	}
}
