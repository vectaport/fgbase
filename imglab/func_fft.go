package imglab

import (
	"github.com/ledyba/go-fft/fft"
	"github.com/vectaport/flowgraph"
)


func fftFire (n *flowgraph.Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	x.DstPut(a.SrcGet())
	data,ok := x.Val.([]complex128)
	if !ok {
		n.LogError("type is not []complex128\n")
		x.Val = nil
	}
	bv := b.SrcGet()
	if bv.(bool) {
		fft.InvFft(data)
	} else {
		fft.Fft(data)
	}
}

// FuncFFT does an FFT on a slice of complex128 (x=fft(data: a, inverse: b)).
func FuncFFT(a, b, x flowgraph.Edge) flowgraph.Node {

	node := flowgraph.MakeNode("fft", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x}, nil, fftFire)
	return node

}
