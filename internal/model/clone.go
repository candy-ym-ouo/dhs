package model

func (n Node) Clone() Node {
	x := n
	if n.Labels != nil {
		x.Labels = map[string]string{}
		for k, v := range n.Labels {
			x.Labels[k] = v
		}
	}
	if n.TaskTypes != nil {
		x.TaskTypes = n.TaskTypes
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
