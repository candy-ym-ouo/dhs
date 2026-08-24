package service

import (
	"dhs/internal/model"
	"strings"
)

// AddTask appends a task type to a cached node. It holds the write lock for
// the whole read-modify-write so a concurrent Tasks/AddTask cannot interleave.
// The node is cloned before mutation, so the appended element lands in a fresh
// backing array rather than one shared with a prior Get/Snapshot/List clone —
// an append into spare capacity of a shared array would otherwise be visible
// through every outstanding snapshot.
//
// Validation mirrors RegisterRequest.Validate so the cache never diverges
// from what the store layer would accept on re-registration: trimmed, non-empty
// task name, no duplicates, and a 100-entry ceiling.
func (r *Registry) AddTask(id, task string) bool {
	task = strings.TrimSpace(task)
	if task == "" {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.nodes[id]
	if !ok {
		return false
	}
	if n.HasTask(task) {
		return false
	}
	if len(n.TaskTypes) >= 100 {
		return false
	}
	n = n.Clone()
	n.TaskTypes = append(n.TaskTypes, task)
	r.nodes[id] = n
	return true
}

// Tasks returns a defensive copy of a node's task types. It must never expose
// the slice header the registry itself holds: returning it directly lets a
// caller append into spare capacity and write through into the cached node and
// every previously-returned snapshot. The copy is taken under the read lock so
// it is consistent with the registry's view at return time.
func (r *Registry) Tasks(id string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.nodes[id]
	if !ok {
		return nil
	}
	if n.TaskTypes == nil {
		return nil
	}
	return append([]string(nil), n.TaskTypes...)
}

var _ = model.Online
