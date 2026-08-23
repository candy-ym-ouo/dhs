package scanner

import (
	"context"
	"dhs/internal/config"
	"dhs/internal/model"
	"dhs/internal/service"
	"dhs/internal/store"
	"log/slog"
	"time"
)

type Scanner struct {
	Store   store.Store
	Service *service.Service
	Config  config.Config
	Log     *slog.Logger
}

func (s *Scanner) Run(ctx context.Context) {
	t := time.NewTicker(s.Config.ScanInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			s.Round(ctx, now)
		}
	}
}
func (s *Scanner) Round(ctx context.Context, now time.Time) {
	now = now.UTC()
	ns, e := s.Store.ScanCandidates(ctx, now, s.Config.HeartbeatTimeout)
	if e != nil {
		return
	}
	for _, n := range ns {
		if n.Status == model.Registered || n.Status == model.Online {
			_, _ = s.Store.SetStatus(ctx, n.ID, model.Lost, "heartbeat_timeout", "scanner", "", now)
			_, _ = s.Store.StartRecovery(ctx, n.ID, now)
		}
		if n.Status == model.Recovering && n.LostAt != nil && now.Sub(*n.LostAt) > s.Config.MaxLostDuration {
			_, _ = s.Store.SetStatus(ctx, n.ID, model.Offline, "retired", "scanner", "", now)
		} else if n.Status == model.Recovering && n.LostAt != nil && now.Sub(*n.LostAt) >= s.Config.RecoveryGrace {
			_, _ = s.Store.SetStatus(ctx, n.ID, model.Lost, "recovery_failed", "scanner", "", now)
		}
	}
	_ = s.Store.Cleanup(ctx, now.Add(-s.Config.Retention))
}
