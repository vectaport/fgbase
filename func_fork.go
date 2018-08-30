package fgbase

import ()

func forkFire(n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	y := n.Dsts[1]
	x.Val = a.Val
	if IsSlice(a.Val) {
		y.DstPut(CopySlice(a.SrcGet()))
	} else {
		y.DstPut(a.SrcGet())
	}
}

// FuncFork sends a value two ways (x = a; y = a).
// If the value is a slice it is duplicated onto the second output.
func FuncFork(a, x, y Edge) Node {

	node := MakeNode("fork", []*Edge{&a}, []*Edge{&x, &y}, nil, forkFire)
	return node

}
