package store

import (
	"context"
	"dhs/internal/model"
	"time"
)

type Filter struct {
	Status         model.Status
	Keyword        string
	Page, PageSize int
}
type Store interface {
	Close() error
	Register(context.Context, model.RegisterRequest) (model.Node, error)
	GetNode(context.Context, string) (model.Node, error)
	ListNodes(context.Context, Filter) ([]model.Node, int, error)
	RecordHeartbeat(context.Context, string, model.HeartbeatRequest, time.Time) (model.Heartbeat, error)
	SetStatus(context.Context, string, model.Status, string, string, string, time.Time) (bool, error)
	StartRecovery(context.Context, string, time.Time) (bool, error)
	ListTransitions(context.Context, string, int) ([]model.Transition, error)
	ListHeartbeats(context.Context, string, int) ([]model.Heartbeat, error)
	Stats(context.Context) (map[model.Status]int, error)
	ScanCandidates(context.Context, time.Time, time.Duration) ([]model.Node, error)
	Cleanup(context.Context, time.Time) error
}
