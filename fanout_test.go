package flowgraph

import (
	"testing"
	"time"
)

func tbiFanout(x Edge) Node {

	x.Aux = 0
	node := MakeNode("tbi", nil, []*Edge{&x}, nil, 
		func (n *Node) { 
			x.Val = x.Aux
			x.Aux = x.Aux.(int) + 1
		})
	return node
}

func tboFanout(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestFanout(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(2,4)

	n[0] = tbiFanout(e[0])
	n[1] = FuncPass(e[0], e[1])
	n[2] = tboFanout(e[1])
	n[3] = tboFanout(e[1])

	RunAll(n, time.Second)
}

