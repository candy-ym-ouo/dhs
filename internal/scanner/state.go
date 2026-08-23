package scanner

import "time"

var lastRound time.Time

func markRound(now time.Time) { lastRound = now }
func lastRoundAt() time.Time  { return lastRound }
