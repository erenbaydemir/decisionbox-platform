package notify

import (
	"context"
	"log"
	"sync"
	"time"
)

var (
	registryMu   sync.RWMutex
	channels     = make(map[string]Channel)
	channelMetas = make(map[string]ChannelMeta)
)

// Register makes a notification channel available by type.
// Channel plugins call this in their init() function.
func Register(ch Channel, meta ChannelMeta) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if ch == nil {
		panic("notify: Register channel is nil")
	}
	t := ch.Type()
	if _, exists := channels[t]; exists {
		panic("notify: Register called twice for " + t)
	}
	channels[t] = ch
	meta.Type = t
	channelMetas[t] = meta
}

// GetChannel returns a registered channel by type, or nil if not found.
func GetChannel(channelType string) Channel {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return channels[channelType]
}

// RegisteredChannels returns metadata for all registered notification channels.
func RegisteredChannels() []ChannelMeta {
	registryMu.RLock()
	defer registryMu.RUnlock()
	metas := make([]ChannelMeta, 0, len(channelMetas))
	for _, m := range channelMetas {
		metas = append(metas, m)
	}
	return metas
}

// NotifyAll dispatches an event to all registered channels concurrently
// and waits for them to complete (up to 30 seconds). This ensures
// notifications are sent before short-lived processes (like the agent
// subprocess) exit. Errors are logged but never returned — notification
// failures must not affect discovery completion.
func NotifyAll(ctx context.Context, event Event) {
	registryMu.RLock()
	chs := make([]Channel, 0, len(channels))
	for _, ch := range channels {
		chs = append(chs, ch)
	}
	registryMu.RUnlock()

	if len(chs) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(chs))

	for _, ch := range chs {
		go func(c Channel) {
			defer wg.Done()
			if err := c.Notify(ctx, event); err != nil {
				log.Printf("[notify] %s: %v", c.Type(), err)
			}
		}(ch)
	}

	// Wait with timeout so we don't block forever on a hung channel
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		log.Printf("[notify] timed out waiting for channels to complete")
	}
}
