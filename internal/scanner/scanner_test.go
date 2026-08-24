package scanner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"dhs/internal/config"
	"dhs/internal/model"
	"dhs/internal/service"
	"dhs/internal/store/sqlite"
)

// newTestScanner builds a Scanner backed by an in-memory SQLite store and a
// short scan interval so loops tick promptly under test.
func newTestScanner(t *testing.T, cfg config.Config) *Scanner {
	t.Helper()
	db, e := sqlite.Open(":memory:")
	if e != nil {
		t.Fatalf("open sqlite: %v", e)
	}
	t.Cleanup(func() { _ = db.Close() })
	svc := &service.Service{Store: db, ConfirmCount: cfg.ConfirmCount}
	return &Scanner{Store: db, Service: svc, Config: cfg}
}

// testConfig returns a config with a sub-millisecond scan interval so the loop
// ticks many times during a short test. HeartbeatTimeout is short enough that a
// node with no heartbeats is immediately a candidate.
func testConfig() config.Config {
	return config.Config{
		ScanInterval:     time.Millisecond,
		HeartbeatTimeout: time.Millisecond,
		RecoveryGrace:    time.Hour,
		MaxLostDuration:  2 * time.Hour,
		Retention:        time.Hour,
		ConfirmCount:     1,
	}
}

func TestScanner_SingleLoop(t *testing.T) {
	sc := newTestScanner(t, testConfig())
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	if e := sc.Start(ctx); e != nil {
		t.Fatalf("first Start: %v", e)
	}

	// A second Start on the same scanner must not launch another loop.
	var (
		wg       sync.WaitGroup
		dupCount atomic.Int64
	)
	const n = 16
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if e := sc.Start(ctx); e != nil && errors.Is(e, ErrScannerAlreadyRunning) {
				dupCount.Add(1)
			}
		}()
	}
	wg.Wait()
	if got := dupCount.Load(); got != n {
		t.Fatalf("expected %d ErrScannerAlreadyRunning, got %d", n, got)
	}

	// Stop drains the loop and must allow a fresh Start afterward.
	sc.Stop()
	if e := sc.Start(ctx); e != nil {
		t.Fatalf("restart after Stop: %v", e)
	}
	sc.Stop()
}

func TestScanner_RoundConcurrentNoDupTransitions(t *testing.T) {
	// Two candidates: one ONLINE (fresh), one REGISTERED (fresh). Both are
	// timed out by HeartbeatTimeout. Concurrent rounds must not produce
	// duplicate Lost/recovery_start transitions nor double recover_attempts.
	cfg := config.Config{
		ScanInterval:     time.Millisecond,
		HeartbeatTimeout: 30 * time.Second,
		RecoveryGrace:    time.Hour,
		MaxLostDuration:  2 * time.Hour,
		Retention:        time.Hour,
		ConfirmCount:     1,
	}
	sc := newTestScanner(t, cfg)
	ctx := context.Background()
	now := time.Now().UTC()

	regReq := model.RegisterRequest{ID: "n1", Name: "node-1", Address: "127.0.0.1:1"}
	if _, e := sc.Service.Register(ctx, regReq); e != nil {
		t.Fatalf("register n1: %v", e)
	}
	regReq2 := model.RegisterRequest{ID: "n2", Name: "node-2", Address: "127.0.0.1:2"}
	if _, e := sc.Service.Register(ctx, regReq2); e != nil {
		t.Fatalf("register n2: %v", e)
	}
	// Both nodes are REGISTERED with no heartbeat → candidates once
	// HeartbeatTimeout has elapsed. Use a now far in the future so they're
	// definitely timed out.
	far := now.Add(time.Minute)

	const rounds = 32
	var wg sync.WaitGroup
	wg.Add(rounds)
	for i := 0; i < rounds; i++ {
		go func() {
			defer wg.Done()
			sc.Round(ctx, far)
		}()
	}
	wg.Wait()

	for _, id := range []string{"n1", "n2"} {
		n, e := sc.Store.GetNode(ctx, id)
		if e != nil {
			t.Fatalf("GetNode %s: %v", id, e)
		}
		if n.RecoverAttempts > 1 {
			t.Fatalf("node %s: recover_attempts=%d, want <=1 (duplicate recovery)", id, n.RecoverAttempts)
		}
		if n.RecoverAttempts != 1 {
			t.Fatalf("node %s: recover_attempts=%d, want exactly 1", id, n.RecoverAttempts)
		}
		trs, e := sc.Store.ListTransitions(ctx, id, 100)
		if e != nil {
			t.Fatalf("ListTransitions %s: %v", id, e)
		}
		var lost, recoveryStart int
		for _, tr := range trs {
			switch tr.Reason {
			case "heartbeat_timeout":
				lost++
			case "recovery_start":
				recoveryStart++
			}
		}
		if lost != 1 {
			t.Fatalf("node %s: heartbeat_timeout transitions=%d, want 1", id, lost)
		}
		if recoveryStart != 1 {
			t.Fatalf("node %s: recovery_start transitions=%d, want 1", id, recoveryStart)
		}
	}
}

func TestScanner_StopWaitsForRound(t *testing.T) {
	cfg := testConfig()
	sc := newTestScanner(t, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	if e := sc.Start(ctx); e != nil {
		t.Fatalf("Start: %v", e)
	}
	// Let a few rounds run.
	time.Sleep(20 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		sc.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return within 5s")
	}
	// After Stop, Start must succeed again (lifecycle reset).
	if e := sc.Start(ctx); e != nil {
		t.Fatalf("restart after Stop: %v", e)
	}
	sc.Stop()
}

func TestScanner_LastRoundAtAndMetrics(t *testing.T) {
	cfg := testConfig()
	sc := newTestScanner(t, cfg)
	ctx := context.Background()

	if !sc.LastRoundAt().IsZero() {
		t.Fatal("LastRoundAt should be zero before any round")
	}
	now := time.Now().UTC()
	sc.Round(ctx, now)
	if sc.LastRoundAt().IsZero() {
		t.Fatal("LastRoundAt should be set after a round")
	}
	m := sc.Metrics()
	rounds, ok := m["rounds"].(uint64)
	if !ok {
		t.Fatalf("metrics rounds missing or wrong type: %#v", m["rounds"])
	}
	if rounds != 1 {
		t.Fatalf("rounds=%d, want 1", rounds)
	}
}
