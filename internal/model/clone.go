package model

// Clone returns a deep copy of the node so callers cannot mutate cached
// state by editing shared backing arrays. Labels is a map (always distinct
// on copy), but TaskTypes is a slice: assigning the header directly shares
// the underlying array, so a later append from any holder can write through
// into every other copy. Copy the element storage explicitly.
func (n Node) Clone() Node {
	x := n
	if n.Labels != nil {
		x.Labels = map[string]string{}
		for k, v := range n.Labels {
			x.Labels[k] = v
		}
	}
	if n.TaskTypes != nil {
		x.TaskTypes = make([]string, len(n.TaskTypes))
		copy(x.TaskTypes, n.TaskTypes)
	}
	return x
}

func (h Heartbeat) Clone() Heartbeat {
	x := h
	if h.Extra != nil {
		x.Extra = map[string]any{}
		for k, v := range h.Extra {
			x.Extra[k] = v
		}
	}
	return x
}
