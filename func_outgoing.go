package flowgraph

import ()

func outgoingFire(n *Node) {
	a := n.Srcs[0]
	d := n.Aux.(Putter)
	_ = d.Put(a.SrcGet())
}

// FuncOutgoing accepts one output value from the flowgraph and exports it using a Putter
func FuncOutgoing(a Edge, putter Putter) Node {

	node := MakeNode("outgoing", []*Edge{&a}, nil, nil, outgoingFire)
	node.Aux = putter
	return node
}
