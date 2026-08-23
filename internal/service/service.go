package service

import (
	"context"
	"dhs/internal/heartbeat"
	"dhs/internal/model"
	"dhs/internal/store"
	"errors"
	"time"
)

var ErrNotFound = errors.New("node not found")
var ErrConflict = errors.New("invalid state transition")

type Service struct {
	Store        store.Store
	ConfirmCount int
}

func (s *Service) Register(ctx context.Context, r model.RegisterRequest) (model.Node, error) {
	if err := r.Validate(); err != nil {
		return model.Node{}, err
	}
	r.AddSourceLabel()
	return s.Store.Register(ctx, r)
}
func (s *Service) Heartbeat(ctx context.Context, id string, r model.HeartbeatRequest) (model.Heartbeat, bool, error) {
	n, e := s.Store.GetNode(ctx, id)
	if e != nil {
		return model.Heartbeat{}, false, ErrNotFound
	}
	if err := r.Validate(); err != nil {
		return model.Heartbeat{}, false, err
	}
	now := time.Now().UTC()
	h, e := s.Store.RecordHeartbeat(ctx, id, r, now)
	if e != nil {
		return h, false, e
	}
	changed := false
	switch n.Status {
	case model.Registered:
		changed, e = s.transition(ctx, id, model.Online, "first_heartbeat", "heartbeat", now)
	case model.Lost, model.Recovering:
		changed, e = s.transition(ctx, id, model.Online, "heartbeat_recovered", "heartbeat", now)
	}
	return h, changed, e
}
func (s *Service) transition(ctx context.Context, id string, to model.Status, reason, trigger string, now time.Time) (bool, error) {
	n, e := s.Store.GetNode(ctx, id)
	if e != nil {
		return false, e
	}
	if !heartbeat.CanTransition(n.Status, to) {
		return false, ErrConflict
	}
	return s.Store.SetStatus(ctx, id, to, reason, trigger, "", now)
}
func (s *Service) Recover(ctx context.Context, id string) (bool, error) {
	n, e := s.Store.GetNode(ctx, id)
	if e != nil {
		return false, ErrNotFound
	}
	if n.Status != model.Lost && n.Status != model.Recovering {
		return false, ErrConflict
	}
	return s.Store.StartRecovery(ctx, id, time.Now().UTC())
}
func (s *Service) Retire(ctx context.Context, id string) (bool, error) {
	if _, e := s.Store.GetNode(ctx, id); e != nil {
		return false, ErrNotFound
	}
	return s.transition(ctx, id, model.Offline, "manual_retire", "manual", time.Now().UTC())
}
func (s *Service) Node(ctx context.Context, id string) (model.Node, error) {
	n, e := s.Store.GetNode(ctx, id)
	if e != nil {
		return n, ErrNotFound
	}
	return n, nil
}
func (s *Service) List(ctx context.Context, f store.Filter) ([]model.Node, int, error) {
	return s.Store.ListNodes(ctx, f)
}
func (s *Service) Transitions(ctx context.Context, id string, l int) ([]model.Transition, error) {
	return s.Store.ListTransitions(ctx, id, l)
}
func (s *Service) Heartbeats(ctx context.Context, id string, l int) ([]model.Heartbeat, error) {
	return s.Store.ListHeartbeats(ctx, id, l)
}
func (s *Service) Stats(ctx context.Context) (map[model.Status]int, error) { return s.Store.Stats(ctx) }
