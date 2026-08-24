package service

import (
	"context"
	"dhs/internal/model"
	"dhs/internal/store"
	"errors"
	"testing"
	"time"
)

type errorProbeStore struct{ err error }

func (s *errorProbeStore) Close() error { return nil }
func (s *errorProbeStore) Register(context.Context, model.RegisterRequest) (model.Node, error) {
	return model.Node{}, nil
}
func (s *errorProbeStore) GetNode(context.Context, string) (model.Node, error) {
	return model.Node{}, s.err
}
func (s *errorProbeStore) ListNodes(context.Context, store.Filter) ([]model.Node, int, error) {
	return nil, 0, nil
}
func (s *errorProbeStore) RecordHeartbeat(context.Context, string, model.HeartbeatRequest, time.Time) (model.Heartbeat, error) {
	return model.Heartbeat{}, nil
}
func (s *errorProbeStore) SetStatus(context.Context, string, model.Status, string, string, string, time.Time) (bool, error) {
	return false, nil
}
func (s *errorProbeStore) StartRecovery(context.Context, string, time.Time) (bool, error) {
	return false, nil
}
func (s *errorProbeStore) ListTransitions(context.Context, string, int) ([]model.Transition, error) {
	return nil, nil
}
func (s *errorProbeStore) ListHeartbeats(context.Context, string, int) ([]model.Heartbeat, error) {
	return nil, nil
}
func (s *errorProbeStore) Stats(context.Context) (map[model.Status]int, error) { return nil, nil }
func (s *errorProbeStore) ScanCandidates(context.Context, time.Time, time.Duration) ([]model.Node, error) {
	return nil, nil
}
func (s *errorProbeStore) Cleanup(context.Context, time.Time) error { return nil }

func TestBug008ErrorIdentitySurvivesServiceBoundary(t *testing.T) {
	sentinel := errors.New("storage unavailable")
	service := &Service{Store: &errorProbeStore{err: sentinel}}
	_, err := service.Node(context.Background(), "node-8")
	if !errors.Is(err, sentinel) {
		t.Fatalf("service error lost its identity: %v", err)
	}
}
