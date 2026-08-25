package audit

import (
	"strings"
	"time"
)

type Filter struct {
	MinimumLevel  Level
	Kinds         []string
	Sources       []string
	Subjects      []string
	Labels        map[string]string
	Search        string
	After         time.Time
	Before        time.Time
	AfterSequence uint64
}

func (f Filter) Match(event Event) bool {
	if f.MinimumLevel.Valid() && event.Level.Rank() < f.MinimumLevel.Rank() {
		return false
	}
	if f.AfterSequence != 0 && event.Sequence <= f.AfterSequence {
		return false
	}
	if !f.After.IsZero() && event.OccurredAt.Before(f.After) {
		return false
	}
	if !f.Before.IsZero() && !event.OccurredAt.Before(f.Before) {
		return false
	}
	if !containsFold(f.Kinds, event.Kind) {
		return false
	}
	if !containsFold(f.Sources, event.Source) {
		return false
	}
	if !containsFold(f.Subjects, event.Subject) {
		return false
	}
	for key, expected := range f.Labels {
		actual, ok := event.Labels[key]
		if !ok || !strings.EqualFold(actual, expected) {
			return false
		}
	}
	return matchesSearch(event, f.Search)
}

func (f Filter) Apply(events []Event) []Event {
	matched := make([]Event, 0, len(events))
	for _, event := range events {
		if f.Match(event) {
			matched = append(matched, event.Clone())
		}
	}
	return matched
}

func containsFold(allowed []string, value string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if strings.EqualFold(strings.TrimSpace(candidate), value) {
			return true
		}
	}
	return false
}

func matchesSearch(event Event, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	fields := []string{event.Kind, event.Source, event.Subject, event.Message}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	for key, value := range event.Labels {
		if strings.Contains(strings.ToLower(key), query) || strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}
