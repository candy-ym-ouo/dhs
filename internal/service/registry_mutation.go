package service

import "dhs/internal/model"

func (r *Registry) AddTask(id, task string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.nodes[id]
	if !ok {
		return false
	}
	n.TaskTypes = append(n.TaskTypes, task)
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
	return n.TaskTypes
}

var _ = model.Online
