package imglab

import (
	"testing"
	"time"

	"github.com/vectaport/fgbase"
)

func tbiFFTFire(n *fgbase.Node) {
	x := n.Dsts[0]
	x.Val = make([]complex128, 32, 32)
}

func tbiFFT(x fgbase.Edge) fgbase.Node {
	node:=fgbase.MakeNode("tbi", nil, []*fgbase.Edge{&x}, nil, tbiFFTFire)
	return node
}

func tboFFT(a fgbase.Edge) fgbase.Node {
	node:=fgbase.MakeNode("tbo", []*fgbase.Edge{&a}, nil, nil, nil)
	return node
}

func TestFFT(t *testing.T) {

	fgbase.TraceLevel = fgbase.V

	e,n := fgbase.MakeGraph(3,3)

	e[1].Const(false)

	n[0] = tbiFFT(e[0])
	n[1] = FuncFFT(e[0], e[1], e[2])
	n[2] = tboFFT(e[2])

	fgbase.RunAll(n, time.Second)

}

