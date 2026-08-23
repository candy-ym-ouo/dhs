package model

import "time"

type NodeSummary struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Status Status        `json:"status"`
	Age    time.Duration `json:"age"`
}
type Stats struct {
	Total       int            `json:"total"`
	ByStatus    map[Status]int `json:"by_status"`
	GeneratedAt time.Time      `json:"generated_at"`
}

func BuildStats(nodes []Node, now time.Time) Stats {
	s := Stats{ByStatus: map[Status]int{}, GeneratedAt: now.UTC()}
	for _, n := range nodes {
		s.Total++
		s.ByStatus[n.Status]++
	}
	return s
}
