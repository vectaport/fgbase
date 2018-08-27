package flowgraph

import ()

func incomingFire(n *Node) {
	x := n.Dsts[0]
	r := n.Aux.(Getter)
	v, _ := r.Get()
	x.DstPut(v)
}

// FuncIncoming imports one input value using a Getter and feeds it to the flowgraph
func FuncIncoming(x Edge, receiver Getter) Node {

	node := MakeNode("incoming", nil, []*Edge{&x}, nil, incomingFire)
	node.Aux = receiver
	return node
}
