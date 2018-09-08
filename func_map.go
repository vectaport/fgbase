package fgbase

import ()

func mapFire(n *Node) error {
	a := n.Srcs[0]
	x := n.Dsts
	i := n.Aux.(int)
	x[i].DstPut(n.NodeWrap(a.SrcGet(), x[i].Ack))
	return nil
}

// FuncMap maps a value to one of n reducers.
func FuncMap(a, x []Edge, mapper func(n *Node, datum interface{}) int) *Pool {

	var mapRdy = func(n *Node) bool {
		a := n.Srcs[0]
		x := n.Dsts
		if a.SrcRdy(n) {
			i := mapper(n, a.Val)
			n.Aux = i
			if i < 0 {
				return false
			}
			return x[i].DstRdy(n)
		}
		return false
	}

	// Make a pool of map nodes that share input and output channels
	recurse := false
	spread := true
	return MakePool(len(a), "map", a, x, mapRdy, mapFire, recurse, spread)

}
