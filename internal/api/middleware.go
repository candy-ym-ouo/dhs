package api

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

const requestStart contextKey = "request-start"

func withTiming(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), requestStart, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func startedAt(ctx context.Context) time.Time {
	if v, ok := ctx.Value(requestStart).(time.Time); ok {
		return v
	}
	return time.Time{}
}
func elapsed(ctx context.Context) time.Duration {
	if t := startedAt(ctx); !t.IsZero() {
		return time.Since(t)
	}
	return 0
}
func withTimeout(next http.Handler, d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func withMethods(next http.Handler, allowed ...string) http.Handler {
	set := map[string]bool{}
	for _, m := range allowed {
		set[m] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !set[r.Method] {
			w.Header().Set("Allow", joinMethods(allowed))
			methodNotAllowed(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func joinMethods(xs []string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += ", "
		}
		out += x
	}
	return out
}
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
