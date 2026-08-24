package model

import "testing"

func TestBug006RegisterMetadataHandlesNilLabels(t *testing.T) {
	r := RegisterRequest{}
	defer func() {
		if recover() != nil { t.Fatalf("register metadata must handle nil labels") }
	}()
	r.AddSourceLabel()
	if r.Labels["source"] != "heartbeat-api" { t.Fatalf("source label was not stored") }
}
