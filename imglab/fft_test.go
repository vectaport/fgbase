package imglab

import (
	"testing"
	"time"

	"github.com/vectaport/flowgraph"
)

func tbiFFTFire(n *flowgraph.Node) {
	x := n.Dsts[0]
	x.Val = make([]complex128, 32, 32)
}

func tbiFFT(x flowgraph.Edge) flowgraph.Node {
	node:=flowgraph.MakeNode("tbi", nil, []*flowgraph.Edge{&x}, nil, tbiFFTFire)
	return node
}

func tboFFT(a flowgraph.Edge) flowgraph.Node {
	node:=flowgraph.MakeNode("tbo", []*flowgraph.Edge{&a}, nil, nil, nil)
	return node
}

func TestFFT(t *testing.T) {

	flowgraph.TraceLevel = flowgraph.V

	e,n := flowgraph.MakeGraph(3,3)

	e[1].Const(false)

	n[0] = tbiFFT(e[0])
	n[1] = FuncFFT(e[0], e[1], e[2])
	n[2] = tboFFT(e[2])

	flowgraph.RunAll(n, time.Second)

}

