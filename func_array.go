package fgbase

import ()

// ArrayFire is fire func for FuncArray
func ArrayFire(n *Node) error {
	x := n.Dsts[0]
	arr := n.Aux.([]any)
	if len(arr) > 0 {
		x.DstPut(arr[0])
	} else if len(arr) == 0 {
		x.DstPut(EOS)
		return EOS
	}
	n.Aux = arr[1:]
	return nil
}

// FuncArray streams the contents of an array then stops
func FuncArray(x Edge, arr []any) Node {

	node := MakeNode("array", nil, []*Edge{&x}, nil, ArrayFire)
	node.Aux = arr
	return node
}
