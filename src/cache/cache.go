// implements a simple, in-memory, TTL-based cache
package cache

import (
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// single cached DNS response
type entry struct {
	msg       *dns.Msg  // DNS response received from the upstream
	expiresAt time.Time // expiry date of the entry
}

// stores DNS responses
type Cache struct {
	mu      sync.RWMutex // RWMutex is required to protect writes from multiple goroutines
	entries map[string]entry
}

// constructor for cache
func New() *Cache {
	return &Cache{entries: make(map[string]entry)}
}

// builds a map key for DNS
// e.g example.com|A|IN
func Key(q dns.Question) string {
	return strings.ToLower(q.Name) + "|" +
		dns.TypeToString[q.Qtype] + "|" +
		dns.ClassToString[q.Qclass]
}

// returns a cached response for a key
func (c *Cache) Get(key string) (msg *dns.Msg, ok bool) {
	c.mu.RLock()
	e, found := c.entries[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	remaining := time.Until(e.expiresAt)
	if remaining <= 0 {
		return nil, false // Expired. Cleanup goroutine will delete it.
	}

	out := e.msg.Copy()
	remainingSeconds := uint32(remaining.Seconds())
	rewriteTTL(out.Answer, remainingSeconds)
	rewriteTTL(out.Ns, remainingSeconds)
	rewriteTTL(out.Extra, remainingSeconds)
	return out, true
}

// set every record's TTL to seconds
func rewriteTTL(records []dns.RR, ttl uint32) {
	for _, rr := range records {
		if _, isOPT := rr.(*dns.OPT); isOPT {
			continue
		}
		rr.Header().Ttl = ttl
	}
}

// stores a response under key
func (c *Cache) Set(key string, msg *dns.Msg) {
	ttl := MinTTL(msg)
	if ttl == 0 {
		return
	}
	c.mu.Lock()
	c.entries[key] = entry{
		msg:       msg.Copy(),
		expiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}
	c.mu.Unlock()
}

// returns smallest TTL among answer records
func MinTTL(msg *dns.Msg) uint32 {
	var min uint32
	first := true
	for _, rr := range msg.Answer {
		t := rr.Header().Ttl
		if first || t < min {
			min, first = t, false
		}
	}
	if first {
		return 0
	}
	return min
}

// returns number of currently stored entries
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// removes expired entries
func (c *Cache) DeleteExpired() int {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	removed := 0
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
			removed++
		}
	}
	return removed
}

// launch a goroutine that calls DeleteExpired every interval
func (c *Cache) StartCleanup(interval time.Duration) (stop func()) {
	ticker := time.NewTicker(interval)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				c.DeleteExpired()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()
	return func() { close(done) }
}
