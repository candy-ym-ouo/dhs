package scanner

import (
	"sync"
	"testing"
	"time"
)

func TestBug004RoundMarkerIsSafeAcrossWorkers(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ { wg.Add(1); go func(i int) { defer wg.Done(); markRound(time.Unix(int64(i), 0)) }(i) }
	wg.Wait()
	if lastRoundAt().IsZero() { t.Fatalf("round marker was not updated") }
}
