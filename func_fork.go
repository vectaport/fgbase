package flowgraph

import (
)

var PoolFork Pool

// ForkSel is the function signature used by FuncFork to select an output edge given
// current input and internal state of the Node.
type ForkSel func(*Node) int

func forkFire (n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]
	if (ZeroTest(a.Val)) {
		x.Val = b.Val
		y.NoOut = true
	} else {
		y.Val = b.Val
		x.NoOut = true
	}
}

// FuncFork steers (or maps) a value into one of n directions.
func FuncFork(a Edge, x []Edge, xSelect ForkSel) *Pool {

	i := 0
	var forkRdy = func (n *Node) bool {
		a := n.Srcs[0]
		x := n.Dsts
		if a.SrcRdy(n) {
			i = xSelect(n)
			if i<0 {return false} else {return x[i].DstRdy(n)}
		}
		return false
	}

	var forkFire = func (n *Node) {
		a := n.Srcs[0]
		x := n.Dsts
		for j := range x {
			x[j].NoOut = true
		}
		x[i].Val = n.NodeWrap(a.Val)
		x[i].NoOut = false
	}

	// Make a pool of fork nodes that share input and output channels
	poolSz := len(x)
	PoolFork = MakePool(poolSz, "fork", []Edge{a}, x, forkRdy, forkFire)
	return &PoolFork

}
