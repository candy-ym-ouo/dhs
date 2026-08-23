package heartbeat

import "time"

type Window struct {
	Start time.Time
	End   time.Time
}

func NewWindow(end time.Time, width time.Duration) Window {
	return Window{Start: end.Add(-width), End: end}
}
func (w Window) Contains(t time.Time) bool { return !t.Before(w.Start) && !t.After(w.End) }
func (w Window) Duration() time.Duration   { return w.End.Sub(w.Start) }
func (w Window) Empty() bool               { return !w.End.After(w.Start) }
func (w Window) Overlaps(other Window) bool {
	return w.Start.Before(other.End) && other.Start.Before(w.End)
}
func (w Window) Clamp(t time.Time) time.Time {
	if t.Before(w.Start) {
		return w.Start
	}
	if t.After(w.End) {
		return w.End
	}
	return t
}
func (w Window) Split(parts int) []Window {
	if parts < 1 || w.Empty() {
		return nil
	}
	step := w.Duration() / time.Duration(parts)
	out := make([]Window, 0, parts)
	for i := 0; i < parts; i++ {
		a := w.Start.Add(step * time.Duration(i))
		b := a.Add(step)
		if i == parts-1 {
			b = w.End
		}
		out = append(out, Window{a, b})
	}
	return out
}
