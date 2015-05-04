package flowgraph

import (
	"math"
	"testing"
	"time"
)

func tbiSub(x, y Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x, &y}, nil, nil)
	node.RunFunc = tbiSubRun
	return node
}

func tbiSubRun(node *Node) {
	x := node.Dsts[0]
	y := node.Dsts[1]

	x.Aux = 0
	y.Aux = 0
	var i int = 0
	for {
		if (i>10) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			x.Aux = x.Aux.(int) + 2
			y.Aux = y.Aux.(int) + 1
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = float32(0)
	y.Aux = float32(0)
	i = 0
	for {
		if (i>9) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			x.Aux = x.Aux.(float32) - 1.
			y.Aux = y.Aux.(float32) + 1.
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}
	
	x.Aux = uint64(math.MaxUint64)
	y.Aux = -1
	i = 0
	for {
		if (i > 0) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = int8(-1)
	y.Aux = uint64(math.MaxUint64)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = int8(-1)
	y.Aux = uint32(math.MaxUint32)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	for  {
		node.RecvOne()
	}

}

func tboSub(a Edge) Node {
	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestSub(t *testing.T) {

	TraceLevel = V

	e := MakeEdges(3)

	var n [3]Node
	n[0] = tbiSub(e[0], e[1])
	n[1] = FuncSub(e[0], e[1], e[2])
	n[2] = tboSub(e[2])

	RunAll(n[:], time.Second)

}

