package model

import "testing"

func TestBug005NodeCloneOwnsTaskTypeStorage(t *testing.T) {
	original := Node{TaskTypes: []string{"orders", "billing"}}
	clone := original.Clone()
	clone.TaskTypes[0] = "mutated"
	if original.TaskTypes[0] != "orders" { t.Fatalf("clone mutation contaminated original node") }
}
