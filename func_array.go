package fgbase

import (
	"io"
)

type arrayStruct struct {
	arr []interface{}
	cur int
}

func arrayFire(n *Node) error {
	x := n.Dsts[0]
	as := n.Aux.(arrayStruct)
	if as.cur < len(as.arr) {
		x.DstPut(as.arr[as.cur])
	} else if as.cur == len(as.arr) {
		x.DstPut(io.EOF)
	}
	n.Aux = arrayStruct{as.arr, as.cur + 1}
	return nil
}

// FuncArray streams the contents of an array then stops
func FuncArray(x Edge, arr []interface{}) Node {

	node := MakeNode("array", nil, []*Edge{&x}, nil, arrayFire)
	node.Aux = arrayStruct{arr, 0}
	return node
}
