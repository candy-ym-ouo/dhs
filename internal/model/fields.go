package model

import (
	"sort"
	"strings"
	"time"
)

func NormalizeNode(n Node) Node {
	n.ID = strings.TrimSpace(n.ID)
	n.Name = strings.TrimSpace(n.Name)
	n.Address = strings.TrimSpace(n.Address)
	if n.Labels == nil {
		n.Labels = map[string]string{}
	}
	if n.TaskTypes == nil {
		n.TaskTypes = []string{}
	}
	sort.Strings(n.TaskTypes)
	return n
}
func (n Node) DisplayName() string {
	if n.Name != "" {
		return n.Name
	}
	return n.ID
}
func (n Node) HasLabel(k string) bool { _, ok := n.Labels[k]; return ok }
func (n Node) Label(k string) string  { return n.Labels[k] }
func (n Node) HasTask(t string) bool {
	for _, x := range n.TaskTypes {
		if x == t {
			return true
		}
	}
	return false
}
func (n Node) LastSeen(now time.Time) time.Duration {
	if n.LastHeartbeatAt == nil {
		return 0
	}
	return now.Sub(*n.LastHeartbeatAt)
}
func (n Node) IsStale(now time.Time, d time.Duration) bool {
	return n.LastHeartbeatAt == nil || n.LastSeen(now) >= d
}
func (n Node) IsLost() bool       { return n.Status == Lost || n.Status == Recovering }
func (n Node) IsAvailable() bool  { return n.Status == Online }
func (n Node) IsRegistered() bool { return n.Status == Registered }
func (n Node) IsOffline() bool    { return n.Status == Offline }
func (n Node) LabelKeys() []string {
	out := make([]string, 0, len(n.Labels))
	for k := range n.Labels {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func (n Node) TaskTypeString() string { return strings.Join(n.TaskTypes, ",") }
func (n Node) LabelString() string {
	parts := make([]string, 0, len(n.Labels))
	for _, k := range n.LabelKeys() {
		parts = append(parts, k+"="+n.Labels[k])
	}
	return strings.Join(parts, ",")
}
func (n Node) Summary(now time.Time) NodeSummary {
	return NodeSummary{ID: n.ID, Name: n.DisplayName(), Status: n.Status, Age: n.LastSeen(now)}
}
func SortByName(ns []Node) []Node {
	out := append([]Node(nil), ns...)
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].DisplayName()) < strings.ToLower(out[j].DisplayName())
	})
	return out
}
func SortByHeartbeat(ns []Node) []Node {
	out := append([]Node(nil), ns...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LastHeartbeatAt == nil {
			return false
		}
		if out[j].LastHeartbeatAt == nil {
			return true
		}
		return out[i].LastHeartbeatAt.After(*out[j].LastHeartbeatAt)
	})
	return out
}
