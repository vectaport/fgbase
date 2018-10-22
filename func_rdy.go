package fgbase

import ()

// RdyFire is fire func for FuncRdy
func RdyFire(n *Node) error {
	/*
		a := n.Srcs[0]
		b := n.Srcs[1]
		x := n.Dsts[0]
		b.Flow = true
		x.DstPut(a.SrcGet())
		return nil
	*/
	n.Srcs[n.SrcCnt()-1].Flow = true
	for i := 0; i < len(n.Srcs)-1; i++ {
		n.Dsts[i].DstPut(n.Srcs[i].SrcGet())
	}
	return nil
}

// FuncRdy waits for two values before passing on the first (b; x = a).
func FuncRdy(a, b, x Edge) Node {

	node := MakeNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil, RdyFire)
	return node
}
