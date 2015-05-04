package flowgraph

import (
	"testing"
	"time"
)

func tbiSteervWork(n *Node) {
	x := n.Dsts[0]
	x.Val = x.Aux
	if (x.Aux.(int)<=1) {
		x.Aux = (x.Aux.(int) + 1)%2
	} else {
		x.Aux = x.Aux.(int) + 1
	}
}

func tbiSteerv(x Edge) Node {
	node:=MakeNode("tbi", nil, []*Edge{&x}, nil, tbiSteervWork)
	return node
}

func tboSteerv(a Edge) Node {
	node:=MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestSteerv(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(4,5)

	// initialize different state in the two source testbenches (tbiSteerv)
	e[0].Aux = 0
	e[1].Aux = 1000

	n[0] = tbiSteerv(e[0])
	n[1] = tbiSteerv(e[1])
	n[2] = FuncSteerv(e[0], e[1], e[2], e[3])
	n[3] = tboSteerv(e[2])
	n[4] = tboSteerv(e[3])

	RunAll(n, time.Second)

}

