package flowgraph

import ()

func outgoingFire(n *Node) {
	a := n.Srcs[0]
	d := n.Aux.(Deliverer)
	_ = d.Deliver(a.SrcGet())
}

// FuncOutgoing accepts one output value from the flowgraph and exports it using a Deliverer
func FuncOutgoing(a Edge, deliverer Deliverer) Node {

	node := MakeNode("outgoing", []*Edge{&a}, nil, nil, outgoingFire)
	node.Aux = deliverer
	return node
}
