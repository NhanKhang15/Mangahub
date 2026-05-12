package server

import (
	"net"
	"sync"
	"time"
)

// Registry tracks every client that has sent a REGISTER datagram. Entries
// expire after TTL elapses without a heartbeat, mirroring the way UDP clients
// often disappear silently.
type Registry struct {
	mu      sync.Mutex
	entries map[string]*entry // key = addr.String()
	ttl     time.Duration
}

type entry struct {
	addr     *net.UDPAddr
	clientID string
	lastSeen time.Time
}

func NewRegistry(ttl time.Duration) *Registry {
	return &Registry{
		entries: make(map[string]*entry),
		ttl:     ttl,
	}
}

// Register adds or refreshes a client. Returns whether this was a new entry.
func (r *Registry) Register(addr *net.UDPAddr, clientID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := addr.String()
	_, existed := r.entries[key]
	r.entries[key] = &entry{addr: addr, clientID: clientID, lastSeen: time.Now()}
	return !existed
}

// Touch refreshes lastSeen on an existing entry. No-op if unknown.
func (r *Registry) Touch(addr *net.UDPAddr) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entries[addr.String()]
	if !ok {
		return false
	}
	e.lastSeen = time.Now()
	return true
}

// Remove drops an entry on UNREGISTER. Returns whether anything was removed.
func (r *Registry) Remove(addr *net.UDPAddr) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := addr.String()
	if _, ok := r.entries[key]; !ok {
		return false
	}
	delete(r.entries, key)
	return true
}

// Active returns a snapshot of every non-expired UDP address. The returned
// slice is safe to use without holding the lock.
func (r *Registry) Active() []*net.UDPAddr {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-r.ttl)
	out := make([]*net.UDPAddr, 0, len(r.entries))
	for _, e := range r.entries {
		if e.lastSeen.After(cutoff) {
			out = append(out, e.addr)
		}
	}
	return out
}

// Count returns the total number of registered (not necessarily active) clients.
func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}

// GC drops entries whose lastSeen is older than TTL. Returns how many were
// removed.
func (r *Registry) GC() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-r.ttl)
	removed := 0
	for k, e := range r.entries {
		if e.lastSeen.Before(cutoff) {
			delete(r.entries, k)
			removed++
		}
	}
	return removed
}
