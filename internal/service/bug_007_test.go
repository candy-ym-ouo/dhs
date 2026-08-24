package service

import (
	"context"
	"dhs/internal/model"
	"dhs/internal/store"
	"testing"
	"time"
)

type contextProbeStore struct {
	node     model.Node
	recorded context.Context
}

func (s *contextProbeStore) Close() error { return nil }
func (s *contextProbeStore) Register(context.Context, model.RegisterRequest) (model.Node, error) {
	return model.Node{}, nil
}
func (s *contextProbeStore) GetNode(context.Context, string) (model.Node, error) { return s.node, nil }
func (s *contextProbeStore) ListNodes(context.Context, store.Filter) ([]model.Node, int, error) {
	return nil, 0, nil
}
func (s *contextProbeStore) RecordHeartbeat(ctx context.Context, id string, req model.HeartbeatRequest, at time.Time) (model.Heartbeat, error) {
	s.recorded = ctx
	return model.Heartbeat{NodeID: id, Load: req.Load, ReportedAt: at}, nil
}
func (s *contextProbeStore) SetStatus(context.Context, string, model.Status, string, string, string, time.Time) (bool, error) {
	return false, nil
}
func (s *contextProbeStore) StartRecovery(context.Context, string, time.Time) (bool, error) {
	return false, nil
}
func (s *contextProbeStore) ListTransitions(context.Context, string, int) ([]model.Transition, error) {
	return nil, nil
}
func (s *contextProbeStore) ListHeartbeats(context.Context, string, int) ([]model.Heartbeat, error) {
	return nil, nil
}
func (s *contextProbeStore) Stats(context.Context) (map[model.Status]int, error) { return nil, nil }
func (s *contextProbeStore) ScanCandidates(context.Context, time.Time, time.Duration) ([]model.Node, error) {
	return nil, nil
}
func (s *contextProbeStore) Cleanup(context.Context, time.Time) error { return nil }

func TestBug007CancelledContextReachesDownstream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	probe := &contextProbeStore{node: model.Node{Status: model.Offline}}
	service := &Service{Store: probe}
	if _, _, err := service.Heartbeat(ctx, "node-7", model.HeartbeatRequest{Load: 0.5}); err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if probe.recorded == nil || probe.recorded.Err() != context.Canceled {
		t.Fatalf("cancelled context was replaced or lost")
	}
}
