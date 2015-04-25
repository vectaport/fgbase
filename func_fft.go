package flowgraph

import (
	"fmt"

	"github.com/ledyba/go-fft/fft"
)


func fftFire (n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	x.Val = a.Val
	data,ok := x.Val.([]complex128)
	if !ok {
		x.Val = fmt.Errorf("type is not []complex128\n")
	} else {
		if b.Val.(bool) {
			fft.InvFft(data)
		} else {
			fft.Fft(data)
		}
	}
}

// FuncFft does an fft on a slice of complex128 (fft(data: a, inverse: b)).
func FuncFft(a, b, x Edge) Node {

	node := MakeNode("fft", []*Edge{&a, &b}, []*Edge{&x}, nil, fftFire)
	return node

}
