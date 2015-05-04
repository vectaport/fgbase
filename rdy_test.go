package flowgraph

import (
	"testing"
	"time"
)

func tbiRdy(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil, 
		func (n *Node) { 
			x.Val = x.Aux
			x.Aux = x.Aux.(int) + 1
		})
	return node
}

func tboRdy(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestRdy(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,4)
 
	e[0].Aux = 0
	e[1].Aux = 1000

	n[0] = tbiRdy(e[0])
	n[1] = tbiRdy(e[1])
	n[2] = FuncRdy(e[0], e[1], e[2])
	n[3] = tboRdy(e[2])

	RunAll(n, time.Second)

}

