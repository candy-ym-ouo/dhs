package service

import (
	"dhs/internal/model"
	"sort"
)

func (r *Registry) AddTask(id, task string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.nodes[id]
	if !ok {
		return false
	}
	if n.HasTask(task) {
		return true
	}
	n.TaskTypes = append([]string(nil), n.TaskTypes...)
	n.TaskTypes = append(n.TaskTypes, task)
	sort.Strings(n.TaskTypes)
	r.nodes[id] = n
	return true
}

func (r *Registry) Tasks(id string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.nodes[id]
	if !ok {
		return nil
	}
	return append([]string(nil), n.TaskTypes...)
}

var _ = model.Online
