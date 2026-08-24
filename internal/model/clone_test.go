package model

import (
	"reflect"
	"testing"
)

// TestClone_TaskTypesDeepCopy locks down the original bug: Clone shared the
// TaskTypes backing array, so an append into spare capacity on any clone wrote
// through into the source and every sibling clone. The clone must own its
// storage.
func TestClone_TaskTypesDeepCopy(t *testing.T) {
	// cap > len so append reuses the backing array instead of allocating.
	src := Node{ID: "n1", TaskTypes: []string{"a", "b", "c"}}

	a := src.Clone()
	b := src.Clone()

	// Append into spare capacity on both clones.
	a.TaskTypes = append(a.TaskTypes, "x")
	b.TaskTypes = append(b.TaskTypes, "y")

	if got := src.TaskTypes; !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("source mutated via shared backing array: got %v", got)
	}
	if got := a.TaskTypes; !reflect.DeepEqual(got, []string{"a", "b", "c", "x"}) {
		t.Fatalf("clone a polluted: got %v", got)
	}
	if got := b.TaskTypes; !reflect.DeepEqual(got, []string{"a", "b", "c", "y"}) {
		t.Fatalf("clone b polluted by sibling append: got %v", got)
	}
}

// TestClone_NilAndEmptyTaskTypes ensures the copy path handles both the nil
// (unset) and empty cases without aliasing.
func TestClone_NilAndEmptyTaskTypes(t *testing.T) {
	if c := (Node{ID: "n"}).Clone(); c.TaskTypes != nil {
		t.Fatalf("expected nil TaskTypes for nil source, got %v", c.TaskTypes)
	}
	src := Node{ID: "n", TaskTypes: []string{}}
	c := src.Clone()
	if c.TaskTypes == nil || len(c.TaskTypes) != 0 {
		t.Fatalf("expected empty non-nil TaskTypes, got %v", c.TaskTypes)
	}
	// Mutating the empty clone must not panic or alias.
	c.TaskTypes = append(c.TaskTypes, "z")
	if len(src.TaskTypes) != 0 {
		t.Fatalf("empty source mutated via shared array: got %v", src.TaskTypes)
	}
}

// TestClone_LabelsDeepCopy confirms the labels map is still isolated (it was
// already correct; this guards against regressions when touching Clone).
func TestClone_LabelsDeepCopy(t *testing.T) {
	src := Node{ID: "n", Labels: map[string]string{"env": "prod"}}
	c := src.Clone()
	c.Labels["env"] = "dev"
	if src.Labels["env"] != "prod" {
		t.Fatalf("source labels mutated via shared map: got %v", src.Labels["env"])
	}
}
