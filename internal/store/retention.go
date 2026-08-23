package store

import (
	"dhs/internal/model"
	"sort"
	"time"
)

type RetentionPolicy struct {
	Heartbeats  time.Duration
	Transitions time.Duration
	MaxRows     int
}

func DefaultRetention() RetentionPolicy {
	return RetentionPolicy{Heartbeats: 7 * 24 * time.Hour, Transitions: 30 * 24 * time.Hour, MaxRows: 100000}
}
func (p RetentionPolicy) Valid() bool                              { return p.Heartbeats > 0 && p.Transitions > 0 && p.MaxRows > 0 }
func (p RetentionPolicy) HeartbeatCutoff(now time.Time) time.Time  { return now.Add(-p.Heartbeats) }
func (p RetentionPolicy) TransitionCutoff(now time.Time) time.Time { return now.Add(-p.Transitions) }
func SortTransitions(xs []model.Transition) []model.Transition {
	out := append([]model.Transition(nil), xs...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}
func SortHeartbeats(xs []model.Heartbeat) []model.Heartbeat {
	out := append([]model.Heartbeat(nil), xs...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReportedAt.After(out[j].ReportedAt) })
	return out
}
func LimitTransitions(xs []model.Transition, n int) []model.Transition {
	if n < 1 || len(xs) <= n {
		return xs
	}
	return xs[:n]
}
func LimitHeartbeats(xs []model.Heartbeat, n int) []model.Heartbeat {
	if n < 1 || len(xs) <= n {
		return xs
	}
	return xs[:n]
}
func CountByReason(xs []model.Transition) map[string]int {
	m := map[string]int{}
	for _, x := range xs {
		m[x.Reason]++
	}
	return m
}
func CountByStatus(xs []model.Node) map[model.Status]int {
	m := map[model.Status]int{}
	for _, x := range xs {
		m[x.Status]++
	}
	return m
}
func LatestHeartbeat(xs []model.Heartbeat) (model.Heartbeat, bool) {
	if len(xs) == 0 {
		return model.Heartbeat{}, false
	}
	best := xs[0]
	for _, x := range xs[1:] {
		if x.ReportedAt.After(best.ReportedAt) {
			best = x
		}
	}
	return best, true
}
func LatestTransition(xs []model.Transition) (model.Transition, bool) {
	if len(xs) == 0 {
		return model.Transition{}, false
	}
	best := xs[0]
	for _, x := range xs[1:] {
		if x.CreatedAt.After(best.CreatedAt) {
			best = x
		}
	}
	return best, true
}
