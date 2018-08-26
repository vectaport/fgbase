package flowgraph

import (
)

func incomingFire(n *Node) {
	x := n.Dsts[0]
	r := n.Aux.(Receiver)
	v,_ := r.Receive()
	x.DstPut(v)
}

// FuncIncoming imports one input value using a Receiver and feeds it to the flowgraph
func FuncIncoming(x Edge, receiver Receiver) Node {

	node:=MakeNode("incoming", nil, []*Edge{&x}, nil, incomingFire)
	node.Aux = receiver
	return node
}
