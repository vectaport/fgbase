package flowgraph

import (
	"fmt"
	"github.com/ledyba/go-fft/fft"
)


func fftFire (n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	x.Val = a.Val
	data,ok := x.Val.([]complex128)
	if !ok {
		x.Val = fmt.Errorf("wrong type\n")
	} else {
		fft.Fft(data)
	}
}

// FuncFft does an fft on a slice of complex128
func FuncFft(a, x Edge) {

	node := MakeNode("fft", []*Edge{&a}, []*Edge{&x}, nil, fftFire)
	node.Run()

}
