package flowgraph

import (
	"math/rand"
	"testing"
	"time"
)

func tbi(x Edge) Node {


	node:=MakeNode("tbi", nil, []*Edge{&x}, nil, 
		func(n *Node) {
			x.Val = x.Aux
			x.Aux = (x.Aux.(int) + 1)
			time.Sleep(time.Duration(rand.Intn(10000))*time.Microsecond)
		})

	return node
	
}

func tbo(a Edge) Node {
	
	node:=MakeNode("tbo", []*Edge{&a}, nil, nil, 
		func (n *Node) {
			time.Sleep(time.Duration(rand.Intn(10000))*time.Microsecond)
		})
	return node

}

func TestArbit(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,4)

	e[0].Aux = 0
	e[1].Aux = 1000

	n[0] = tbi(e[0])
	n[1] = tbi(e[1])
	n[2] = FuncArbit(e[0], e[1], e[2])
	n[3] = tbo(e[2])

	RunAll(n, time.Second)

}

