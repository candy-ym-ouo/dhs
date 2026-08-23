package service

import (
	"dhs/internal/model"
	"time"
)

type Policy struct {
	Timeout time.Duration
	Grace   time.Duration
	MaxLost time.Duration
	Confirm int
}

func DefaultPolicy() Policy {
	return Policy{Timeout: 30 * time.Second, Grace: 60 * time.Second, MaxLost: 30 * time.Minute, Confirm: 1}
}
func (p Policy) Valid() bool {
	return p.Timeout > 0 && p.Grace > 0 && p.MaxLost >= p.Grace && p.Confirm > 0
}
func (p Policy) IsTimedOut(last, now time.Time) bool {
	return last.IsZero() || now.Sub(last) >= p.Timeout
}
func (p Policy) ShouldRetire(lostAt, now time.Time) bool {
	return !lostAt.IsZero() && now.Sub(lostAt) >= p.MaxLost
}
func StatusChanged(before, after model.Status) bool { return before != after }
