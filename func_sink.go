package flowgraph

import (
)

func sinkFire(n *Node) {
	a := n.Srcs[0]
	a.SrcGet()
}

// FuncSink sinks a single value.
func FuncSink(a Edge) Node {

	node:=MakeNode("sink", []*Edge{&a}, nil, nil, sinkFire)
	return node
}
