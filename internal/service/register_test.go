package service

import (
	"context"
	"testing"

	"dhs/internal/model"
	"dhs/internal/store/sqlite"
)

// TestRegisterNoLabelsCrash reproduces the runtime panic where AddSourceLabel
// wrote to a nil map when labels were omitted, and asserts that registration
// still persists the node with status REGISTERED.
func TestRegisterNoLabelsCrash(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
	}{
		{"nil labels", nil},
		{"empty map", map[string]string{}},
		{"existing labels", map[string]string{"region": "us"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, e := sqlite.Open(":memory:")
			if e != nil {
				t.Fatalf("open: %v", e)
			}
			defer db.Close()
			s := &Service{Store: db}

			req := model.RegisterRequest{ID: "n1", Name: "node1", Labels: c.labels}
			n, e := s.Register(context.Background(), req)
			if e != nil {
				t.Fatalf("register: %v", e)
			}
			if n.Status != model.Registered {
				t.Fatalf("status = %q, want REGISTERED", n.Status)
			}
			if n.Labels["source"] != "heartbeat-api" {
				t.Fatalf("source label missing = %q", n.Labels["source"])
			}

			// Persistence: reload from the durable store and re-check.
			got, e := s.Node(context.Background(), "n1")
			if e != nil {
				t.Fatalf("get: %v", e)
			}
			if got.Status != model.Registered {
				t.Fatalf("persisted status = %q, want REGISTERED", got.Status)
			}
			if got.Labels["source"] != "heartbeat-api" {
				t.Fatalf("persisted source label = %q", got.Labels["source"])
			}
		})
	}
}
