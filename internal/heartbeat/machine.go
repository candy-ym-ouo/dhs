package heartbeat

import (
	"dhs/internal/model"
	"fmt"
	"time"
)

type Machine struct {
	Status   model.Status
	Since    time.Time
	Attempts int
}

func NewMachine(status model.Status, now time.Time) Machine {
	return Machine{Status: status, Since: now.UTC()}
}
func (m Machine) Age(now time.Time) time.Duration { return now.Sub(m.Since) }
func (m Machine) Healthy() bool                   { return m.Status == model.Online }
func (m Machine) NeedsHeartbeat(now time.Time, timeout time.Duration) bool {
	return !m.Healthy() || m.Age(now) >= timeout
}
func (m Machine) String() string {
	return fmt.Sprintf("%s since %s attempts=%d", m.Status, m.Since.Format(time.RFC3339), m.Attempts)
}
func (m Machine) WithStatus(next model.Status, now time.Time) (Machine, error) {
	if !CanTransition(m.Status, next) {
		return m, ErrIllegalTransition{m.Status, next}
	}
	m.Status = next
	m.Since = now.UTC()
	return m, nil
}
func (m Machine) Register(now time.Time) (Machine, error) { return m.WithStatus(model.Registered, now) }
func (m Machine) Online(now time.Time) (Machine, error)   { return m.WithStatus(model.Online, now) }
func (m Machine) Lost(now time.Time) (Machine, error)     { return m.WithStatus(model.Lost, now) }
func (m Machine) Recover(now time.Time) (Machine, error) {
	x, e := m.WithStatus(model.Recovering, now)
	if e == nil {
		x.Attempts++
	}
	return x, e
}
func (m Machine) Offline(now time.Time) (Machine, error) { return m.WithStatus(model.Offline, now) }
func (m Machine) Can(next model.Status) bool             { return CanTransition(m.Status, next) }
func (m Machine) Event(next model.Status, reason, trigger string, now time.Time) (model.Transition, error) {
	if !m.Can(next) {
		return model.Transition{}, ErrIllegalTransition{m.Status, next}
	}
	if reason == "" || trigger == "" {
		return model.Transition{}, ErrMissingTransitionMetadata{}
	}
	return model.Transition{From: m.Status, To: next, Reason: reason, Trigger: trigger, CreatedAt: now.UTC()}, nil
}
func (m Machine) Recoverable() bool { return IsRecoverable(m.Status) }
func (m Machine) Terminal() bool    { return IsTerminal(m.Status) }
func (m Machine) Operational() bool { return IsOperational(m.Status) }
