package main

import (
	"context"
	"dhs/internal/api"
	"dhs/internal/config"
	"dhs/internal/scanner"
	"dhs/internal/service"
	"dhs/internal/store/sqlite"
	"dhs/internal/web"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, e := config.FromFlags()
	if e != nil {
		panic(e)
	}
	if e = cfg.Validate(); e != nil {
		panic(e)
	}
	db, e := sqlite.Open(cfg.Database)
	if e != nil {
		panic(e)
	}
	defer db.Close()
	svc := &service.Service{Store: db, ConfirmCount: cfg.ConfirmCount}
	a := &api.API{S: svc}
	srv := &http.Server{Addr: cfg.Listen, Handler: a.Router(web.Handler())}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	sc := &scanner.Scanner{Store: db, Service: svc, Config: cfg, Log: slog.Default()}
	go sc.Run(ctx)
	go func() {
		slog.Info("heartbeat service listening", "addr", cfg.Listen)
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			slog.Error("server stopped", "error", e)
			cancel()
		}
	}()
	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
}
