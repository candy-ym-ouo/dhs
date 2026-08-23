package store

import (
	"dhs/internal/model"
	"sort"
	"strings"
)

func Match(n model.Node, f Filter) bool {
	if f.Status != "" && n.Status != f.Status {
		return false
	}
	if f.Keyword == "" {
		return true
	}
	q := strings.ToLower(f.Keyword)
	return strings.Contains(strings.ToLower(n.ID), q) || strings.Contains(strings.ToLower(n.Name), q) || strings.Contains(strings.ToLower(n.Address), q)
}
func SortNodes(nodes []model.Node) []model.Node {
	out := append([]model.Node(nil), nodes...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func SliceNodes(nodes []model.Node, p Page) []model.Node {
	start := p.Offset()
	if start >= len(nodes) {
		return []model.Node{}
	}
	end := start + p.Size
	if end > len(nodes) {
		end = len(nodes)
	}
	return nodes[start:end]
}
func StatusOrder(s model.Status) int {
	switch s {
	case model.Online:
		return 1
	case model.Recovering:
		return 2
	case model.Lost:
		return 3
	case model.Registered:
		return 4
	case model.Offline:
		return 5
	}
	return 99
}
