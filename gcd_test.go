package flowgraph

import (
	"math/rand"
	"testing"
	"time"
)

func tbmGCD(x Edge) Node {

	node := MakeNode("tbm", nil, []*Edge{&x}, nil,
		func(n *Node) { n.Dsts[0].Val = rand.Intn(15)+1 })
	return node
}

func tbnGCD(x Edge) Node {

	node := MakeNode("tbm", nil, []*Edge{&x}, nil,
		func(n *Node) { n.Dsts[0].Val = rand.Intn(15)+1 })
	return node
}

func tboGCD(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestGCD(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(11, 10)

	e[7].Val = 0

	n[0] = tbmGCD(e[0])
	n[1] = tbnGCD(e[1])

	n[2] = FuncRdy(e[0], e[7], e[2])
	n[3] = FuncRdy(e[1], e[7], e[3])

	n[4] = FuncEither(e[2], e[10], e[4])
	n[5] = FuncEither(e[3], e[8], e[5])

	n[6] = FuncMod(e[4], e[5], e[6])

	n[7] = FuncSteerc(e[6], e[7], e[8])
	n[8] = FuncSteerv(e[6], e[5], e[9], e[10])

	n[9] = tboGCD(e[9])

	RunAll(n, time.Second)

}
