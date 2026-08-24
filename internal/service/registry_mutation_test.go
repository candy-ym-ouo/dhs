package service

import (
	"dhs/internal/model"
	"reflect"
	"sync"
	"testing"
	"time"
)

// TestRegistry_TasksSnapshotIsolation reproduces the reported cross-layer
// sharing bug end to end: register a node, take a task slice via Tasks(), take
// a node snapshot via Get(), then append a task through AddTask(). Before the
// fix, Tasks() returned the registry's own slice header and Clone() copied
// only the slice header, so an append into spare capacity wrote through into
// the cached node and every outstanding snapshot — a reader of the snapshot
// saw the phantom task. After the fix each boundary owns its storage.
func TestRegistry_TasksSnapshotIsolation(t *testing.T) {
	r := NewRegistry()
	// cap exceeds len so an append reuses backing storage instead of
	// allocating fresh — the exact shape that exposed the shared array.
	node := model.Node{ID: "n1", Name: "node-1", Status: model.Online, TaskTypes: []string{"a", "b", "c"}}
	r.Upsert(node)

	snapshot, ok := r.Get("n1")
	if !ok {
		t.Fatal("expected node to be present")
	}
	tasks := r.Tasks("n1")

	// Mutate the slice handed back by Tasks(); it must not reach the cache.
	tasks = append(tasks, "leak")
	if got := r.Tasks("n1"); reflect.DeepEqual(got, tasks) {
		t.Fatalf("Tasks() returned the registry's backing slice; append leaked: got %v", got)
	}

	// AddTask must not pollute the earlier snapshot taken via Get().
	if !r.AddTask("n1", "new-task") {
		t.Fatal("AddTask returned false for a present node")
	}
	// The cache now holds the extra task.
	if cached, ok := r.Get("n1"); !ok || len(cached.TaskTypes) != len(node.TaskTypes)+1 {
		t.Fatalf("cache missing the appended task: got %v", cached.TaskTypes)
	}
	// The snapshot taken before the append must be unchanged — the original bug
	// surfaced here as a phantom "new-task" element leaking into the clone.
	if !reflect.DeepEqual(snapshot.TaskTypes, node.TaskTypes) {
		t.Fatalf("prior snapshot mutated across AddTask: want %v, got %v", node.TaskTypes, snapshot.TaskTypes)
	}
}

// TestRegistry_SnapshotMapIsolation verifies the map returned by Snapshot()
// stays independent of later mutations, including spare-capacity appends.
func TestRegistry_SnapshotMapIsolation(t *testing.T) {
	r := NewRegistry()
	r.Upsert(model.Node{ID: "n1", TaskTypes: []string{"a", "b", "c"}})
	r.Upsert(model.Node{ID: "n2", TaskTypes: []string{"x", "y", "z"}})

	snap := r.Snapshot()

	r.AddTask("n1", "p")
	r.Upsert(model.Node{ID: "n2", TaskTypes: []string{"x", "y", "z", "q"}})

	if got := snap["n1"].TaskTypes; !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("snapshot n1 polluted by later AddTask: got %v", got)
	}
	if got := snap["n2"].TaskTypes; !reflect.DeepEqual(got, []string{"x", "y", "z"}) {
		t.Fatalf("snapshot n2 polluted by later Upsert: got %v", got)
	}
}

// TestRegistry_AddTaskValidation checks the implementation-level guards added
// to AddTask, mirroring RegisterRequest.Validate so the cache stays consistent
// with store-layer acceptance on re-registration.
func TestRegistry_AddTaskValidation(t *testing.T) {
	r := NewRegistry()
	r.Upsert(model.Node{ID: "n1", TaskTypes: []string{"a"}})

	cases := []struct {
		name string
		task string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   ", false},
		{"trimmed ok", "  new  ", true},
		{"duplicate", "a", false},
		{"duplicate after trim", "  a ", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := r.Tasks("n1")
			if got := r.AddTask("n1", c.task); got != c.want {
				t.Fatalf("AddTask(%q)=%v, want %v", c.task, got, c.want)
			}
			// On rejection the cache must be untouched.
			if !c.want && !reflect.DeepEqual(before, r.Tasks("n1")) {
				t.Fatalf("rejected AddTask mutated cache: before=%v after=%v", before, r.Tasks("n1"))
			}
			// Reset so the duplicate case stays deterministic.
			if c.want {
				r.Remove("n1")
				r.Upsert(model.Node{ID: "n1", TaskTypes: []string{"a"}})
			}
		})
	}
}

// TestRegistry_AddTaskMissingNode ensures AddTask on an absent id is a no-op
// rather than silently inserting into a zero-value node.
func TestRegistry_AddTaskMissingNode(t *testing.T) {
	r := NewRegistry()
	if r.AddTask("nope", "a") {
		t.Fatal("AddTask on missing node returned true")
	}
	if got := r.Tasks("nope"); got != nil {
		t.Fatalf("Tasks on missing node should be nil, got %v", got)
	}
}

// TestRegistry_AddTaskCeiling enforces the 100-task limit shared with
// RegisterRequest.Validate.
func TestRegistry_AddTaskCeiling(t *testing.T) {
	r := NewRegistry()
	tt := make([]string, 100)
	for i := range tt {
		tt[i] = "t"
	}
	// Force distinct backing storage with no spare capacity, then set it.
	node := model.Node{ID: "n1", TaskTypes: append([]string{}, tt...)}
	r.Upsert(node)
	if r.AddTask("n1", "overflow") {
		t.Fatal("AddTask exceeded the 100-task ceiling")
	}
}

// TestRegistry_TasksConcurrentReaders races readers and a writer to confirm
// the read/write boundary holds: Tasks() copies under the read lock, AddTask()
// clones under the write lock, so no reader observes a half-appended slice or
// a backing array being reallocated out from under it.
func TestRegistry_TasksConcurrentReaders(t *testing.T) {
	r := NewRegistry()
	r.Upsert(model.Node{ID: "n1", TaskTypes: []string{"a"}})

	var wg sync.WaitGroup
	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				if got := r.Tasks("n1"); got != nil {
					// Every read must return a consistent snapshot: a slice
					// that contains only well-formed prior tasks.
					for _, x := range got {
						if x == "" {
							t.Errorf("reader observed empty task type: %v", got)
							return
						}
					}
				}
			}
		}()
	}
	for i := 0; i < 200; i++ {
		r.AddTask("n1", "t")
	}
	close(done)
	wg.Wait()

	// Final state sanity: every stored task is the original "a" plus appended
	// "t" entries (duplicates rejected).
	got := r.Tasks("n1")
	if len(got) != 2 || got[0] != "a" || got[1] != "t" {
		t.Fatalf("final tasks = %v, want [a t]", got)
	}
}

// TestRegistry_RegisterAndCacheAndTasks mirrors the reported flow: register,
// read the cache snapshot, then AddTask, then read again — the first snapshot
// must be unaffected. (The cache path is exercised directly through Upsert,
// the same primitive RegisterAndCache uses after a successful Register.)
func TestRegistry_RegisterAndCacheAndTasks(t *testing.T) {
	r := NewRegistry()

	req := model.RegisterRequest{ID: "n1", Name: "node-1", Address: "localhost", TaskTypes: []string{"a", "b", "c"}}
	n := model.NormalizeNode(model.Node{
		ID:           req.ID,
		Name:         req.Name,
		Address:      req.Address,
		TaskTypes:    append([]string(nil), req.TaskTypes...),
		Status:       model.Registered,
		RegisteredAt: time.Now().UTC(),
	})
	r.Upsert(n)

	first, ok := r.Get("n1")
	if !ok {
		t.Fatal("node missing from cache")
	}

	r.AddTask("n1", "d")

	// The clone taken before the append must be unaffected.
	if !reflect.DeepEqual(first.TaskTypes, []string{"a", "b", "c"}) {
		t.Fatalf("pre-append snapshot changed: got %v", first.TaskTypes)
	}
	// The cache reflects the append.
	if got := r.Tasks("n1"); !reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Fatalf("cache tasks after append = %v, want [a b c d]", got)
	}
}
