package fgbase

import ()

// FuncPass passes a value on (x = a).
func FuncPass(a, x Edge) Node {

	node := MakeNode("pass", []*Edge{&a}, []*Edge{&x}, nil, func(n *Node) error { n.Dsts[0].DstPut(n.Srcs[0].SrcGet());return nil })
	return node

}
