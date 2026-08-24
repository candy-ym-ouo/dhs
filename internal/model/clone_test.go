package model

import "testing"

// Clone must deep-copy TaskTypes so the returned node shares no backing array
// with the source. Mutating the clone's TaskTypes (by index or via append that
// stays within capacity) must not touch the original.
func TestNodeCloneTaskTypesIsIndependent(t *testing.T) {
	// Craft a slice with spare capacity so an in-cap write is detectable.
	tt := make([]string, 2, 4)
	tt[0], tt[1] = "alpha", "beta"
	src := Node{ID: "n1", Labels: map[string]string{"k": "v"}, TaskTypes: tt}

	clone := src.Clone()

	// In-cap index write on the clone must not affect src.
	clone.TaskTypes[0] = "CHANGED"
	if src.TaskTypes[0] == "CHANGED" {
		t.Fatalf("clone index write mutated original: src=%v", src.TaskTypes)
	}

	// In-cap append on the clone must not affect src. cap 4 / len 2 means the
	// shallow-Clone bug would have let append write into the spare slots the
	// original still owns — we detect it via src's own in-cap append landing
	// on unmutated values.
	clone2 := src.Clone()
	clone2.TaskTypes = append(clone2.TaskTypes, "gamma")
	if len(clone2.TaskTypes) != 3 || clone2.TaskTypes[2] != "gamma" {
		t.Fatalf("clone did not get its own append: %v", clone2.TaskTypes)
	}
	if len(src.TaskTypes) != 2 || src.TaskTypes[0] != "alpha" || src.TaskTypes[1] != "beta" {
		t.Fatalf("clone in-cap append leaked into original: src=%v", src.TaskTypes)
	}
	// If the shallow-Clone bug were present, the original's own in-cap append
	// would now reveal "gamma" that the clone wrote into the shared buffer.
	src.TaskTypes = append(src.TaskTypes, "delta")
	if src.TaskTypes[2] == "gamma" {
		t.Fatalf("clone and original share backing array (shallow Clone): src=%v", src.TaskTypes)
	}
}

// Clone must also copy Labels independently (regression guard for the sibling
// field fixed alongside TaskTypes).
func TestNodeCloneLabelsIsIndependent(t *testing.T) {
	src := Node{ID: "n1", Labels: map[string]string{"k": "v"}}
	clone := src.Clone()
	clone.Labels["k"] = "CHANGED"
	if src.Labels["k"] == "CHANGED" {
		t.Fatalf("clone label write mutated original")
	}
}

func TestNodeCloneNilTaskTypesStaysNil(t *testing.T) {
	src := Node{ID: "n1"}
	c := src.Clone()
	if c.TaskTypes != nil {
		t.Fatalf("nil TaskTypes should stay nil after clone, got %v", c.TaskTypes)
	}
}
