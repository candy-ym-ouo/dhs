package model

import (
	"fmt"
	"time"
)

type Status string

const (
	Registered Status = "REGISTERED"
	Online     Status = "ONLINE"
	Lost       Status = "LOST"
	Recovering Status = "RECOVERING"
	Offline    Status = "OFFLINE"
)

func (s Status) Valid() bool {
	return s == Registered || s == Online || s == Lost || s == Recovering || s == Offline
}
func ParseStatus(v string) (Status, error) {
	s := Status(v)
	if !s.Valid() {
		return "", fmt.Errorf("invalid status %q", v)
	}
	return s, nil
}

type Node struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Address         string            `json:"address"`
	Labels          map[string]string `json:"labels,omitempty"`
	TaskTypes       []string          `json:"task_types,omitempty"`
	Version         string            `json:"version,omitempty"`
	Status          Status            `json:"status"`
	LastHeartbeatAt *time.Time        `json:"last_heartbeat_at,omitempty"`
	RegisteredAt    time.Time         `json:"registered_at"`
	LostAt          *time.Time        `json:"lost_at,omitempty"`
	RecoverAttempts int               `json:"recover_attempts"`
}

type Heartbeat struct {
	ID         int64          `json:"id"`
	NodeID     string         `json:"node_id"`
	Load       float64        `json:"load"`
	Extra      map[string]any `json:"extra,omitempty"`
	ReportedAt time.Time      `json:"reported_at"`
}
type Transition struct {
	ID                      int64  `json:"id"`
	NodeID                  string `json:"node_id"`
	From                    Status `json:"from_status"`
	To                      Status `json:"to_status"`
	Reason, Trigger, Detail string
	CreatedAt               time.Time `json:"created_at"`
}
type RegisterRequest struct {
	ID, Name, Address string
	Labels            map[string]string
	TaskTypes         []string
	Version           string
}
type HeartbeatRequest struct {
	Load  float64
	Extra map[string]any
}
