package service

import (
	"context"
	"testing"
)

func TestBug002CancelledContextReachesDownstream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := downstreamContext(ctx)
	if got.Err() == nil { t.Fatalf("cancelled context was replaced or lost") }
}
