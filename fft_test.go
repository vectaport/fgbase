package flowgraph

import (
	"testing"
	"time"
)

func tbiFFTWork(n *Node) {
	x := n.Dsts[0]
	x.Val = make([]complex128, 32, 32)
}

func tbiFFT(x Edge) Node {
	node:=MakeNode("tbi", nil, []*Edge{&x}, nil, tbiFFTWork)
	return node
}

func tboFFT(a Edge) Node {
	node:=MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestFFT(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,3)

	e[1].Const(false)

	n[0] = tbiFFT(e[0])
	n[1] = FuncFFT(e[0], e[1], e[2])
	n[2] = tboFFT(e[2])

	RunAll(n, time.Second)

}

