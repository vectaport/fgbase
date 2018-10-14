package fgbase

import ()

func SteercRdy(n *Node) bool {
	a := n.Srcs[0]
	if a.SrcRdy(n) {
		if ZeroTest(a.Val) {
			return n.Dsts[0].DstRdy(n)
		}
		i := min(Int(a.Val), len(n.Dsts)-1)
		return n.Dsts[i].DstRdy(n)
	}
	return false
}

func SteercFire(n *Node) error {
	a := n.Srcs[0]
	av := a.SrcGet()
	if ZeroTest(av) {
		n.Dsts[0].DstPut(av)
	} else {
		i := min(Int(av), len(n.Dsts)-1)
		n.Dsts[i].DstPut(av)
	}
	return nil
}

// FuncSteerc steers a condition one of two ways (if a==0 { x = a } else { y = a }).
func FuncSteerc(a, x, y Edge) Node {

	node := MakeNode("steerc", []*Edge{&a}, []*Edge{&x, &y}, SteercRdy, SteercFire)
	return node

}
