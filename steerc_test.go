package flowgraph

import (
	"testing"
	"time"
)

func tbiSteercFire(n *Node) {
	x := n.Dsts[0]
	x.Val = x.Aux
	x.Aux = (x.Aux.(int) + 1)%2
}

func tbiSteerc(x Edge) Node {

	node:=MakeNode("tbi", nil, []*Edge{&x}, nil, tbiSteercFire)
	x.Aux = 0
	return node
	
}

func tboSteerc(a Edge) Node {
	node:=MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestSteerc(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,4)

	n[0] = tbiSteerc(e[0])
	n[1] = FuncSteerc(e[0], e[1], e[2])
	n[2] = tboSteerc(e[1])
	n[3] = tboSteerc(e[2])

	RunAll(n, time.Second)

}

