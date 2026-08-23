package service

import (
	"context"
	"dhs/internal/model"
	"sort"
	"strings"
	"sync"
	"time"
)

// Registry is a small in-process view used by diagnostics and tests. The
// durable source of truth remains the configured Store implementation.
type Registry struct {
	mu      sync.RWMutex
	nodes   map[string]model.Node
	updated time.Time
}

func NewRegistry() *Registry { return &Registry{nodes: map[string]model.Node{}} }
func (r *Registry) Upsert(n model.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[n.ID] = n.Clone()
	r.updated = time.Now().UTC()
}
func (r *Registry) Remove(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[id]; !ok {
		return false
	}
	delete(r.nodes, id)
	r.updated = time.Now().UTC()
	return true
}
func (r *Registry) Get(id string) (model.Node, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.nodes[id]
	return n.Clone(), ok
}
func (r *Registry) Count() int           { r.mu.RLock(); defer r.mu.RUnlock(); return len(r.nodes) }
func (r *Registry) UpdatedAt() time.Time { r.mu.RLock(); defer r.mu.RUnlock(); return r.updated }
func (r *Registry) List(status model.Status, keyword string) []model.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(keyword))
	out := make([]model.Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		if status != "" && n.Status != status {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(n.ID+" "+n.Name+" "+n.Address), q) {
			continue
		}
		out = append(out, n.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Registry) Snapshot() map[string]model.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]model.Node, len(r.nodes))
	for k, n := range r.nodes {
		out[k] = n.Clone()
	}
	return out
}
func (s *Service) RegisterAndCache(ctx context.Context, r *Registry, req model.RegisterRequest) (model.Node, error) {
	n, e := s.Register(ctx, req)
	if e == nil {
		r.Upsert(n)
	}
	return n, e
}
func (s *Service) HeartbeatAndCache(ctx context.Context, r *Registry, id string, req model.HeartbeatRequest) (model.Heartbeat, bool, error) {
	h, c, e := s.Heartbeat(ctx, id, req)
	if e == nil {
		if n, x := s.Node(ctx, id); x == nil {
			r.Upsert(n)
		}
	}
	return h, c, e
}
func (r *Registry) StatusCounts() map[model.Status]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := map[model.Status]int{}
	for _, n := range r.nodes {
		out[n.Status]++
	}
	return out
}
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.nodes))
	for id := range r.nodes {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
func (r *Registry) Replace(nodes []model.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes = make(map[string]model.Node, len(nodes))
	for _, n := range nodes {
		r.nodes[n.ID] = n.Clone()
	}
	r.updated = time.Now().UTC()
}
func (r *Registry) Stale(now time.Time, timeout time.Duration) []model.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []model.Node{}
	for _, n := range r.nodes {
		if n.LastHeartbeatAt == nil || now.Sub(*n.LastHeartbeatAt) >= timeout {
			out = append(out, n.Clone())
		}
	}
	return out
}
