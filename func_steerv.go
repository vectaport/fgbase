package fgbase

import ()

func steervRdy(n *Node) bool {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]
	if a.SrcRdy(n) && b.SrcRdy(n) {
		if ZeroTest(a.Val) {
			return x.DstRdy(n)
		}
		return y.DstRdy(n)
	}
	return false
}

func steervFire(n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]
	if ZeroTest(a.SrcGet()) {
		x.DstPut(b.SrcGet())
	} else {
		y.DstPut(b.SrcGet())
	}
}

// FuncSteerv steers the second value by the first (if a==0 { x = b } else { y = b }).
func FuncSteerv(a, b, x, y Edge) Node {

	node := MakeNode("steerv", []*Edge{&a, &b}, []*Edge{&x, &y}, steervRdy, steervFire)
	return node

}
