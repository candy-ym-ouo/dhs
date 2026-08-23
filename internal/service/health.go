package service

import (
	"context"
	"dhs/internal/model"
	"time"
)

type Health struct {
	Status    string    `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
	Nodes     int       `json:"nodes"`
	Online    int       `json:"online"`
}

func (s *Service) Health(ctx context.Context) (Health, error) {
	stats, err := s.Stats(ctx)
	if err != nil {
		return Health{}, err
	}
	online := stats[model.Online]
	total := 0
	for _, n := range stats {
		total += n
	}
	return Health{Status: "ok", CheckedAt: time.Now().UTC(), Nodes: total, Online: online}, nil
}
