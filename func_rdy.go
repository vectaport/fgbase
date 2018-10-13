package fgbase

import ()

func RdyFire(n *Node) error {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	b.Flow = true
	x.DstPut(a.SrcGet())
	return nil
}

// FuncRdy waits for two values before passing on the first (b; x = a).
func FuncRdy(a, b, x Edge) Node {

	node := MakeNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil, RdyFire)
	return node
}
