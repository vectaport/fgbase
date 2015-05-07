package flowgraph

import (
	"github.com/ledyba/go-fft/fft"
)


func fftFire (n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	x.Val = a.Val
	data,ok := x.Val.([]complex128)
	if !ok {
		n.LogError("type is not []complex128\n")
		x.Val = nil
	}
	if b.Val.(bool) {
		fft.InvFft(data)
	} else {
		fft.Fft(data)
	}
}

// FuncFFT does an FFT on a slice of complex128 (x=fft(data: a, inverse: b)).
func FuncFFT(a, b, x Edge) Node {

	node := MakeNode("fft", []*Edge{&a, &b}, []*Edge{&x}, nil, fftFire)
	return node

}
