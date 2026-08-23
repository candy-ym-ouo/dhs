package scanner

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	Rounds    atomic.Uint64
	Lost      atomic.Uint64
	Recovered atomic.Uint64
	Retired   atomic.Uint64
	LastRound atomic.Int64
}

func (m *Metrics) Begin(now time.Time) { m.Rounds.Add(1); m.LastRound.Store(now.UnixNano()) }
func (m *Metrics) Snapshot() map[string]any {
	return map[string]any{"rounds": m.Rounds.Load(), "lost": m.Lost.Load(), "recovered": m.Recovered.Load(), "retired": m.Retired.Load(), "last_round": time.Unix(0, m.LastRound.Load()).UTC()}
}
