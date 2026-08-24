package scanner

import (
	"context"
	"dhs/internal/config"
	"dhs/internal/model"
	"dhs/internal/service"
	"dhs/internal/store"
	"errors"
	"log/slog"
	"sync"
	"time"
)

// ErrScannerAlreadyRunning is returned by Start when a scan loop is already
// active for this Scanner. A Scanner owns at most one live loop at a time.
var ErrScannerAlreadyRunning = errors.New("scanner: loop already running")

type Scanner struct {
	Store   store.Store
	Service *service.Service
	Config  config.Config
	Log     *slog.Logger

	metrics Metrics
	state   roundState

	// roundMu serializes Round execution. At most one Round runs at any time,
	// so concurrent callers never overlap the read-candidates/transition-write
	// window that caused duplicate migrations and double recover_attempts.
	roundMu sync.Mutex

	// runMu guards the single-loop lifecycle fields so Start/Stop are
	// themselves concurrency-safe.
	runMu      sync.Mutex
	running    bool
	loopCancel context.CancelFunc
	loopDone   chan struct{}
}

// Start launches the single background scan loop. It is idempotent: calling it
// while a loop is already running returns ErrScannerAlreadyRunning and does not
// start a second loop. The loop runs until either ctx is canceled or Stop is
// called. Stop must be invoked to release the loop before starting again.
func (s *Scanner) Start(ctx context.Context) error {
	s.runMu.Lock()
	if s.running {
		s.runMu.Unlock()
		return ErrScannerAlreadyRunning
	}
	loopCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	s.loopCancel = cancel
	s.loopDone = done
	s.running = true
	s.runMu.Unlock()

	go func() {
		defer close(done)
		defer cancel()
		s.Run(loopCtx)
	}()
	return nil
}

// Stop signals the scan loop to exit and waits for it to drain. It is safe to
// call when no loop is active (no-op). After Stop returns the Scanner may be
// started again via Start.
func (s *Scanner) Stop() {
	s.runMu.Lock()
	if !s.running {
		s.runMu.Unlock()
		return
	}
	cancel := s.loopCancel
	done := s.loopDone
	s.runMu.Unlock()

	cancel()
	<-done

	s.runMu.Lock()
	s.running = false
	s.loopCancel = nil
	s.loopDone = nil
	s.runMu.Unlock()
}

// Run is the scan loop body. It ticks at Config.ScanInterval, running Round on
// each tick until ctx is canceled. Start invokes this on a managed goroutine;
// callers may also invoke it directly, but prefer Start/Stop for lifecycle
// control.
func (s *Scanner) Run(ctx context.Context) {
	t := time.NewTicker(s.Config.ScanInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			s.Round(ctx, now)
		}
	}
}

// Round executes one scan pass. The entire pass is serialized by roundMu, so
// overlapping rounds (from a double-started loop or concurrent callers) cannot
// race on shared state or issue duplicate transitions.
func (s *Scanner) Round(ctx context.Context, now time.Time) {
	s.roundMu.Lock()
	defer s.roundMu.Unlock()

	now = now.UTC()
	s.state.mark(now)
	s.metrics.Begin(now)

	ns, e := s.Store.ScanCandidates(ctx, now, s.Config.HeartbeatTimeout)
	if e != nil {
		return
	}
	for _, n := range ns {
		if n.Status == model.Registered || n.Status == model.Online {
			if changed, _ := s.Store.SetStatus(ctx, n.ID, model.Lost, "heartbeat_timeout", "scanner", "", now); changed {
				s.metrics.Lost.Add(1)
			}
			_, _ = s.Store.StartRecovery(ctx, n.ID, now)
		}
		if n.Status == model.Recovering && n.LostAt != nil && now.Sub(*n.LostAt) > s.Config.MaxLostDuration {
			if changed, _ := s.Store.SetStatus(ctx, n.ID, model.Offline, "retired", "scanner", "", now); changed {
				s.metrics.Retired.Add(1)
			}
		} else if n.Status == model.Recovering && n.LostAt != nil && now.Sub(*n.LostAt) >= s.Config.RecoveryGrace {
			if changed, _ := s.Store.SetStatus(ctx, n.ID, model.Lost, "recovery_failed", "scanner", "", now); changed {
				s.metrics.Lost.Add(1)
			}
		}
	}
	_ = s.Store.Cleanup(ctx, now.Add(-s.Config.Retention))
}

// LastRoundAt returns the wall time of the most recent Round start, or the zero
// time if no round has run. Safe to call concurrently with Round.
func (s *Scanner) LastRoundAt() time.Time { return s.state.lastRoundAt() }

// Metrics returns a snapshot of the scanner's atomic counters. Safe for
// concurrent use.
func (s *Scanner) Metrics() map[string]any { return s.metrics.Snapshot() }
