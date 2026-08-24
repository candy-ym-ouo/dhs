package scanner

import (
	"context"
	"dhs/internal/model"
	"dhs/internal/store"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var errRoundProbe = errors.New("round probe")

type roundProbeStore struct{ scans atomic.Int64 }

func (s *roundProbeStore) Close() error { return nil }
func (s *roundProbeStore) Register(context.Context, model.RegisterRequest) (model.Node, error) {
	return model.Node{}, nil
}
func (s *roundProbeStore) GetNode(context.Context, string) (model.Node, error) {
	return model.Node{}, nil
}
func (s *roundProbeStore) ListNodes(context.Context, store.Filter) ([]model.Node, int, error) {
	return nil, 0, nil
}
func (s *roundProbeStore) RecordHeartbeat(context.Context, string, model.HeartbeatRequest, time.Time) (model.Heartbeat, error) {
	return model.Heartbeat{}, nil
}
func (s *roundProbeStore) SetStatus(context.Context, string, model.Status, string, string, string, time.Time) (bool, error) {
	return false, nil
}
func (s *roundProbeStore) StartRecovery(context.Context, string, time.Time) (bool, error) {
	return false, nil
}
func (s *roundProbeStore) ListTransitions(context.Context, string, int) ([]model.Transition, error) {
	return nil, nil
}
func (s *roundProbeStore) ListHeartbeats(context.Context, string, int) ([]model.Heartbeat, error) {
	return nil, nil
}
func (s *roundProbeStore) Stats(context.Context) (map[model.Status]int, error) { return nil, nil }
func (s *roundProbeStore) ScanCandidates(context.Context, time.Time, time.Duration) ([]model.Node, error) {
	s.scans.Add(1)
	return nil, errRoundProbe
}
func (s *roundProbeStore) Cleanup(context.Context, time.Time) error { return nil }

func TestBug009RoundMarkerIsSafeAcrossWorkers(t *testing.T) {
	probe := &roundProbeStore{}
	scanner := &Scanner{Store: probe}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			scanner.Round(context.Background(), time.Unix(int64(i), 0))
		}(i)
	}
	close(start)
	wg.Wait()
	if probe.scans.Load() != 64 {
		t.Fatalf("expected 64 scans, got %d", probe.scans.Load())
	}
}
