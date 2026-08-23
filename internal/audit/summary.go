package audit

import (
	"cmp"
	"slices"
	"time"
)

type Count struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Summary struct {
	Total       int       `json:"total"`
	First       time.Time `json:"first,omitempty"`
	Last        time.Time `json:"last,omitempty"`
	ByLevel     []Count   `json:"by_level,omitempty"`
	ByKind      []Count   `json:"by_kind,omitempty"`
	BySource    []Count   `json:"by_source,omitempty"`
	ValueTotals []Count   `json:"value_totals,omitempty"`
}

func Summarize(events []Event) Summary {
	result := Summary{Total: len(events)}
	if len(events) == 0 {
		return result
	}
	levels := make(map[string]int)
	kinds := make(map[string]int)
	sources := make(map[string]int)
	values := make(map[string]int64)
	for i, event := range events {
		if i == 0 || event.OccurredAt.Before(result.First) {
			result.First = event.OccurredAt
		}
		if i == 0 || event.OccurredAt.After(result.Last) {
			result.Last = event.OccurredAt
		}
		levels[string(event.Level)]++
		kinds[event.Kind]++
		if event.Source != "" {
			sources[event.Source]++
		}
		for name, value := range event.Values {
			values[name] += value
		}
	}
	result.ByLevel = sortedCounts(levels)
	result.ByKind = sortedCounts(kinds)
	result.BySource = sortedCounts(sources)
	result.ValueTotals = sortedValueCounts(values)
	return result
}

func sortedCounts(source map[string]int) []Count {
	result := make([]Count, 0, len(source))
	for name, count := range source {
		result = append(result, Count{Name: name, Count: count})
	}
	slices.SortFunc(result, func(left, right Count) int {
		if order := cmp.Compare(right.Count, left.Count); order != 0 {
			return order
		}
		return cmp.Compare(left.Name, right.Name)
	})
	return result
}

func sortedValueCounts(source map[string]int64) []Count {
	result := make([]Count, 0, len(source))
	for name, count := range source {
		result = append(result, Count{Name: name, Count: clampInt(count)})
	}
	slices.SortFunc(result, func(left, right Count) int {
		if order := cmp.Compare(right.Count, left.Count); order != 0 {
			return order
		}
		return cmp.Compare(left.Name, right.Name)
	})
	return result
}

func clampInt(value int64) int {
	converted := int(value)
	if int64(converted) == value {
		return converted
	}
	if value < 0 {
		return -int(^uint(0)>>1) - 1
	}
	return int(^uint(0) >> 1)
}
