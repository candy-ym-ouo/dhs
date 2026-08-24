package model

import "testing"

func TestAddSourceLabel(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
		want   map[string]string
	}{
		{"nil labels", nil, map[string]string{"source": "heartbeat-api"}},
		{"empty map", map[string]string{}, map[string]string{"source": "heartbeat-api"}},
		{"existing labels", map[string]string{"region": "us"}, map[string]string{"source": "heartbeat-api", "region": "us"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := RegisterRequest{ID: "n1", Name: "node1", Labels: c.labels}
			r.AddSourceLabel()
			if got := r.Labels["source"]; got != "heartbeat-api" {
				t.Fatalf("source label = %q, want %q", got, "heartbeat-api")
			}
			for k, v := range c.want {
				if r.Labels[k] != v {
					t.Fatalf("label %q = %q, want %q", k, r.Labels[k], v)
				}
			}
		})
	}
}
