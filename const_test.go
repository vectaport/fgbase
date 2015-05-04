package flowgraph

import (
	"testing"
	"time"
)

func tbiConst(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil,
		func (n *Node) {
			x.Val = x.Aux
			x.Aux = x.Aux.(int) + 1
		})

	x.Aux = 0
	return node

}

func tboConst(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node

}

func TestConst(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,4)

	n[0] = tbiConst(e[0])
	n[1] = FuncConst(e[1], 1000)
	n[2] = FuncAdd(e[0], e[1], e[2])
	n[3] = tboConst(e[2])

	RunAll(n, time.Second)

}

