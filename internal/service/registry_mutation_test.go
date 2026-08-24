package service

import (
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"dhs/internal/model"
)

func nodeFixture(id string, tasks ...string) model.Node {
	n := model.Node{
		ID:        id,
		Name:      id,
		Status:    model.Online,
		TaskTypes: append([]string(nil), tasks...),
	}
	return model.NormalizeNode(n)
}

// Callers modifying a returned value (Get/Snapshot/List) must not mutate the
// cache. This is the core cross-layer pollution guarantee.
func TestRegistryReturnedNodesAreIndependent(t *testing.T) {
	r := NewRegistry()
	// Spare capacity so an in-cap write is the exact failure mode a shallow
	// Clone would have aliased.
	r.nodes["n1"] = nodeFixture("n1", "alpha", "beta") // ignore lock for setup

	got, ok := r.Get("n1")
	if !ok {
		t.Fatal("missing node")
	}
	// In-cap index write on the returned copy must not reach the cache.
	got.TaskTypes[0] = "CHANGED"
	again, _ := r.Get("n1")
	if again.TaskTypes[0] == "CHANGED" {
		t.Fatalf("Get return-value mutation leaked into cache: %v", again.TaskTypes)
	}

	// Snapshot independence.
	snap := r.Snapshot()
	snap["n1"].TaskTypes[1] = "LEAKED"
	after, _ := r.Get("n1")
	if after.TaskTypes[1] == "LEAKED" {
		t.Fatalf("Snapshot mutation leaked into cache: %v", after.TaskTypes)
	}
}

// AddTask must not overwrite data a caller is still holding via a snapshot —
// "追加任务不能覆盖已有快照". It also must keep TaskTypes sorted and behave
// idempotently (set semantics, matching HasTask).
func TestRegistryAddTaskDoesNotOverwriteSnapshotAndStaysSorted(t *testing.T) {
	r := NewRegistry()
	r.nodes["n1"] = nodeFixture("n1", "compute", "storage")

	snap := r.Snapshot()
	snapTasks := append([]string(nil), snap["n1"].TaskTypes...)

	if !r.AddTask("n1", "network") {
		t.Fatal("AddTask returned false for existing node")
	}
	// The previously published snapshot must be untouched.
	if !reflect.DeepEqual(snap["n1"].TaskTypes, snapTasks) {
		t.Fatalf("AddTask overwrote existing snapshot: got %v want %v", snap["n1"].TaskTypes, snapTasks)
	}
	// Cache reflects the addition, sorted.
	got := r.Tasks("n1")
	want := []string{"compute", "network", "storage"}
	if !sort.StringsAreSorted(got) || !reflect.DeepEqual(got, want) {
		t.Fatalf("AddTask did not keep sorted set: got %v want %v", got, want)
	}

	// Idempotent: adding an existing task must not duplicate.
	if !r.AddTask("n1", "network") {
		t.Fatal("AddTask of existing task should be idempotent (return true)")
	}
	got = r.Tasks("n1")
	if len(got) != 3 {
		t.Fatalf("duplicate task accumulated after idempotent AddTask: %v", got)
	}
}

// AddTask on a missing node must report failure and not panic.
func TestRegistryAddTaskMissingNode(t *testing.T) {
	r := NewRegistry()
	if r.AddTask("nope", "x") {
		t.Fatal("AddTask on missing node should return false")
	}
}

// Tasks must return an independent copy: mutating it must not affect the cache
// or subsequent reads.
func TestRegistryTasksReturnsCopy(t *testing.T) {
	r := NewRegistry()
	r.nodes["n1"] = nodeFixture("n1", "alpha")

	a := r.Tasks("n1")
	a[0] = "CHANGED"

	b := r.Tasks("n1")
	if b[0] == "CHANGED" {
		t.Fatalf("Tasks return-value mutation leaked into cache: %v", b)
	}
}

// Concurrency: AddTask (write) under heavy parallel readers/writers must not
// race or corrupt. Preserves the existing locking semantics.
func TestRegistryAddTaskConcurrent(t *testing.T) {
	r := NewRegistry()
	r.nodes["n1"] = nodeFixture("n1")

	var wg sync.WaitGroup
	tasks := []string{"a", "b", "c", "d", "e", "a", "b", "c"}
	for _, tk := range tasks {
		wg.Add(1)
		go func(tk string) {
			defer wg.Done()
			_ = r.AddTask("n1", tk)
		}(tk)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = r.Get("n1")
			_ = r.Snapshot()
			_ = r.Tasks("n1")
		}()
	}
	wg.Wait()

	got := r.Tasks("n1")
	if !sort.StringsAreSorted(got) {
		t.Fatalf("TaskTypes not sorted after concurrent AddTask: %v", got)
	}
	// Set invariant: no duplicates.
	seen := map[string]bool{}
	for _, v := range got {
		if seen[v] {
			t.Fatalf("duplicate task after concurrent AddTask: %v", got)
		}
		seen[v] = true
	}
}

// Replace and Stale/copy paths keep handing out clones (guard against regressions
// in the copy discipline after the Clone fix).
func TestRegistryReplaceAndListAreIndependent(t *testing.T) {
	r := NewRegistry()
	r.Replace([]model.Node{nodeFixture("n1", "alpha")})

	lst := r.List("", "")
	lst[0].TaskTypes[0] = "CHANGED"
	again := r.List("", "")
	if again[0].TaskTypes[0] == "CHANGED" {
		t.Fatalf("List return-value mutation leaked into cache: %v", again[0].TaskTypes)
	}

	now := time.Now()
	stale := r.Stale(now, 0)
	if len(stale) != 1 {
		t.Fatalf("Stale should return the node, got %d", len(stale))
	}
	stale[0].TaskTypes[0] = "CHANGED"
	if r.nodes["n1"].TaskTypes[0] == "CHANGED" {
		t.Fatalf("Stale return-value mutation leaked into cache")
	}
}
