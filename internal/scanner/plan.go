package scanner

import (
	"dhs/internal/model"
	"time"
)

type Action string

const (
	ActionLost    Action = "lost"
	ActionRecover Action = "recover"
	ActionRetire  Action = "retire"
	ActionNone    Action = "none"
)

type Decision struct {
	NodeID string
	Action Action
	Reason string
}

func Decide(n model.Node, now time.Time, timeout, maxLost time.Duration) Decision {
	if n.Status == model.Offline {
		return Decision{NodeID: n.ID, Action: ActionNone}
	}
	if n.LostAt != nil && now.Sub(*n.LostAt) >= maxLost {
		return Decision{n.ID, ActionRetire, "max_lost_duration"}
	}
	if n.LastHeartbeatAt == nil || now.Sub(*n.LastHeartbeatAt) >= timeout {
		return Decision{n.ID, ActionLost, "heartbeat_timeout"}
	}
	return Decision{n.ID, ActionNone, "healthy"}
}
func CountActions(ds []Decision) map[Action]int {
	m := map[Action]int{}
	for _, d := range ds {
		m[d.Action]++
	}
	return m
}
