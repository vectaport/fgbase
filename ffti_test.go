package flowgraph

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

const infitesimal=1.e-15

func tbiFFTIWork(n *Node) {
	x := n.Dsts[0]
	const sz = 128
	var vec = make([]complex128, sz, sz)
	rand.Seed(0x1515)
	
	delta := 3*2*math.Pi/float64(sz)
	domain := float64(0)

	for i := range vec {
		vec[i] = complex(math.Sin(domain), 0.0)
		domain += delta
	}
	x.Val = vec
}

func tbiFFTI(x Edge) Node {
	node:=MakeNode("tbi", nil, []*Edge{&x}, nil, tbiFFTIWork)
	return node
}

func tboFFTIWork(n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	av := a.Val.([]complex128)
	bv := b.Val.([]complex128)
	if (len(av)==len(bv)) {
		for i := range av {
			if (real(av[i])-real(bv[i])) < -infitesimal || (real(av[i])-real(bv[i]))>infitesimal || 
				(imag(av[i])-imag(bv[i])) < -infitesimal || (imag(av[i])-imag(bv[i]))>infitesimal {
				n.Tracef("!SAME:  for %d delta is %v\n", i, av[i]-bv[i])
				n.Tracef("!SAME:  a = %v,  b = %v\n", av[i], bv[i])
				return
			}
		}
		n.Tracef("SAME all differences smaller than %v\n", infitesimal)
		return
	} 
	n.Tracef("!SAME:  different sizes\n")
}

func tboFFTI(a, b Edge) Node {
	node:=MakeNode("tbo", []*Edge{&a, &b}, nil, nil, tboFFTIWork)
	return node
}

func TestFFTI(t *testing.T) {

	TraceLevel = V
	
	e,n := MakeGraph(9,7)

	e[7].Const(false)
	e[8].Const(true)

	n[0] = tbiFFTI(e[0])

	n[1] = FuncFork(e[0], e[1], e[2])

	n[2] = FuncFFT(e[1], e[7], e[3])
	n[3] = FuncPass(e[2], e[4])

	n[4] = FuncFFT(e[3], e[8], e[5])
	n[5] = FuncPass(e[4], e[6])

	n[6] = tboFFTI(e[5], e[6])

	RunAll(n, time.Second)

}

