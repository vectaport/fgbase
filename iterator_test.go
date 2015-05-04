package flowgraph

import (
	"math/rand"
	"testing"
	"time"
)

func tbiIterator(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil,
		func(n *Node) { n.Dsts[0].Val = rand.Intn(7)+1 })
	return node
}

func TestIterator(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(7,5)

	e[3].Const(1)
	e[5].Val = 0

	n[0] = tbiIterator(e[0])
	n[1] = FuncRdy(e[0], e[5], e[1])
	n[2] = FuncEither(e[1], e[6], e[2])
	n[3] = FuncSub(e[2], e[3], e[4])
	n[4] = FuncSteerc(e[4], e[5], e[6])

	RunAll(n, time.Second)

}
