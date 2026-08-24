package scanner

import (
	"sync"
	"time"
)

// roundState holds in-process view-of-the-world state shared between the scan
// loop and any readers (e.g. diagnostics). It is guarded by mu so concurrent
// goroutines never race on lastRound.
type roundState struct {
	mu        sync.RWMutex
	lastRound time.Time
}

func (r *roundState) mark(now time.Time) {
	r.mu.Lock()
	r.lastRound = now
	r.mu.Unlock()
}

func (r *roundState) lastRoundAt() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastRound
}
