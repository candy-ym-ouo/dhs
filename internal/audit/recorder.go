package audit

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrClosed    = errors.New("audit recorder is closed")
	ErrEmptyKind = errors.New("audit event kind is required")
)

type Clock func() time.Time

type Recorder struct {
	mu          sync.RWMutex
	clock       Clock
	buffer      []Event
	capacity    int
	start       int
	length      int
	next        uint64
	closed      bool
	subscribers map[uint64]*subscriber
	nextSubID   uint64
}

type subscriber struct {
	filter Filter
	ch     chan Event
}

func NewRecorder(capacity int) *Recorder {
	return NewRecorderWithClock(capacity, time.Now)
}

func NewRecorderWithClock(capacity int, clock Clock) *Recorder {
	if capacity < 1 {
		capacity = 1
	}
	if clock == nil {
		clock = time.Now
	}
	return &Recorder{
		clock:       clock,
		buffer:      make([]Event, capacity),
		capacity:    capacity,
		next:        1,
		subscribers: make(map[uint64]*subscriber),
	}
}

func (r *Recorder) Record(event Event) (Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return Event{}, ErrClosed
	}
	event = event.normalized(r.clock().UTC())
	if event.Kind == "" {
		return Event{}, ErrEmptyKind
	}
	event.Sequence = r.next
	r.next++
	r.appendLocked(event)
	r.publishLocked(event)
	return event.Clone(), nil
}

func (r *Recorder) appendLocked(event Event) {
	if r.length < r.capacity {
		index := (r.start + r.length) % r.capacity
		r.buffer[index] = event.Clone()
		r.length++
		return
	}
	r.buffer[r.start] = event.Clone()
	r.start = (r.start + 1) % r.capacity
}

func (r *Recorder) publishLocked(event Event) {
	for _, sub := range r.subscribers {
		if !sub.filter.Match(event) {
			continue
		}
		select {
		case sub.ch <- event.Clone():
		default:
		}
	}
}

func (r *Recorder) Snapshot() []Event {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.snapshotLocked()
}

func (r *Recorder) Query(filter Filter, limit int) []Event {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit < 0 {
		limit = 0
	}
	result := make([]Event, 0, r.length)
	for i := r.length - 1; i >= 0; i-- {
		index := (r.start + i) % r.capacity
		event := r.buffer[index]
		if filter.Match(event) {
			result = append(result, event.Clone())
			if limit > 0 && len(result) == limit {
				break
			}
		}
	}
	reverseEvents(result)
	return result
}

func (r *Recorder) Since(sequence uint64, limit int) []Event {
	return r.Query(Filter{AfterSequence: sequence}, limit)
}

func (r *Recorder) Subscribe(ctx context.Context, filter Filter, buffer int) <-chan Event {
	if buffer < 1 {
		buffer = 1
	}
	output := make(chan Event, buffer)
	r.mu.Lock()
	if r.closed {
		close(output)
		r.mu.Unlock()
		return output
	}
	id := r.nextSubID
	r.nextSubID++
	r.subscribers[id] = &subscriber{filter: filter, ch: output}
	r.mu.Unlock()
	go r.unsubscribeOnDone(ctx, id)
	return output
}

func (r *Recorder) unsubscribeOnDone(ctx context.Context, id uint64) {
	<-ctx.Done()
	r.mu.Lock()
	sub, ok := r.subscribers[id]
	if ok {
		delete(r.subscribers, id)
		close(sub.ch)
	}
	r.mu.Unlock()
}

func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrClosed
	}
	r.closed = true
	for id, sub := range r.subscribers {
		delete(r.subscribers, id)
		close(sub.ch)
	}
	return nil
}

func (r *Recorder) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.length
}

func (r *Recorder) Capacity() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.capacity
}

func (r *Recorder) Dropped() uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stored := uint64(r.length)
	recorded := r.next - 1
	if recorded <= stored {
		return 0
	}
	return recorded - stored
}

func (r *Recorder) snapshotLocked() []Event {
	events := make([]Event, 0, r.length)
	for i := 0; i < r.length; i++ {
		index := (r.start + i) % r.capacity
		events = append(events, r.buffer[index].Clone())
	}
	return events
}

func reverseEvents(events []Event) {
	for left, right := 0, len(events)-1; left < right; left, right = left+1, right-1 {
		events[left], events[right] = events[right], events[left]
	}
}
