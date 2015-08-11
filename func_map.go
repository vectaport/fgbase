package flowgraph

import (
)

func mapFire (n *Node) {
	a := n.Srcs[0]
	x := n.Dsts
	for j := range x {
		x[j].NoOut = true
	}
	i := a.Aux.(int)
	x[i].Val = n.NodeWrap(a.Val, x[i].Ack)
	x[i].NoOut = false

}

// FuncMap maps a value to one of n reducers.
func FuncMap(a, x []Edge, mapper func(n *Node, datum Datum) int) Pool {

	var  mapRdy = func (n *Node) bool {
		a := n.Srcs[0]
		x := n.Dsts
		if a.SrcRdy(n) {
			i := mapper(n, a.Val)
			a.Aux = i
			if i<0 {return false} 
			return x[i].DstRdy(n)
		}
		return false
	}
	
	// Make a pool of map nodes that share input and output channels
	recurse := false
	spread := true
	return MakePool(len(a), "map", a, x, mapRdy, mapFire, recurse, spread)

}
