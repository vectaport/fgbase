package flowgraph

import (
	"io"
)

func incomingFire(n *Node) {
	x := n.Dsts[0]
	r := n.Aux.(Getter)
	v, err := r.Get(n.Owner)
	if err != nil {
		if err != io.EOF {
			n.LogError(err.Error())
		}
		return
	}
	x.DstPut(v)
}

// FuncIncoming imports one input value using a Getter and feeds it to the flowgraph
func FuncIncoming(x Edge, receiver Getter) Node {

	node := MakeNode("incoming", nil, []*Edge{&x}, nil, incomingFire)
	node.Aux = receiver
	return node
}
