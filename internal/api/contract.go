package api

import (
	"dhs/internal/model"
	"encoding/json"
	"net/http"
	"time"
)

type Envelope struct {
	Code  int            `json:"code"`
	Data  any            `json:"data,omitempty"`
	Error *ErrorBody     `json:"error,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
}
type ListData[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Pages    int `json:"pages"`
}
type RegisterResponse struct {
	Node    model.Node `json:"node"`
	Created bool       `json:"created"`
}
type HeartbeatResponse struct {
	NodeID     string       `json:"node_id"`
	Status     model.Status `json:"status"`
	Changed    bool         `json:"changed"`
	ReceivedAt time.Time    `json:"received_at"`
}
type TransitionResponse struct {
	Items []model.Transition `json:"items"`
	Total int                `json:"total"`
}
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version,omitempty"`
	Uptime    string    `json:"uptime,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}
type StatsResponse struct {
	Total     int                  `json:"total"`
	ByStatus  map[model.Status]int `json:"by_status"`
	CheckedAt time.Time            `json:"checked_at"`
}

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Code: 0, Data: v})
}
func respondCreated(w http.ResponseWriter, v any) { respond(w, http.StatusCreated, v) }
func respondNoContent(w http.ResponseWriter)      { w.WriteHeader(http.StatusNoContent) }
func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func encode(v any) ([]byte, error) { return json.Marshal(v) }
func totalStats(m map[model.Status]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
func statusLabel(s model.Status) string {
	switch s {
	case model.Online:
		return "online"
	case model.Lost:
		return "lost"
	case model.Recovering:
		return "recovering"
	case model.Registered:
		return "registered"
	case model.Offline:
		return "offline"
	}
	return "unknown"
}
func statusLabels(m map[model.Status]int) map[string]int {
	out := map[string]int{}
	for s, n := range m {
		out[statusLabel(s)] = n
	}
	return out
}
func requestID(r *http.Request) string {
	if v := r.Header.Get("X-Request-ID"); v != "" {
		return v
	}
	return time.Now().UTC().Format("20060102150405.000000000")
}
