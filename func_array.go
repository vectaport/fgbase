package fgbase

import ()

// ArrayFire is fire func for FuncArray
func ArrayFire(n *Node) error {
	x := n.Dsts[0]
	arr := n.Aux.([]interface{})
	if len(arr) > 0 {
		x.DstPut(arr[0])
	} else if len(arr) == 0 {
		x.DstPut(EOF)
		return EOF // ??? causes write on channel to never finish ???
	}
	n.Aux = arr[1:]
	return nil
}

// FuncArray streams the contents of an array then stops
func FuncArray(x Edge, arr []interface{}) Node {

	node := MakeNode("array", nil, []*Edge{&x}, nil, ArrayFire)
	node.Aux = arr
	return node
}
