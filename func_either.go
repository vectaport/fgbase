package fgbase

import ()

func eitherFire(n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	if a.SrcRdy(n) {
		x.DstPut(a.SrcGet())
	} else {
		x.DstPut(b.SrcGet())
	}
}

func eitherRdy(n *Node) bool {
	return (n.Srcs[0].SrcRdy(n) || n.Srcs[1].SrcRdy(n)) && n.Dsts[0].DstRdy(n)
}

// FuncEither passes on one of two values (if =a { x = a } else { x = b }).
func FuncEither(a, b, x Edge) Node {

	node := MakeNode("either", []*Edge{&a, &b}, []*Edge{&x}, eitherRdy, eitherFire)
	return node

}
