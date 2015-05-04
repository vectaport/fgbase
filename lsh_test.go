package flowgraph

import (
	"math"
	"testing"
	"time"
)

func tbiLsh(x, y Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x, &y}, nil, nil)
	node.RunFunc = tbiLshRun
	return node
}

func tbiLshRun (node *Node) {
	x := node.Dsts[0]
	y := node.Dsts[1]

	x.Aux = uint(0)
	y.Aux = uint(0)
	var i uint = 0
	for {
		if (i>10) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			x.Aux = x.Aux.(uint) + 1
			y.Aux = y.Aux.(uint) + 1
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
			x.Aux = x.Aux.(float32) + 1
			y.Aux = y.Aux.(float32) + 1
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

	x.Aux = uint8(0)
	y.Aux = uint64(0)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = uint8(0)
	y.Aux = uint16(0)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}


	x.Aux = "Can you left shift a string by an int?"
	y.Aux = uint8(77)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = [4]complex128 {0+0i,0+0i,0+0i,0+0i}
	y.Aux = uint8(77)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}
	

	// read all the acks to clean up
	for  {
		node.RecvOne()
	}
	

}

func tboLsh(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node

}

func TestLsh(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,3)

	n[0] = tbiLsh(e[0], e[1])
	n[1] = FuncLsh(e[0], e[1], e[2])
	n[2] = tboLsh(e[2])

	RunAll(n, time.Second)

}

