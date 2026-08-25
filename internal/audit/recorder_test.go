package audit

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecorderRetainsNewestEvents(t *testing.T) {
	now := time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC)
	recorder := NewRecorderWithClock(2, func() time.Time { return now })
	for _, kind := range []string{"created", "updated", "removed"} {
		if _, err := recorder.Record(Event{Kind: kind}); err != nil {
			t.Fatal(err)
		}
	}
	events := recorder.Snapshot()
	if len(events) != 2 || events[0].Kind != "updated" || events[1].Kind != "removed" {
		t.Fatalf("unexpected snapshot: %#v", events)
	}
	if recorder.Dropped() != 1 {
		t.Fatalf("dropped = %d, want 1", recorder.Dropped())
	}
}

func TestRecorderCopiesMutableEventData(t *testing.T) {
	recorder := NewRecorder(1)
	labels := map[string]string{"region": "east"}
	recorded, err := recorder.Record(Event{Kind: "heartbeat", Labels: labels})
	if err != nil {
		t.Fatal(err)
	}
	labels["region"] = "west"
	recorded.Labels["region"] = "north"
	if got := recorder.Snapshot()[0].Labels["region"]; got != "east" {
		t.Fatalf("stored label = %q, want east", got)
	}
}

func TestSubscriptionFiltersAndCloses(t *testing.T) {
	recorder := NewRecorder(4)
	ctx, cancel := context.WithCancel(context.Background())
	updates := recorder.Subscribe(ctx, Filter{MinimumLevel: LevelWarn}, 2)
	if _, err := recorder.Record(Event{Kind: "healthy", Level: LevelInfo}); err != nil {
		t.Fatal(err)
	}
	warning, err := recorder.Record(Event{Kind: "late", Level: LevelWarn})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-updates:
		if got.Sequence != warning.Sequence {
			t.Fatalf("sequence = %d, want %d", got.Sequence, warning.Sequence)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for subscription event")
	}
	cancel()
	select {
	case _, ok := <-updates:
		if ok {
			t.Fatal("subscription remained open")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for subscription close")
	}
}

func TestQuerySummaryAndJSONLines(t *testing.T) {
	recorder := NewRecorder(4)
	for _, event := range []Event{
		{Kind: "probe", Source: "node-a", Level: LevelInfo, Values: map[string]int64{"attempts": 1}},
		{Kind: "probe", Source: "node-b", Level: LevelError, Values: map[string]int64{"attempts": 2}},
	} {
		if _, err := recorder.Record(event); err != nil {
			t.Fatal(err)
		}
	}
	errorsOnly := recorder.Query(Filter{MinimumLevel: LevelError}, 0)
	if len(errorsOnly) != 1 || errorsOnly[0].Source != "node-b" {
		t.Fatalf("unexpected query: %#v", errorsOnly)
	}
	summary := Summarize(recorder.Snapshot())
	if summary.Total != 2 || len(summary.ValueTotals) != 1 || summary.ValueTotals[0].Count != 3 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	var encoded bytes.Buffer
	if err := WriteJSONLines(&encoded, recorder.Snapshot()); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadJSONLines(&encoded, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 2 || decoded[1].Sequence != 2 {
		t.Fatalf("unexpected decoded events: %#v", decoded)
	}
}

func TestRecordValidationAndClose(t *testing.T) {
	recorder := NewRecorder(1)
	if _, err := recorder.Record(Event{}); !errors.Is(err, ErrEmptyKind) {
		t.Fatalf("error = %v, want ErrEmptyKind", err)
	}
	if err := recorder.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(Event{Kind: "late"}); !errors.Is(err, ErrClosed) {
		t.Fatalf("error = %v, want ErrClosed", err)
	}
}
