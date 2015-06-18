package flowgraph

import (
	"testing"
	"time"
)

func tbiAddRingFire(n *Node) {
	x := n.Dsts[0]
	y := n.Dsts[1]
	x.Val = x.Aux
	y.Val = y.Aux
	x.Aux = x.Aux.(int) + 1
	y.Aux = y.Aux.(int) + 1
}

func tbiAddRing(a, x, y Edge) Node {
	node := MakeNode("tbi", []*Edge{&a}, []*Edge{&x, &y}, nil, tbiAddRingFire)
	x.Aux = 1
	y.Aux = 1
	return node
}

func tboAddRingFire(n *Node) {
//	x := n.Dsts[0]
//	x.Val = true
}

func tboAddRing(a, x Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, []*Edge{&x}, nil, tboAddRingFire)
	return node

}

func TestAddRing(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(4,3)

	e[3].Val = true // initialize data wavefront

	n[0] = tbiAddRing(e[3], e[0], e[1])
	n[1] = FuncAdd(e[0], e[1], e[2])
	n[2] = tboAddRing(e[2], e[3])

	RunAll(n, time.Second)

}

