package api

import (
	"dhs/internal/model"
	"dhs/internal/service"
	"dhs/internal/store"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type API struct{ S *service.Service }

func (a *API) Router(static http.Handler) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { write(w, map[string]any{"status": "ok"}) })
	m.Handle("GET /", static)
	m.HandleFunc("POST /api/v1/nodes/register", a.register)
	m.HandleFunc("POST /api/v1/nodes/", a.nodeAction)
	m.HandleFunc("GET /api/v1/nodes", a.list)
	m.HandleFunc("GET /api/v1/nodes/", a.detail)
	m.HandleFunc("GET /api/v1/stats", a.stats)
	return m
}
func write(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": v})
}
func fail(w http.ResponseWriter, c int, e error) { http.Error(w, e.Error(), c) }
func statusFor(e error) int {
	switch {
	case errors.Is(e, service.ErrNotFound) || errors.Is(e, store.ErrNotFound):
		return 404
	case errors.Is(e, service.ErrConflict) || errors.Is(e, store.ErrConflict):
		return 409
	default:
		return 500
	}
}
func (a *API) register(w http.ResponseWriter, r *http.Request) {
	var x model.RegisterRequest
	if json.NewDecoder(r.Body).Decode(&x) != nil {
		fail(w, 400, service.ErrConflict)
		return
	}
	n, e := a.S.Register(r.Context(), x)
	if e != nil {
		fail(w, 400, e)
		return
	}
	write(w, n)
}
func (a *API) nodeAction(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(p) < 5 {
		fail(w, 404, service.ErrNotFound)
		return
	}
	id := p[3]
	switch p[4] {
	case "heartbeat":
		var x model.HeartbeatRequest
		_ = json.NewDecoder(r.Body).Decode(&x)
		h, c, e := a.S.Heartbeat(r.Context(), id, x)
		if e != nil {
			fail(w, 400, e)
			return
		}
		write(w, map[string]any{"heartbeat": h, "changed": c})
	case "recover":
		c, e := a.S.Recover(r.Context(), id)
		if e != nil {
			fail(w, 409, e)
			return
		}
		write(w, map[string]any{"changed": c})
	case "retire":
		c, e := a.S.Retire(r.Context(), id)
		if e != nil {
			fail(w, 409, e)
			return
		}
		write(w, map[string]any{"changed": c})
	default:
		fail(w, 404, service.ErrNotFound)
	}
}
func (a *API) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p, _ := strconv.Atoi(q.Get("page"))
	ps, _ := strconv.Atoi(q.Get("page_size"))
	st, err := model.ParseStatus(q.Get("status"))
	if q.Get("status") != "" && err != nil {
		fail(w, 400, err)
		return
	}
	ns, total, e := a.S.List(r.Context(), store.Filter{Status: st, Keyword: q.Get("keyword"), Page: p, PageSize: ps})
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, map[string]any{"items": ns, "total": total, "page": p, "page_size": ps})
}
func (a *API) stats(w http.ResponseWriter, r *http.Request) {
	m, e := a.S.Stats(r.Context())
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, map[string]any{"by_status": m})
}
func (a *API) detail(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(p) < 4 {
		fail(w, 404, service.ErrNotFound)
		return
	}
	id := p[3]
	if len(p) == 4 {
		n, e := a.S.Node(r.Context(), id)
		if e != nil {
			fail(w, statusFor(e), e)
			return
		}
		write(w, n)
		return
	}
	l, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if p[4] == "transitions" {
		v, e := a.S.Transitions(r.Context(), id, l)
		if e != nil {
			fail(w, statusFor(e), e)
			return
		}
		write(w, v)
	} else if p[4] == "heartbeats" {
		v, e := a.S.Heartbeats(r.Context(), id, l)
		if e != nil {
			fail(w, statusFor(e), e)
			return
		}
		write(w, v)
	} else {
		fail(w, 404, service.ErrNotFound)
	}
}
