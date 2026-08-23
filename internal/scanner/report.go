package scanner

import (
	"fmt"
	"strings"
	"time"
)

type Report struct {
	Round                    int64
	Started, Finished        time.Time
	Lost, Recovered, Retired int
	Errors                   []string
}

func (r Report) Duration() time.Duration { return r.Finished.Sub(r.Started) }
func (r Report) Successful() bool        { return len(r.Errors) == 0 }
func (r Report) AddError(e error) {
	if e != nil {
		r.Errors = append(r.Errors, e.Error())
	}
}
func (r Report) Summary() string {
	status := "ok"
	if !r.Successful() {
		status = "error"
	}
	return fmt.Sprintf("round=%d status=%s duration=%s lost=%d recovered=%d retired=%d", r.Round, status, r.Duration(), r.Lost, r.Recovered, r.Retired)
}
func (r Report) Lines() []string {
	out := []string{r.Summary()}
	for _, e := range r.Errors {
		out = append(out, "error: "+e)
	}
	return out
}
func (r Report) Text() string { return strings.Join(r.Lines(), "\n") }
func (r Report) Counters() map[string]int {
	return map[string]int{"lost": r.Lost, "recovered": r.Recovered, "retired": r.Retired, "errors": len(r.Errors)}
}
func (r Report) StartedAt() string  { return r.Started.UTC().Format(time.RFC3339Nano) }
func (r Report) FinishedAt() string { return r.Finished.UTC().Format(time.RFC3339Nano) }
