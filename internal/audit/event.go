package audit

import (
	"maps"
	"strings"
	"time"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

func (l Level) Valid() bool {
	switch l {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return true
	default:
		return false
	}
}

func (l Level) Rank() int {
	switch l {
	case LevelDebug:
		return 0
	case LevelInfo:
		return 1
	case LevelWarn:
		return 2
	case LevelError:
		return 3
	default:
		return -1
	}
}

type Event struct {
	Sequence   uint64            `json:"sequence"`
	OccurredAt time.Time         `json:"occurred_at"`
	Level      Level             `json:"level"`
	Kind       string            `json:"kind"`
	Source     string            `json:"source,omitempty"`
	Subject    string            `json:"subject,omitempty"`
	Message    string            `json:"message,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Values     map[string]int64  `json:"values,omitempty"`
}

func (e Event) Clone() Event {
	e.Labels = maps.Clone(e.Labels)
	e.Values = maps.Clone(e.Values)
	return e
}

func (e Event) normalized(now time.Time) Event {
	e.Kind = strings.TrimSpace(e.Kind)
	e.Source = strings.TrimSpace(e.Source)
	e.Subject = strings.TrimSpace(e.Subject)
	e.Message = strings.TrimSpace(e.Message)
	if e.OccurredAt.IsZero() {
		e.OccurredAt = now
	} else {
		e.OccurredAt = e.OccurredAt.UTC()
	}
	if !e.Level.Valid() {
		e.Level = LevelInfo
	}
	e.Labels = cleanLabels(e.Labels)
	e.Values = maps.Clone(e.Values)
	return e
}

func cleanLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	cleaned := make(map[string]string, len(labels))
	for key, value := range labels {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			cleaned[key] = value
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}
