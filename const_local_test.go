package flowgraph

import (
	"testing"
	"time"
)

func tbiConstLocal(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil, 
		func (n *Node) {
			x.Val = x.Aux
			x.Aux = x.Aux.(int) + 1
		})
			
	x.Aux = 0
	return node

}

func tboConstLocal(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
	
}

func TestConstLocal(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,3)

	e[1].Const(1000)

	n[0] = tbiConstLocal(e[0])
	n[1] = FuncAdd(e[0], e[1], e[2])
	n[2] = tboConstLocal(e[2])

	RunAll(n, time.Second)

}

