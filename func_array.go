package fgbase

import ()

type arrayStruct struct {
	arr []interface{}
	cur int
}

func arrayFire(n *Node) {
	x := n.Dsts[0]
	as := n.Aux.(arrayStruct)
	if as.cur < len(as.arr) {
		x.DstPut(as.arr[as.cur])
		n.Aux = arrayStruct{as.arr, as.cur + 1}
	}
}

// FuncArray streams the contents of an array then stops
func FuncArray(x Edge, arr []interface{}) Node {

	node := MakeNode("array", nil, []*Edge{&x}, nil, arrayFire)
	node.Aux = arrayStruct{arr, 0}
	return node
}
